package service

import (
	"backend/config"
	"backend/internal/model"
	"backend/internal/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/melbahja/goph"
)

type ServerAutomationService interface {
	CheckConnection(server *model.Server) error
	InstallK8s(server *model.Server, branch, argoCDPassword string) error
	DeployArgoCDApp(app *model.App) error
}

type serverAutomationService struct {
	serverRepo repository.ServerRepository
}

func NewServerAutomationService(serverRepo repository.ServerRepository) ServerAutomationService {
	return &serverAutomationService{
		serverRepo: serverRepo,
	}
}

// CheckConnection tests SSH connectivity to the server
func (s *serverAutomationService) CheckConnection(server *model.Server) error {
	client, err := goph.New(server.Username, server.IpAddress, goph.Password(server.Password))
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	// Test with a simple command
	_, err = client.Run("echo 'connection test'")
	if err != nil {
		return fmt.Errorf("failed to run test command: %w", err)
	}

	return nil
}

// InstallK8s installs K3s, Helm, and ArgoCD in one go
func (s *serverAutomationService) InstallK8s(server *model.Server, branch, argoCDPassword string) error {
	client, err := goph.New(server.Username, server.IpAddress, goph.Password(server.Password))
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	//unstage: Install K3s, Helm, and ArgoCD

	fmt.Println("🚀 Starting K8s uninstall...")
	k3sUninstallCmd := `sudo /usr/local/bin/k3s-uninstall.sh`
	_, err = client.Run(k3sUninstallCmd)

	if err != nil {
		fmt.Printf("failed to uinstall K3s: %v", err)
	}

	time.Sleep(10 * time.Second)

	//  Step 1: Install K3s
	fmt.Println("📦 Installing K3s...")
	k3sInstallCmd := `curl -sfL https://get.k3s.io | sh -`
	_, err = client.Run(k3sInstallCmd)
	if err != nil {
		return fmt.Errorf("failed to install K3s: %w", err)
	}

	// Wait for K3s to be ready
	time.Sleep(10 * time.Second)

	// Export kubeconfig
	_, err = client.Run("sudo cat /etc/rancher/k3s/k3s.yaml > /tmp/kubeconfig")
	if err != nil {
		return fmt.Errorf("failed to export kubeconfig: %w", err)
	}

	// Step 2: Install Helm
	fmt.Println("📦 Installing Helm...")
	helmInstallCmd := `curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash`
	_, err = client.Run(helmInstallCmd)
	if err != nil {
		return fmt.Errorf("failed to install Helm: %w", err)
	}

	// Step 3: Install ArgoCD
	fmt.Println("📦 Installing ArgoCD...")

	// Create argocd namespace
	_, _ = client.Run("sudo k3s kubectl create namespace argocd")

	// Install ArgoCD
	argoCDInstallCmd := `sudo k3s kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml`
	_, err = client.Run(argoCDInstallCmd)
	if err != nil {
		return fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	// Wait for ArgoCD to be ready
	time.Sleep(30 * time.Second)

	// Step 4: Configure ArgoCD
	fmt.Println("⚙️ Configuring ArgoCD...")

	// Change ArgoCD admin password if provided
	if argoCDPassword != "" {
		// Get ArgoCD server pod
		getPodCmd := `sudo k3s kubectl get pods -n argocd -l app.kubernetes.io/name=argocd-server -o jsonpath='{.items[0].metadata.name}'`
		podName, err := client.Run(getPodCmd)
		if err != nil {
			return fmt.Errorf("failed to get ArgoCD server pod: %w", err)
		}

		// Update password
		updatePasswordCmd := fmt.Sprintf(
			`sudo k3s kubectl -n argocd exec %s -- argocd account update-password --current-password $(sudo k3s kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d) --new-password %s`,
			strings.TrimSpace(string(podName)),
			argoCDPassword,
		)
		_, _ = client.Run(updatePasswordCmd) // Ignore errors as password might already be set
	}

	dir, err := os.Getwd()

	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	secretStore, err := os.ReadFile(dir + "/internal/template/external-secret.yaml")

	if err != nil {
		return fmt.Errorf("failed to read external secret template: %w", err)
	}

	// Create external secret operation

	//install external secret operator

	fmt.Println("📦 Installing External Secret Operator...")
	if secretStore != nil {
		createSecretStoreCmd := `helm repo add external-secrets https://charts.external-secrets.io`
		_, err = client.Run(createSecretStoreCmd)
		if err != nil {
			return fmt.Errorf("failed to add external secrets helm repo: %w", err)
		}

		// Install external secrets operator via helm
		installESOHelmCmd := `KUBECONFIG=/tmp/kubeconfig helm install external-secrets external-secrets/external-secrets --namespace external-secrets --create-namespace`

		result, err := client.Run(installESOHelmCmd)
		if err != nil {
			return fmt.Errorf("failed to install external secrets operator via helm: %s", result)
		}
		// Apply external secret manifest
		createAppCmd := fmt.Sprintf(`cat <<EOF | sudo k3s kubectl apply -f -
%s
EOF`, string(secretStore))
		// Wait for external secrets ready
		time.Sleep(30 * time.Second)
		result, err = client.Run(createAppCmd)
		fmt.Printf("apply external secret operation result: %s", string(result))
		if err != nil {
			return fmt.Errorf("failed to create ArgoCD application: %w", err)
		}
	}

	// Expose ArgoCD server (change to NodePort for easy access)
	exposeCmd := `sudo k3s kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "NodePort"}}'`
	_, err = client.Run(exposeCmd)
	if err != nil {
		return fmt.Errorf("failed to expose ArgoCD server: %w", err)
	}

	fmt.Println("✅ K8s installation completed successfully!")
	return nil
}

func (s *serverAutomationService) DeployArgoCDApp(app *model.App) error {
	client, err := goph.New(app.Server.Username, app.Server.IpAddress, goph.Password(app.Server.Password))
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	// create enviroment in gitlab via api
	config := config.LoadEnv()

	gitlabApiUrl := config.GitlabApiUrl
	gitlabPrivateToken := config.GitlabPrivateToken
	gitProjectID := config.GitlabProjectID

	for _, svc := range app.Services {
		var stringParse interface{}
		err := json.Unmarshal([]byte(svc.EnvRaw), &stringParse)
		if err != nil {
			return fmt.Errorf("failed to unmarshal service env: %w", err)
		}

		jsonEnv, err := json.Marshal(stringParse)

		if err != nil {
			return fmt.Errorf("failed to marshal service env: %w", err)
		}

		log.Printf("json env %v", jsonEnv)

		variableName := app.Name + "_" + svc.Name
		variable := struct {
			Key              string `json:"key"`
			Value            string `json:"value"`
			Protected        bool   `json:"protected"`
			EnvironmentScope string `json:"environment_scope"`
		}{
			Key:              variableName,
			Value:            string(jsonEnv),
			Protected:        true,
			EnvironmentScope: "deployment",
		}

		jsonData, err := json.Marshal(variable)
		if err != nil {
			return fmt.Errorf("failed to marshal variable to JSON: %w", err)
		}

		req, err := http.NewRequest(http.MethodPost, gitlabApiUrl+"/projects/"+gitProjectID+"/variables", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create variable gitlab api: %w", err)
		}

		req.Header.Set("PRIVATE-TOKEN", gitlabPrivateToken)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}

		reps, err := client.Do(req)

		if err != nil {
			return fmt.Errorf("failed to call gitlab api to create variable: %w", err)
		}

		defer req.Body.Close()

		body, err := io.ReadAll(reps.Body)

		if err != nil {
			return fmt.Errorf("failed to read gitlab api response body: %w", err)
		}

		log.Println("Gitlab environment response body:", string(body))

	}

	// push change env path on gitops

	dir, err := os.Getwd()

	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	gitopRepo := filepath.Join(dir, "../gitops-repo")

	err = os.RemoveAll(gitopRepo)

	if err != nil {
		return fmt.Errorf("failed to remove gitops repo dir: %w", err)
	}

	//clone with
	clone, err := git.PlainClone(gitopRepo, false, &git.CloneOptions{
		URL:           config.GitOpsRepo,
		ReferenceName: plumbing.NewBranchReferenceName("chart/blocker"),
		SingleBranch:  true,
		Depth:         1,
		Progress:      os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("failed to clone gitops repo: %w", err)
	}

	worktree, err := clone.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	newDir := filepath.Join(gitopRepo, "envs", app.Name)
	fileValueApi := filepath.Join(newDir, "values-api.yaml")
	// fileValueCms := filepath.Join(newDir, "values-cms.yaml")

	os.MkdirAll(newDir, 0755)
	os.WriteFile(fileValueApi, []byte("keyEnv: "+app.Name+"_"+app.Services[0].Name), 0644)
	// os.WriteFile(fileValueCms, []byte("keyEnv: "+app.Name+"_"+app.Services[1].Name), 0644)

	_, err = worktree.Add(".")

	if err != nil {
		return fmt.Errorf("failed to add env path to worktree: %w", err)
	}

	_, err = worktree.Commit("Add env for app "+app.Name, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Adminitrator",
			Email: "quyenpv020803@gmail.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	err = clone.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &githttp.BasicAuth{
			Username: "Administrator",
			Password: config.GitOpsToken,
		},
		Progress: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}
	//install argocd

	templateArgoDir := filepath.Join(dir, "internal", "template", "argocd.yaml")

	argoCDManifet, err := os.ReadFile(templateArgoDir)

	if err != nil {
		return fmt.Errorf("failed to read argocd template: %w", err)
	}

	replacer := strings.NewReplacer(
		"{{ gitopsRepo }}", config.GitOpsRepo,
		"{{ gitopsRevision }}", app.HelmChart,
	)

	manifet := replacer.Replace(string(argoCDManifet))

	// Apply ArgoCD application manifest
	createAppCmd := fmt.Sprintf(`cat <<EOF | sudo k3s kubectl apply -f -
%s	
EOF`, indent(manifet, 0))

	_, err = client.Run(createAppCmd)
	fmt.Printf("create argocd application result: %s", string(createAppCmd))
	if err != nil {
		return fmt.Errorf("failed to create ArgoCD application: %w", err)
	}

	return nil
}

func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
}
