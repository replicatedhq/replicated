package initkotsapp

import (
	"fmt"
	kotskinds "github.com/replicatedhq/kots/kotskinds/apis/kots/v1beta1"
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta1"
	"github.com/replicatedhq/yaml/v3"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"strings"
)

func HelmChartFile(chartYaml ChartYaml, kotsManifestsPath string) error {
	kotsHelmCrd := kotskinds.HelmChart{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HelmChart",
			APIVersion: "kots.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: chartYaml.Name,
		},
		Spec: kotskinds.HelmChartSpec{
			Chart: kotskinds.ChartIdentifier{
				ChartVersion: chartYaml.Version,
				Name:         chartYaml.Name,
			},
			Values: map[string]kotskinds.MappedChartValue{
				"foo": {},
				"bar": {},
				"baz": {},
			},
		},
	}

	helmChartFileName := fmt.Sprintf("%s.yaml", chartYaml.Name)
	helmChartFilePath := filepath.Join(kotsManifestsPath, helmChartFileName)

	err := WriteKotsYAML(kotsHelmCrd, helmChartFilePath)
	if err != nil {
		return err
	}

	return nil
}

func MakeFile(kotsBasePath string) error {
	makeFileContents := `
SHELL := /bin/bash -o pipefail

app_slug := "${REPLICATED_APP}"

# Generate channel and release notes. We need to do this differently for github actions vs. command line because of how git works differently in GH actions. 
ifeq ($(origin GITHUB_ACTIONS), undefined)
release_notes := "CLI release of $(shell git symbolic-ref HEAD) triggered by ${shell git config --global user.name}: $(shell basename $$(git remote get-url origin) .git) [SHA: $(shell git rev-parse HEAD)]"
channel := $(shell git rev-parse --abbrev-ref HEAD)
else
release_notes := "GitHub Action release of ${GITHUB_REF} triggered by ${GITHUB_ACTOR}: [$(shell echo $${GITHUB_SHA::7})](https://github.com/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA})"
channel := ${GITHUB_BRANCH_NAME}
endif

# If we're on the master channel, translate that to the "Unstable" channel
ifeq ($(channel), master)
channel := Unstable
endif

# version based on branch/channel
version := $(channel)-$(shell git rev-parse HEAD | head -c7)$(shell git diff --no-ext-diff --quiet --exit-code || echo "-dirty")

.PHONY: deps-vendor-cli
deps-vendor-cli: upstream_version = $(shell  curl --silent --location --fail --output /dev/null --write-out %{url_effective} https://github.com/replicatedhq/replicated/releases/latest | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+$$')
deps-vendor-cli: dist = $(shell uname | tr '[:upper:]' '[:lower:]')
deps-vendor-cli: cli_version = ""
deps-vendor-cli: cli_version = $(shell [[ -x deps/replicated ]] && deps/replicated version | grep version | head -n1 | cut -d: -f2 | tr -d , | tr -d '"' | tr -d " " )

deps-vendor-cli:
	: CLI Local Version $(cli_version)
	: CLI Upstream Version $(upstream_version)
	@if [[ "$(cli_version)" == "$(upstream_version)" ]]; then \
	   echo "Latest CLI version $(upstream_version) already present"; \
	 else \
	   echo '-> Downloading Replicated CLI to ./deps '; \
	   mkdir -p deps/; \
	   curl -s https://api.github.com/repos/replicatedhq/replicated/releases/latest \
	   | grep "browser_download_url.*$(dist)_amd64.tar.gz" \
	   | cut -d : -f 2,3 \
	   | tr -d \" \
	   | wget -O- -qi - \
	   | tar xvz -C deps; \
	 fi


.PHONY: lint
lint: check-api-token check-app deps-vendor-cli
	deps/replicated release lint --app $(app_slug) --yaml-dir manifests

.PHONY: check-api-token
check-api-token:
	@if [ -z "${REPLICATED_API_TOKEN}" ]; then echo "Missing REPLICATED_API_TOKEN"; exit 1; fi

.PHONY: check-app
check-app:
	@if [ -z "$(app_slug)" ]; then echo "Missing REPLICATED_APP"; exit 1; fi

.PHONY: list-releases
list-releases: check-api-token check-app deps-vendor-cli
	deps/replicated release ls --app $(app_slug)

.PHONY: helm-package
helm-package: deps-vendor-cli
	helm package ../. -d manifests/

.PHONY: release
release: check-api-token check-app deps-vendor-cli helm-package lint
	deps/replicated release create \
		--app $(app_slug) \
		--yaml-dir manifests \
		--promote $(channel) \
		--version $(version) \
		--release-notes $(release_notes) \
		--ensure-channel

.PHONY: release-kurl-installer
release-kurl-installer: check-api-token check-app deps-vendor-cli
	deps/replicated installer create \
		--app $(app_slug) \
		--yaml-file kurl-installer.yaml \
		--promote $(channel) \
		--ensure-channel
`
	makeFilePath := filepath.Join(kotsBasePath, "Makefile")

	err := ioutil.WriteFile(makeFilePath, []byte(makeFileContents), 0644)
	if err != nil {
		return err
	}

	return nil
}

