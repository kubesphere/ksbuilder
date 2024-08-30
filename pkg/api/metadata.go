package api

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"os"
	ospath "path"
	"strings"

	"github.com/go-playground/validator/v10"
	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/util/json"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/iso639"
)

const MetadataFilename = "extension.yaml"

type Metadata struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	// The name of the chart. Required.
	Name                 string                                               `json:"name" validate:"required"`
	Version              string                                               `json:"version" validate:"required"`
	DisplayName          corev1alpha1.Locales                                 `json:"displayName" validate:"required"`
	Description          corev1alpha1.Locales                                 `json:"description" validate:"required"`
	Category             string                                               `json:"category" validate:"required"`
	Keywords             []string                                             `json:"keywords,omitempty"`
	Home                 string                                               `json:"home,omitempty"`
	Docs                 string                                               `json:"docs,omitempty"`
	Sources              []string                                             `json:"sources,omitempty"`
	KubeVersion          string                                               `json:"kubeVersion,omitempty"`
	KSVersion            string                                               `json:"ksVersion,omitempty"`
	Maintainers          []*chart.Maintainer                                  `json:"maintainers,omitempty"`
	Provider             map[corev1alpha1.LanguageCode]*corev1alpha1.Provider `json:"provider" validate:"required"`
	StaticFileDirectory  string                                               `json:"staticFileDirectory,omitempty"`
	Icon                 string                                               `json:"icon" validate:"required"`
	Screenshots          []string                                             `json:"screenshots,omitempty"`
	Dependencies         []*chart.Dependency                                  `json:"dependencies,omitempty"`
	InstallationMode     corev1alpha1.InstallationMode                        `json:"installationMode,omitempty"`
	Namespace            string                                               `json:"namespace,omitempty"`
	Images               []string                                             `json:"images,omitempty"`
	ExternalDependencies []corev1alpha1.ExternalDependency                    `json:"externalDependencies,omitempty"`
	Annotations          map[string]string                                    `json:"annotations,omitempty"`
}

type Options struct {
	encodeIcon bool
}

func WithEncodeIcon(encodeIcon bool) func(opts *Options) {
	return func(opts *Options) {
		opts.encodeIcon = encodeIcon
	}
}

func LoadMetadata(path string, options ...func(*Options)) (*Metadata, error) {
	opts := &Options{
		encodeIcon: true,
	}
	for _, f := range options {
		f(opts)
	}

	content, err := os.ReadFile(ospath.Join(path, MetadataFilename))
	if err != nil {
		return nil, err
	}
	metadata, err := ParseMetadata(content)
	if err != nil {
		return nil, err
	}

	if IsLocalFile(metadata.Icon) && opts.encodeIcon {
		base64EncodedIcon, err := encodeIcon(ospath.Join(path, metadata.Icon))
		if err != nil {
			return nil, err
		}
		metadata.Icon = base64EncodedIcon
	}

	if err = metadata.Validate(); err != nil {
		return nil, err
	}
	return metadata, nil
}

func ParseMetadata(data []byte) (*Metadata, error) {
	metadata := new(Metadata)
	if err := yaml.Unmarshal(data, metadata); err != nil {
		return nil, err
	}

	// set default value for necessary fields
	if metadata.InstallationMode == "" {
		metadata.InstallationMode = corev1alpha1.InstallationModeHostOnly
	}
	return metadata, nil
}

func validateLanguageCode(code corev1alpha1.LanguageCode) error {
	if !iso639.IsValidLanguageCode(code) {
		return fmt.Errorf("invalid language code %s, see https://www.loc.gov/standards/iso639-2/php/code_list.php for more details", code)
	}
	return nil
}

func (md *Metadata) validateLanguageCode() error {
	for code := range md.DisplayName {
		if err := validateLanguageCode(code); err != nil {
			return err
		}
	}
	for code := range md.Description {
		if err := validateLanguageCode(code); err != nil {
			return err
		}
	}
	for code := range md.Provider {
		if err := validateLanguageCode(code); err != nil {
			return err
		}
	}
	return nil
}

func (md *Metadata) Validate() error {
	if err := validator.New().Struct(md); err != nil {
		return err
	}
	return md.validateLanguageCode()
}

func (md *Metadata) ToChartYaml() *chart.Metadata {
	var c = chart.Metadata{
		APIVersion:   chart.APIVersionV2,
		Name:         md.Name,
		Version:      md.Version,
		Keywords:     md.Keywords,
		Sources:      md.Sources,
		KubeVersion:  md.KubeVersion,
		Home:         md.Home,
		Dependencies: md.Dependencies,
		Description:  string(md.Description[corev1alpha1.DefaultLanguageCode]),
		Icon:         md.Icon,
		Maintainers:  md.Maintainers,
		Annotations:  md.Annotations,
	}
	return &c
}

func DeepCopy(md *chart.Metadata) *chart.Metadata {
	data, _ := json.Marshal(md)
	out := &chart.Metadata{}
	_ = json.Unmarshal(data, out)
	return out
}

func IsLocalFile(path string) bool {
	if strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "data:image") {
		return false
	}
	return true
}

func encodeIcon(iconPath string) (string, error) {
	content, err := os.ReadFile(iconPath)
	if err != nil {
		return "", err
	}
	var base64Encoding string

	mimeType := mime.TypeByExtension(ospath.Ext(iconPath))
	if mimeType == "" {
		mimeType = http.DetectContentType(content)
	}

	base64Encoding += "data:" + mimeType + ";base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(content)
	return base64Encoding, nil
}
