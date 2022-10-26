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
	"helm.sh/helm/v3/pkg/chart"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const LanguageCodeEn = "en"
const LanguageCodeZhCn = "zh-cn"

type LanguageCode string
type LocaleString string
type Locales map[LanguageCode]LocaleString

func (l Locales) Default() string {
	if en, ok := l[LanguageCodeEn]; ok {
		return string(en)
	}
	if zhCn, ok := l[LanguageCodeZhCn]; ok {
		return string(zhCn)
	}
	// pick up the first value
	for _, ls := range l {
		return string(ls)
	}
	return ""
}

type Metadata struct {
	// The name of the chart. Required.
	Name         string              `json:"name,omitempty"`
	DisplayName  Locales             `json:"displayName,omitempty"`
	Description  Locales             `json:"description,omitempty"`
	ApiVersion   string              `json:"apiVersion,omitempty"`
	Icon         string              `json:"icon,omitempty"`
	Version      string              `json:"version,omitempty"`
	Keywords     []string            `json:"keywords,omitempty"`
	Sources      []string            `json:"sources,omitempty"`
	KubeVersion  string              `json:"kubeVersion,omitempty"`
	KsVersion    string              `json:"ksVersion,omitempty"`
	Home         string              `json:"home,omitempty"`
	Dependencies []*chart.Dependency `json:"dependencies,omitempty"`
}

func (md *Metadata) Validate(p string) error {
	// TODO: validate metadata
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
		Name:         md.Name,
		APIVersion:   "v1",
		Version:      md.Version,
		Keywords:     md.Keywords,
		Sources:      md.Sources,
		KubeVersion:  md.KubeVersion,
		Home:         md.Home,
		Dependencies: md.Dependencies,
		Description:  md.Description.Default(),
		Icon:         md.Icon,
	}

	return &c, nil
}

type Extension struct {
	Metadata  *Metadata
	ChartData []byte
}

var (
	extensionTmpl        = template.New("Extension").Funcs(sprig.FuncMap())
	extensionVersionTmpl = template.New("ExtensionVersion").Funcs(sprig.FuncMap())
)

func init() {
	var err error
	extensionTmpl, err = extensionTmpl.Parse(`
apiVersion: kubesphere.io/v1alpha1
kind: Extension
metadata:
  name: {{.Name}}
spec:
  description: {{.Description | toJson}}
  displayName: {{.DisplayName | toJson}}
`)
	if err != nil {
		panic(err)
	}
	extensionVersionTmpl, err = extensionVersionTmpl.Parse(`
apiVersion: kubesphere.io/v1alpha1
kind: ExtensionVersion
metadata:
  name: {{.Name}}-{{.Version}}
spec:
  chartDataRef: 
    namespace: kubesphere-system
    name: extension-{{.Name}}-{{.Version}}-chart
    key: chart.tgz
  description: {{.Description | toJson}}
  displayName: {{.DisplayName | toJson}}
  home: {{.Home | quote}}
  icon: {{.Icon | quote}}
  keywords: {{.Keywords | toJson}}
  ksVersion: {{.KsVersion | quote}}
  kubeVersion: {{.KubeVersion | quote}}
  sources: {{.Sources | toJson}}
  version: {{.Version | quote}}
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
		Data: map[string]string{
			"chart.tgz": base64.StdEncoding.EncodeToString(ext.ChartData),
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
