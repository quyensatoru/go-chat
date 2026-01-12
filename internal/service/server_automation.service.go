package service

import (
	"backend/config"
	"backend/internal/model"
	"backend/internal/repository"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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
	InstallK8s(server *model.Server, argoCDPassword string) error
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
func (s *serverAutomationService) InstallK8s(server *model.Server, argoCDPassword string) error {
	client, err := goph.New(server.Username, server.IpAddress, goph.Password(server.Password))
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	//unstage: Install K3s, Helm, and ArgoCD

	fmt.Println("🚀 Starting K8s uninstall...")

	k3sUninstallCmd := `sudo /usr/local/bin/k3s-uninstall.sh`

	client.Run(k3sUninstallCmd)

	time.Sleep(10 * time.Second)

	//  Step 1: Install K3s with disabled traefik and servicelb
	fmt.Println("📦 Installing K3s...")
	k3sInstallCmd := `curl -sfL https://get.k3s.io | sh -s - --disable=traefik --disable=servicelb`
	_, err = client.Run(k3sInstallCmd)
	if err != nil {
		return fmt.Errorf("failed to install Kubenetes: %w", err)
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
	fmt.Println("📦 Installing Nginx Ingress Controller...")
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
			`kubectl -n argocd patch secret argocd-secret -p '{"stringData": {"admin.password": "'$(htpasswd -nbBC 10 "" "%s" | tr -d ':\n' | sed 's/$2y/$2a/')'", "admin.passwordMtime": "'$(date +%%FT%%T%%Z)'"}}'`,
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

	allService := map[string]string{}

	for i := 0; i < len(app.Services); i++ {
		serviceUrlEnv := fmt.Sprintf("$(%s_URL)", strings.ToUpper(app.Name+"_"+app.Services[i].Name)) // format APP_SERVICE_URL environment
		domainService := fmt.Sprintf("https://%s/%s", app.Domain, app.Name+"-"+app.Services[i].Name)
		allService[serviceUrlEnv] = domainService
	}

	for _, svc := range app.Services {
		var stringParse interface{}
		//replace env into raw env with dynamic placeholders based on sub-services
		replacements := map[string]string{
			// MongoDB placeholders
			"$(MONGODB_HOST)": app.Name + "-" + svc.Name + "-mongodb",
			"$(MONGODB_PORT)": "27017",

			// Redis placeholders
			"$(REDIS_HOST)": app.Name + "-" + svc.Name + "-redis-master",
			"$(REDIS_PORT)": "6379",

			// RabbitMQ placeholders
			"$(RABBITMQ_HOST)":     app.Name + "-" + svc.Name + "-rabbitmq",
			"$(RABBITMQ_PORT)":     "5672",
			"$(RABBITMQ_USER)":     "dev",
			"$(RABBITMQ_PASSWORD)": "Dev1234567",

			// PostgreSQL placeholders
			"$(POSTGRESQL_HOST)":   app.Name + "-" + svc.Name + "-postgresql",
			"$(POSTGRESQL_PORT)":   "5432",
			"$(POSTGRES_USER)":     "postgres",
			"$(POSTGRES_PASSWORD)": "postgres",
			"$(POSTGRES_DB)":       "postgres",

			// MySQL placeholders
			"$(MYSQL_HOST)":     app.Name + "-" + svc.Name + "-mysql",
			"$(MYSQL_PORT)":     "3306",
			"$(MYSQL_USER)":     "root",
			"$(MYSQL_PASSWORD)": "root",
			"$(MYSQL_DATABASE)": "mydb",

			// Elasticsearch placeholders
			"$(ELASTICSEARCH_HOST)": app.Name + "-" + svc.Name + "-elasticsearch",
			"$(ELASTICSEARCH_PORT)": "9200",
		}

		//replace service domain
		for k, v := range allService {
			replacements[k] = v
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

		client := &http.Client{}
		//Delete variable if exists
		req, err := http.NewRequest(http.MethodDelete, gitlabApiUrl+"/projects/"+gitProjectID+"/variables/"+variableName, nil)
		if err != nil {
			return fmt.Errorf("failed to create delete variable request: %w", err)
		}
		req.Header.Set("PRIVATE-TOKEN", gitlabPrivateToken)
		req.Header.Set("Content-Type", "application/json")

		client.Do(req)
		req, err = http.NewRequest(http.MethodPost, gitlabApiUrl+"/projects/"+gitProjectID+"/variables", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create variable gitlab api: %w", err)
		}

		req.Header.Set("PRIVATE-TOKEN", gitlabPrivateToken)
		req.Header.Set("Content-Type", "application/json")
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
	err = autoCreateHelmChart(app)
	if err != nil {
		return err
	}

	//install argocd
	dir, err := os.Getwd()
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

func autoCreateHelmChart(app *model.App) error {
	// build gitops helm chart for application need automation deploy
	dir, err := os.Getwd()
	config := config.LoadEnv()

	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	gitopRepo := filepath.Join(dir, "../gitops-repo")

	err = os.RemoveAll(gitopRepo)

	if err != nil {
		return fmt.Errorf("failed to remove gitops repo dir: %w", err)
	}

	//clone with create new branch for application
	clone, err := git.PlainClone(gitopRepo, false, &git.CloneOptions{
		URL:           config.GitOpsRepo,
		ReferenceName: plumbing.NewBranchReferenceName("blank"),
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

	branchRef := plumbing.NewBranchReferenceName(app.HelmChart)

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
	})

	if err != nil {
		// Branch chưa tồn tại → tạo branch mới từ main
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Create: true,
		})
		if err != nil {
			return err
		}
	}

	os.RemoveAll(filepath.Join(gitopRepo, "apps"))
	os.Mkdir(filepath.Join(gitopRepo, "apps"), 0755)
	templateHelmChart := filepath.Join(dir, "internal", "template", "chart")

	for _, svc := range app.Services {
		//create build chart
		pathChartSvc := filepath.Join(gitopRepo, "apps", strings.ToLower(svc.Name))
		pathTemplate := filepath.Join(pathChartSvc, "templates")
		os.Mkdir(pathChartSvc, 0755)
		err = copyDir(templateHelmChart, pathTemplate)
		if err != nil {
			log.Fatal(err)
		}

		// Generate dynamic dependencies based on SubServices
		dependencies := ""
		if len(svc.SubServices) > 0 {
			dependencies = "\ndependencies:"
			for _, subSvc := range svc.SubServices {
				switch subSvc {
				case "mongodb":
					dependencies += `
  - name: mongodb
    version: 15.x.x
    repository: "https://charts.bitnami.com/bitnami"
    condition: mongodb.enabled`
				case "redis":
					dependencies += `
  - name: redis
    version: 19.x.x
    repository: "https://charts.bitnami.com/bitnami"
    condition: redis.enabled`
				case "rabbitmq":
					dependencies += `
  - name: rabbitmq
    version: 14.x.x
    repository: "https://charts.bitnami.com/bitnami"
    condition: rabbitmq.enabled`
				case "postgresql":
					dependencies += `
  - name: postgresql
    version: 13.x.x
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled`
				case "mysql":
					dependencies += `
  - name: mysql
    version: 9.x.x
    repository: "https://charts.bitnami.com/bitnami"
    condition: mysql.enabled`
				case "elasticsearch":
					dependencies += `
  - name: elasticsearch
    version: 19.x.x
    repository: "https://charts.bitnami.com/bitnami"
    condition: elasticsearch.enabled`
				}
			}
		}

		//create Chart.yaml
		contentChart := fmt.Sprintf(`apiVersion: v2
name: %s
description: A Helm chart for %s
version: 0.1.0
appVersion: "1.0.0"
type: application%s`, app.Name+"-"+svc.Name+"-chart", app.Name+" "+svc.Name, dependencies)
		os.WriteFile(filepath.Join(pathChartSvc, "Chart.yaml"), []byte(contentChart), 0644)

		// Generate sub-chart configurations based on SubServices
		subChartConfigs := ""
		for _, subSvc := range svc.SubServices {
			switch subSvc {
			case "mongodb":
				subChartConfigs += `

mongodb:
  image:
    repository: bitnamilegacy/mongodb
    tag: "7.0.12"
  volumePermissions:
    enabled: true
    image:
      repository: bitnamilegacy/mongodb
      tag: "7.0.12"
  auth:
    enabled: false
  persistence:
    enabled: false
  service:
    port: 27017`
			case "redis":
				subChartConfigs += `

redis:
  image:
    repository: bitnamilegacy/redis
    tag: "8.2.1-debian-12-r0"
  auth:
    enabled: false
  architecture: standalone
  master:
    persistence:
      enabled: false`
			case "rabbitmq":
				subChartConfigs += `

rabbitmq:
  image:
    registry: docker.io
    repository: bitnamilegacy/rabbitmq
    tag: "4.1.3-debian-12-r1"
  auth:
    username: dev
    password: Dev1234567
  persistence:
    enabled: false
  service:
    port: 5672`
			case "postgresql":
				subChartConfigs += `

postgresql:
  image:
    repository: bitnamilegacy/postgresql
    tag: "16.1.0"
  auth:
    enablePostgresUser: true
    postgresPassword: postgres
  primary:
    persistence:
      enabled: false
  service:
    port: 5432`
			case "mysql":
				subChartConfigs += `

mysql:
  image:
    repository: bitnamilegacy/mysql
    tag: "8.0.35"
  auth:
    rootPassword: root
  primary:
    persistence:
      enabled: false
  service:
    port: 3306`
			case "elasticsearch":
				subChartConfigs += `

elasticsearch:
  image:
    repository: bitnamilegacy/elasticsearch
    tag: "8.11.3"
  master:
    replicaCount: 1
  data:
    replicaCount: 1
  coordinating:
    replicaCount: 1
  service:
    port: 9200`
			}
		}

		// Use service-specific image URL and tag
		imageRepo := svc.ImageURL
		imageTag := svc.ImageTag
		if imageTag == "" {
			imageTag = "latest"
		}

		serviceDomain := app.Domain
		if serviceDomain == "" {
			serviceDomain = app.Name + "-" + svc.Name + ".local" // fallback domain
		}

		//Create values.yaml with service-specific config and sub-chart configs
		contentValues := fmt.Sprintf(`image:
  repository: %s
  tag: %s
  pullPolicy: Always

replicaCount: 1

configmap:
  enabled: false
appConfig:

secret:
  enabled: true
appSecret:

service:
  type: NodePort
  port: 80
  targetPort: 3000

ingress:
  enabled: true
  className: nginx
  host: %s
  tls: false
  tlsSecret: ""%s`, imageRepo, imageTag, serviceDomain, subChartConfigs)
		os.WriteFile(filepath.Join(pathChartSvc, "values.yaml"), []byte(contentValues), 0644)

		//Done setup values
		envDir := filepath.Join(gitopRepo, "envs", app.Name)
		fileValue := filepath.Join(envDir, fmt.Sprintf("values-%s.yaml", svc.Name))
		os.MkdirAll(envDir, 0755)
		contentEnv := fmt.Sprintf(
			"keyEnv: %s\npath: %s",
			app.Name+"_"+svc.Name,
			"/"+app.Name+"-"+svc.Name,
		)
		os.WriteFile(fileValue, []byte(contentEnv), 0644)
	}

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
	return nil
}

func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		return copyFile(path, targetPath)
	})
}