func WriteKotsYAML(kotsCrds interface{}, filePath string) error {

	bytes, err := yaml.Marshal(kotsCrds)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

type ChartYaml struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Icon    string `yaml:"icon"`
}

func PreflightFile(chartYaml ChartYaml, kotsManifestsPath string) error {

	kotsPreflightCRD := v1beta1.Preflight{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Preflight",
			APIVersion: "troubleshoot.replicated.com/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: chartYaml.Name,
		},
		Spec: v1beta1.PreflightSpec{
			Analyzers: []*v1beta1.Analyze{
				{
					ClusterVersion: &v1beta1.ClusterVersion{
						AnalyzeMeta: v1beta1.AnalyzeMeta{
							CheckName: "Kubernetes Version",
						},
						Outcomes: []*v1beta1.Outcome{
							{
								Fail: &v1beta1.SingleOutcome{
									When:    "< 1.15.0",
									Message: "This app requires at least Kubernetes 1.15.0",
									URI:     "https://www.kubernetes.io",
								},
							},
							{
								Pass: &v1beta1.SingleOutcome{
									When:    ">= 1.15.0",
									Message: "This app has at least Kubernetes 1.15.0",
									URI:     "https://www.kubernetes.io",
								},
							},
						},
					},
				},

				{
					NodeResources: &v1beta1.NodeResources{
						AnalyzeMeta: v1beta1.AnalyzeMeta{
							CheckName: "Total CPU Capacity",
						},
						Outcomes: []*v1beta1.Outcome{
							{
								Fail: &v1beta1.SingleOutcome{
									When:    "sum(cpuCapacity) < 4",
									Message: "This app requires a cluster with at least 4 cores.",
									URI:     "https://kurl.sh/docs/install-with-kurl/system-requirements",
								},
							},
							{
								Pass: &v1beta1.SingleOutcome{
									Message: "This cluster has at least 4 cores.",
								},
							},
						},
					},
				},
			},
		},
	}

	preflightFilePath := filepath.Join(kotsManifestsPath, "preflight.yaml")

	err := WriteKotsYAML(kotsPreflightCRD, preflightFilePath)
	if err != nil {
		return err
	}

	return nil
}

func ConfigFile(chartYaml ChartYaml, kotsManifestsPath string) error {
	kotsConfigCrd := kotskinds.Config{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Config",
			APIVersion: "kots.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: chartYaml.Name,
		},
		Spec: kotskinds.ConfigSpec{
			Groups: []kotskinds.ConfigGroup{
				{
					Name:        "Config",
					Title:       "Config Options",
					Description: "A default example of how to collect configuration from an end user. This can be mapped to helm values",
					Items: []kotskinds.ConfigItem{
						{
							Name:     "username",
							Title:    "Username",
							Type:     "text",
							HelpText: "Enter the default admin username",
						},
					},
				},
			},
		},
	}

	configFilePath := filepath.Join(kotsManifestsPath, "config.yaml")

	err := WriteKotsYAML(kotsConfigCrd, configFilePath)
	if err != nil {
		return err
	}

	return nil
}

func SupportBundleFile(kotsManifestsPath string) error {

	kotsCollectorCRD := v1beta1.Collector{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Collector",
			APIVersion: "troubleshoot.replicated.com/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "collector",
		},
		Spec: v1beta1.CollectorSpec{
			Collectors: []*v1beta1.Collect{
				{
					ClusterInfo: &v1beta1.ClusterInfo{},
				},
				{
					ClusterResources: &v1beta1.ClusterResources{},
				},
			},
		},
	}

	supportBundleFilePath := filepath.Join(kotsManifestsPath, "support-bundle.yaml")

	err := WriteKotsYAML(kotsCollectorCRD, supportBundleFilePath)
	if err != nil {
		return err
	}

	return nil
}

func AppFile(chartYaml ChartYaml, appName string, kotsManifestsPath string) error {

	kotsAppCrd := kotskinds.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "kots.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: kotskinds.ApplicationSpec{
			Title: appName,
			Icon:  chartYaml.Icon,
		},
	}

	appFilePath := filepath.Join(kotsManifestsPath, "replicated-app.yaml")

	err := WriteKotsYAML(kotsAppCrd, appFilePath)
	if err != nil {
		return err
	}

	return nil
}

func Helmignore(baseDirectory string) error {
	helmignoreContents := `
kots/`
	helmignorePath := filepath.Join(baseDirectory, ".helmignore")

	f, err := os.OpenFile(helmignorePath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if !strings.Contains(string(bytes), "kots/") {
		if _, err = f.WriteString(helmignoreContents); err != nil {
			return err
		}
	}

	return nil
}

func Gitignore(kotsBasePath string) error {
	gitignoreContents := `
deps/
manifests/*.tgz
`
	gitignorePath := filepath.Join(kotsBasePath, ".gitignore")

	err := ioutil.WriteFile(gitignorePath, []byte(gitignoreContents), 0644)
	if err != nil {
		return err
	}

	return nil
}
