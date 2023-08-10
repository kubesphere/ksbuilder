package extension

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-playground/validator/v10"
	"helm.sh/helm/v3/pkg/chart"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	LanguageCodeEn = "en"
	LanguageCodeZh = "zh"
)

type LanguageCode string
type Locales map[LanguageCode]string

func (l Locales) Default() string {
	if en, ok := l[LanguageCodeEn]; ok {
		return en
	}
	if zh, ok := l[LanguageCodeZh]; ok {
		return zh
	}
	// pick up the first value
	for _, ls := range l {
		return ls
	}
	return ""
}

var Categories = []string{
	"ai-machine-learning", "database", "integration-delivery", "monitoring-logging", "networking", "security",
	"storage", "streaming-messaging",
}

type Metadata struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	// The name of the chart. Required.
	Name             string                             `json:"name" validate:"required"`
	Version          string                             `json:"version" validate:"required"`
	DisplayName      Locales                            `json:"displayName" validate:"required"`
	Description      Locales                            `json:"description" validate:"required"`
	Category         string                             `json:"category" validate:"required"`
	Keywords         []string                           `json:"keywords,omitempty"`
	Home             string                             `json:"home,omitempty"`
	Sources          []string                           `json:"sources,omitempty"`
	KubeVersion      string                             `json:"kubeVersion,omitempty"`
	KSVersion        string                             `json:"ksVersion,omitempty"`
	Maintainers      []*chart.Maintainer                `json:"maintainers,omitempty"`
	Provider         map[LanguageCode]*chart.Maintainer `json:"provider" validate:"required"`
	Icon             string                             `json:"icon" validate:"required"`
	Screenshots      []string                           `json:"screenshots,omitempty"`
	Dependencies     []*chart.Dependency                `json:"dependencies,omitempty"`
	InstallationMode string                             `json:"installationMode,omitempty"`
}

func (md *Metadata) Validate() error {
	return validator.New().Struct(md)
}

func (md *Metadata) Init(p string) error {
	err := md.LoadIcon(p)
	if err != nil {
		return err
	}
	return nil
}

func (md *Metadata) loadIcon(p string) (string, error) {
	// If the icon is url or base64, you can use it directly.
	// Otherwise, load the file encoding as base64
	if strings.HasPrefix(md.Icon, "http://") ||
		strings.HasPrefix(md.Icon, "https://") ||
		strings.HasPrefix(md.Icon, "data:image") {
		return md.Icon, nil
	}
	content, err := os.ReadFile(path.Join(p, md.Icon))
	if err != nil {
		return "", err
	}
	var base64Encoding string

	mimeType := mime.TypeByExtension(path.Ext(md.Icon))
	if mimeType == "" {
		mimeType = http.DetectContentType(content)
	}

	base64Encoding += "data:" + mimeType + ";base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(content)
	return base64Encoding, nil
}

func (md *Metadata) LoadIcon(p string) error {
	icon, err := md.loadIcon(p)
	if err != nil {
		return err
	}
	md.Icon = icon
	return nil
}

func (md *Metadata) ToChartYaml() (*chart.Metadata, error) {
	var c = chart.Metadata{
		APIVersion:   chart.APIVersionV2,
		Name:         md.Name,
		Version:      md.Version,
		Keywords:     md.Keywords,
		Sources:      md.Sources,
		KubeVersion:  md.KubeVersion,
		Home:         md.Home,
		Dependencies: md.Dependencies,
		Description:  md.Description.Default(),
		Icon:         md.Icon,
		Maintainers:  md.Maintainers,
	}
	return &c, nil
}

type Extension struct {
	Metadata  *Metadata
	ChartData []byte
}

type ApplicationClass struct {
	ApplicationClassGroup string            `json:"applicationClassGroup,omitempty"`
	Name                  string            `json:"name,omitempty"`
	Provisioner           string            `json:"provisioner,omitempty"`
	Parameters            map[string]string `json:"parameters,omitempty"`
	AppVersion            string            `json:"appVersion,omitempty"`
	PackageVersion        string            `json:"packageVersion,omitempty"`
	Icon                  string            `json:"icon,omitempty"`
	Description           Locales           `json:"description,omitempty"`
	Maintainer            *chart.Maintainer `json:"maintainer,omitempty"`
}

var (
	extensionTmpl        = template.New("Extension").Funcs(sprig.FuncMap())
	extensionVersionTmpl = template.New("ExtensionVersion").Funcs(sprig.FuncMap())
	applicationClassTmpl = template.New("ApplicationClass").Funcs(sprig.FuncMap())
)

func init() {
	var err error
	extensionTmpl, err = extensionTmpl.Parse(`
apiVersion: kubesphere.io/v1alpha1
kind: Extension
metadata:
  name: {{.Name}}
  labels:
    kubesphere.io/category: {{.Category | quote}}
spec:
  description: {{.Description | toJson}}
  displayName: {{.DisplayName | toJson}}
  icon: {{.Icon | quote}}
  provider: {{.Provider | toJson}}
status:
  recommendedVersion: {{.Version}}
`)
	if err != nil {
		panic(err)
	}
	extensionVersionTmpl, err = extensionVersionTmpl.Parse(`
apiVersion: kubesphere.io/v1alpha1
kind: ExtensionVersion
metadata:
  name: {{.Name}}-{{.Version}}
  labels:
    kubesphere.io/extension-ref: {{.Name}}
    kubesphere.io/category: {{.Category | quote}}
spec:
{{- with .InstallationMode }}
  installationMode: {{.}}
{{- end }}
  chartDataRef:
    namespace: kubesphere-system
    name: extension-{{.Name}}-{{.Version}}-chart
    key: chart.tgz
  description: {{.Description | toJson}}
  displayName: {{.DisplayName | toJson}}
  home: {{.Home | quote}}
  icon: {{.Icon | quote}}
  keywords: {{.Keywords | toJson}}
  ksVersion: {{.KSVersion | quote}}
  kubeVersion: {{.KubeVersion | quote}}
  sources: {{.Sources | toJson}}
  version: {{.Version | quote}}
  category: {{.Category | quote}}
  screenshots: {{.Screenshots | toJson}}
`)
	if err != nil {
		panic(err)
	}

	applicationClassTmpl, err = applicationClassTmpl.Parse(`
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
`)
	if err != nil {
		panic(err)
	}
}

func (ext *Extension) ToKubernetesResources() []byte {
	// TODO: use kubesphere.io/api
	var b bytes.Buffer
	defer func() {
		b.Reset()
	}()

	var cmName = fmt.Sprintf("extension-%s-%s-chart", ext.Metadata.Name, ext.Metadata.Version)

	var cm = v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: "kubesphere-system",
		},
		BinaryData: map[string][]byte{
			"chart.tgz": ext.ChartData,
		},
	}
	cmByte, err := yaml.Marshal(cm)
	if err != nil {
		panic(err)
	}

	b.Write(cmByte)
	b.WriteString("\n---\n")
	err = extensionTmpl.Execute(&b, ext.Metadata)
	if err != nil {
		panic(err)
	}
	b.WriteString("\n---\n")
	err = extensionVersionTmpl.Execute(&b, ext.Metadata)
	if err != nil {
		panic(err)
	}

	return b.Bytes()
}
