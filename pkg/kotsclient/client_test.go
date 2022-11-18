package kotsclient

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
)

var (
	pact              dsl.Pact
	testInstallerYAML = `apiVersion: kurl.sh/v1beta1
kind: Installer
metadata:
  name: 'myapp'
help_text: |
  Please check this file exists in root directory: config.yaml
spec:
  kubernetes:
    version: latest
  docker:
    version: latest
  weave:
    version: latest
  rook:
    version: latest
  contour:
    version: latest
  registry:
    version: latest
  prometheus:
    version: latest
  kotsadm:
    version: latest
`
	testMultiYAML = "[{\"name\":\"example-deployment.yaml\",\"path\":\"example-deployment.yaml\",\"content\":\"---\\napiVersion: apps/v1\\nkind: Deployment\\nmetadata:\\n  name: nginx\\n  labels:\\n    app: nginx\\nspec:\\n  selector:\\n    matchLabels:\\n      app: nginx\\n  template:\\n    metadata:\\n      labels:\\n        app: nginx\\n      annotations:\\n        backup.velero.io/backup-volumes: nginx-content\\n    spec:\\n      containers:\\n      - name: nginx\\n        image: nginx\\n        volumeMounts:\\n        - name: nginx-content\\n          mountPath: /usr/share/nginx/html/\\n        resources:\\n          limits:\\n            memory: '256Mi'\\n            cpu: '500m'\\n          requests:\\n            memory: '32Mi'\\n            cpu: '100m'\\n      volumes:\\n      - name: nginx-content\\n        configMap:\\n          name: nginx-content\\n\",\"children\":[]},{\"name\":\"example-service.yaml\",\"path\":\"example-service.yaml\",\"content\":\"apiVersion: v1\\nkind: Service\\nmetadata:\\n  name: nginx\\n  labels:\\n    app: nginx\\n  annotations:\\n    kots.io/when: '{{repl not IsKurl }}'\\nspec:\\n  type: ClusterIP\\n  ports:\\n    - port: 80\\n  selector:\\n    app: nginx\\n---\\napiVersion: v1\\nkind: Service\\nmetadata:\\n  name: nginx\\n  labels:\\n    app: nginx\\n  annotations:\\n    kots.io/when: '{{repl IsKurl }}'\\nspec:\\n  type: NodePort\\n  ports:\\n    - port: 80\\n      nodePort: 8888\\n  selector:\\n    app: nginx\",\"children\":[]},{\"name\":\"kots-app.yaml\",\"path\":\"kots-app.yaml\",\"content\":\"---\\napiVersion: kots.io/v1beta1\\nkind: Application\\nmetadata:\\n  name: nginx\\nspec:\\n  title: App Name\\n  icon: https://raw.githubusercontent.com/cncf/artwork/master/projects/kubernetes/icon/color/kubernetes-icon-color.png\\n  statusInformers:\\n    - deployment/nginx\\n  ports:\\n    - serviceName: \\\"nginx\\\"\\n      servicePort: 80\\n      localPort: 8888\\n      applicationUrl: \\\"http://nginx\\\"\\n\",\"children\":[]},{\"name\":\"kots-config.yaml\",\"path\":\"kots-config.yaml\",\"content\":\"---\\napiVersion: kots.io/v1beta1\\nkind: Config\\nmetadata:\\n  name: config\\nspec:\\n  groups: []\",\"children\":[]},{\"name\":\"kots-preflight.yaml\",\"path\":\"kots-preflight.yaml\",\"content\":\"apiVersion: troubleshoot.sh/v1beta2\\nkind: Preflight\\nmetadata:\\n  name: preflight-checks\\nspec:\\n  analyzers: []\",\"children\":[]},{\"name\":\"kots-support-bundle.yaml\",\"path\":\"kots-support-bundle.yaml\",\"content\":\"apiVersion: troubleshoot.sh/v1beta2\\nkind: SupportBundle\\nmetadata:\\n  name: support-bundle\\nspec:\\n  collectors:\\n    - clusterInfo: {}\\n    - clusterResources: {}\\n    - logs:\\n        selector:\\n          - app=nginx\\n        namespace: '{{repl Namespace }}'\\n\",\"children\":[]}]"
	ksuidRegex    = "^[a-zA-Z0-9]{27}$"
)

func TestMain(m *testing.M) {
	if os.Getenv("SKIP_PACT_TESTING") != "" {
		return
	}

	pact = createPact()

	pact.Setup(true)

	code := m.Run()

	if err := pact.WritePact(); err != nil {
		log.Fatalf("Error writing pact file: %v", err)
	}
	pact.Teardown()

	os.Exit(code)
}

func createPact() dsl.Pact {
	dir, _ := os.Getwd()

	pactDir := path.Join(dir, "..", "..", "pacts")
	logDir := path.Join(dir, "..", "..", "logs")

	return dsl.Pact{
		Consumer: "replicated-cli",
		Provider: "vendor-api",
		LogDir:   logDir,
		PactDir:  pactDir,
		LogLevel: "debug",
	}
}
