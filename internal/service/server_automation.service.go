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
	k3sInstallCmd := `curl -sfL https://get.k3s.io | sh -s - --disable=traefik --disable=servicelb`
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

	// pre install ingresss controller - stop nginx local
	cmdStopNginx := `systemctl stop nginx`
	_, err = client.Run(cmdStopNginx)

	if err != nil {
		return fmt.Errorf("failed to stop nginx")
	}

	// install nginx ingress controller
	helmAddRepoCmd := `helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx`
	_, err = client.Run(helmAddRepoCmd)
	if err != nil {
		return fmt.Errorf("failed to add ingress-nginx helm repo: %w", err)
	}
	helmUpdateCmd := `helm repo update`
	_, err = client.Run(helmUpdateCmd)
	if err != nil {
		return fmt.Errorf("failed to update helm repos: %w", err)
	}
	helmInstallNginxCmd := `KUBECONFIG=/tmp/kubeconfig helm install nginx-ingress ingress-nginx/ingress-nginx --namespace ingress-nginx --create-namespace --set controller.hostNetwork=true --set controller.kind=DaemonSet --set controller.service.type=ClusterIP --set controller.dnsPolicy=ClusterFirstWithHostNet`
	_, err = client.Run(helmInstallNginxCmd)
	if err != nil {
		return fmt.Errorf("failed to install nginx ingress controller via helm: %w", err)
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
		_, err := client.Run(getPodCmd)
		if err != nil {
			return fmt.Errorf("failed to get ArgoCD server pod: %w", err)
		}

		// Update password
		updatePasswordCmd := fmt.Sprintf(
			`kubectl -n argocd patch secret argocd-secret -p '{"stringData": {"admin.password": "'$(htpasswd -nbBC 10 "" "%s" | tr -d ':\n' | sed 's/$2y/$2a/')'", "admin.passwordMtime": "'$(date +%FT%T%Z)'"}}'`,
			argoCDPassword,
		)
		_, _ = client.Run(updatePasswordCmd) // Ignore errors as password might already be set
	}

	// apply ingress for argocd server
	ingressArgoCmd := `cat <<EOF | sudo k3s kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: argocd-server-ingress
  namespace: argocd
  annotations:
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"
spec:
  ingressClassName: nginx
  rules:
  - host: devops.mida-app.io
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: argocd-server
            port:
              number: 80
EOF`

	_, err = client.Run(ingressArgoCmd)
	if err != nil {
		return fmt.Errorf("failed to create ArgoCD ingress: %w", err)
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

	_, err = client.Run(`kubectl patch deployment argocd-server -n argocd --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--insecure"}]'`)
	if err != nil {
		return fmt.Errorf("failed to disable insecure mode: %w", err)
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
		//replace env into raw env
		replacements := map[string]string{
			"$(MONGODB_HOST)":      app.Name + "-" + svc.Name + "-mongodb",
			"$(REDIS_HOST)":        app.Name + "-" + svc.Name + "-redis-master",
			"$(RABBITMQ_HOST)":     app.Name + "-" + svc.Name + "-rabbitmq",
			"$(MONGODB_PORT)":      "27017",
			"$(REDIS_PORT)":        "6379",
			"$(RABBITMQ_PORT)":     "5672",
			"$(RABBITMQ_USER)":     "guest",
			"$(RABBITMQ_PASSWORD)": "guest",
		}
		rawEnv := svc.EnvRaw

		for oldValue, newValue := range replacements {
			rawEnv = strings.ReplaceAll(rawEnv, oldValue, newValue)
		}
		err := json.Unmarshal([]byte(rawEnv), &stringParse)
		if err != nil {
			return fmt.Errorf("failed to unmarshal service env: %w", err)
		}

		jsonEnv, err := json.Marshal(stringParse)

		if err != nil {
			return fmt.Errorf("failed to marshal service env: %w", err)
		}

		log.Printf("json env %v", string(jsonEnv))

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
	contentEnv := fmt.Sprintf(
		"keyEnv: %s\npath: %s",
		app.Name+"_"+app.Services[0].Name,
		"/"+app.Name+"-"+app.Services[0].Name,
	)
	os.WriteFile(fileValueApi, []byte(contentEnv), 0644)
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

	result, err := client.Run(createAppCmd)
	fmt.Printf("create argocd application result: %s", string(result))
	if err != nil {
		return fmt.Errorf("failed to create ArgoCD application: %w", err)
	}

	return nil
}

func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
}
