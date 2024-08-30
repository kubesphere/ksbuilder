package extension

import (
	"bytes"
	"fmt"
	"os"
	ospath "path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/otiai10/copy"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/api"
	"github.com/kubesphere/ksbuilder/pkg/utils"
)

var applicationClassTmpl = template.Must(template.New("ApplicationClass").Funcs(sprig.FuncMap()).Parse(`
apiVersion: applicationclass.kubesphere.io/v1alpha1
kind: ApplicationClass
metadata:
  name: {{.Name}}-{{.PackageVersion}}
  labels:
    applicationclass.kubesphere.io/group: {{.ApplicationClassGroup}}
provisioner: {{.Provisioner | quote}}
parameters: {{.Parameters | toJson}}
spec:
  appVersion: {{.AppVersion | quote}}
  packageVersion: {{.PackageVersion | quote}}
  icon: {{.Icon | quote}}
  description: {{.Description | toJson}}
  maintainer: {{.Maintainer | toJson}}
`))

type ApplicationClass struct {
	ApplicationClassGroup string               `json:"applicationClassGroup,omitempty"`
	Name                  string               `json:"name,omitempty"`
	Provisioner           string               `json:"provisioner,omitempty"`
	Parameters            map[string]string    `json:"parameters,omitempty"`
	AppVersion            string               `json:"appVersion,omitempty"`
	PackageVersion        string               `json:"packageVersion,omitempty"`
	Icon                  string               `json:"icon,omitempty"`
	Description           corev1alpha1.Locales `json:"description,omitempty"`
	Maintainer            *chart.Maintainer    `json:"maintainer,omitempty"`
}

func LoadApplicationClass(name, tempDir string) error {
	var b bytes.Buffer
	defer func() {
		b.Reset()
	}()

	content, err := os.ReadFile(tempDir + "/applicationclass.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var appClass ApplicationClass
	if err = yaml.Unmarshal(content, &appClass); err != nil {
		return err
	}

	// Validate
	if len(appClass.Name) == 0 {
		return nil
	}

	filePath := ospath.Join(tempDir, "charts/applicationclass")
	if err = os.MkdirAll(filePath, 0644); err != nil {
		return err
	}

	c := &chart.Metadata{
		APIVersion: chart.APIVersionV2,
		Name:       appClass.Name,
		Version:    appClass.PackageVersion,
		AppVersion: appClass.AppVersion,
	}
	appClassChart, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err = os.WriteFile(filePath+"/Chart.yaml", appClassChart, 0644); err != nil {
		return err
	}
	if err = os.MkdirAll(filePath+"/templates", 0644); err != nil {
		return err
	}

	if appClass.Provisioner == "kubesphere.io/helm-application" {
		var cmName = fmt.Sprintf("application-%s-chart", name)
		appClass.Parameters = make(map[string]string)
		appClass.Parameters["configmap"] = cmName
		appClass.Parameters["namespace"] = "kubesphere-system"

		if err = copy.Copy(tempDir+"/application-package.yaml", filePath+"/templates/application-package.yaml"); err != nil {
			return err
		}
	}

	if err = applicationClassTmpl.Execute(&b, appClass); err != nil {
		return err
	}
	return os.WriteFile(filePath+"/templates/applicationclass.yaml", b.Bytes(), 0644)
}

func removeOutDir(path string) string {
	// aaa/xxx.md -> xxx.md
	// aaa/bbb/xxx.md -> bbb/xxx.md
	elements := strings.Split(path, "/")
	if len(elements) == 1 {
		return path
	}
	return filepath.Join(elements[1:]...)
}

func WriteFilesToTempDir(path, tempDir string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return copy.Copy(path, tempDir)
	}

	// the path is a zip file
	zipFile, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	data, err := utils.Unzip(zipFile)
	if err != nil {
		return err
	}
	for name, content := range data {
		p := ospath.Join(tempDir, removeOutDir(name))
		dir, _ := filepath.Split(p)
		if err = os.MkdirAll(dir, 0700); err != nil && !os.IsExist(err) {
			return err
		}
		if err = os.WriteFile(p, content, 0644); err != nil {
			return err
		}
	}
	return nil
}

func Load(path string) (*api.Extension, error) {
	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir) // nolint

	if err = WriteFilesToTempDir(path, tempDir); err != nil {
		return nil, err
	}

	metadata, err := api.LoadMetadata(tempDir)
	if err != nil {
		return nil, err
	}
	var extension api.Extension
	extension.Metadata = metadata

	chartMetadata, err := yaml.Marshal(metadata.ToChartYaml())
	if err != nil {
		return nil, err
	}

	if err = os.WriteFile(tempDir+"/Chart.yaml", chartMetadata, 0644); err != nil {
		return nil, err
	}

	if err = LoadApplicationClass(metadata.Name, tempDir); err != nil {
		return nil, err
	}

	ch, err := loader.LoadDir(tempDir)
	if err != nil {
		return nil, err
	}
	chartFilename, err := chartutil.Save(ch, tempDir)
	if err != nil {
		return nil, err
	}
	chartContent, err := os.ReadFile(chartFilename)
	if err != nil {
		return nil, err
	}

	extension.ChartData = chartContent
	return &extension, nil
}

func LoadFromHelm(path string) (*api.Extension, error) {
	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir) // nolint

	helmClient := action.NewPullWithOpts(action.WithConfig(new(action.Configuration)))
	helmClient.DestDir = tempDir
	helmClient.Settings = cli.New()
	registryClient, err := registry.NewClient()
	if err != nil {
		return nil, err
	}
	helmClient.SetRegistryClient(registryClient)

	var version string
	plist := strings.Split(path, ":")
	if len(plist) > 2 {
		version = plist[len(plist)-1]
		path = strings.Join(plist[:len(plist)-1], ":")
	}
	if version == "" {
		tags, err := registryClient.Tags(strings.TrimPrefix(path, fmt.Sprintf("%s://", registry.OCIScheme)))
		if err != nil {
			return nil, err
		}
		// set latest tag to version
		version = tags[0]
	}
	helmClient.Version = version

	TarName := fmt.Sprintf("%s-%s.tgz", filepath.Base(path), version)

	// pull file
	if _, err := helmClient.Run(path); err != nil {
		return nil, err
	}
	// expandFile to tempDir
	if err := chartutil.ExpandFile(tempDir, filepath.Join(tempDir, TarName)); err != nil {
		return nil, err
	}

	var extension api.Extension
	metadata, err := api.LoadMetadata(filepath.Join(tempDir, filepath.Base(path)))
	if err != nil {
		return nil, err
	}
	extension.Metadata = metadata
	extension.ChartURL = path + ":" + version

	return &extension, nil
}
