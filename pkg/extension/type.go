package extension

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/go-playground/validator/v10"
	"helm.sh/helm/v3/pkg/chart"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/iso639"
)

var Categories = []string{
	"ai-machine-learning",
	"computing",
	"database",
	"dev-tools",
	"integration-delivery",
	"observability",
	"networking",
	"security",
	"storage",
	"streaming-messaging",
	"other",
}

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
		Description:  string(md.Description[corev1alpha1.DefaultLanguageCode]),
		Icon:         md.Icon,
		Maintainers:  md.Maintainers,
		Annotations:  md.Annotations,
	}
	return &c, nil
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

	mimeType := mime.TypeByExtension(path.Ext(iconPath))
	if mimeType == "" {
		mimeType = http.DetectContentType(content)
	}

	base64Encoding += "data:" + mimeType + ";base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(content)
	return base64Encoding, nil
}

type Extension struct {
	Metadata *Metadata
	// ChartURL valid when the chart source online.
	ChartURL string
	// ChartData valid when the chart source local.
	ChartData []byte
}

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

const (
	kubeSphereSystem  = "kubesphere-system"
	configMapDataKey  = "chart.tgz"
	kubeSphereManaged = "kubesphere.io/managed"
)

func (ext *Extension) ToKubernetesResources() []runtimeclient.Object {
	var resources = []runtimeclient.Object{
		&corev1alpha1.Extension{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "kubesphere.io/v1alpha1",
				Kind:       "Extension",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: ext.Metadata.Name,
				Labels: map[string]string{
					corev1alpha1.CategoryLabel: ext.Metadata.Category,
					kubeSphereManaged:          "true",
				},
			},
			Spec: corev1alpha1.ExtensionSpec{
				ExtensionInfo: corev1alpha1.ExtensionInfo{
					Description: ext.Metadata.Description,
					DisplayName: ext.Metadata.DisplayName,
					Icon:        ext.Metadata.Icon,
					Provider:    ext.Metadata.Provider,
					Created:     metav1.Now(),
				},
			},
			Status: corev1alpha1.ExtensionStatus{
				RecommendedVersion: ext.Metadata.Version,
			},
		}}
	extensionVersion := &corev1alpha1.ExtensionVersion{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubesphere.io/v1alpha1",
			Kind:       "ExtensionVersion",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", ext.Metadata.Name, ext.Metadata.Version),
			Labels: map[string]string{
				corev1alpha1.ExtensionReferenceLabel: ext.Metadata.Name,
				corev1alpha1.CategoryLabel:           ext.Metadata.Category,
			},
			Annotations: ext.Metadata.Annotations,
		},
		Spec: corev1alpha1.ExtensionVersionSpec{
			InstallationMode: ext.Metadata.InstallationMode,
			ExtensionInfo: corev1alpha1.ExtensionInfo{
				Description: ext.Metadata.Description,
				DisplayName: ext.Metadata.DisplayName,
				Icon:        ext.Metadata.Icon,
				Provider:    ext.Metadata.Provider,
				Created:     metav1.Now(),
			},
			Docs:                 ext.Metadata.Docs,
			Namespace:            ext.Metadata.Namespace,
			Home:                 ext.Metadata.Home,
			Keywords:             ext.Metadata.Keywords,
			KSVersion:            ext.Metadata.KSVersion,
			KubeVersion:          ext.Metadata.KubeVersion,
			Sources:              ext.Metadata.Sources,
			Version:              ext.Metadata.Version,
			Category:             ext.Metadata.Category,
			Screenshots:          ext.Metadata.Screenshots,
			ExternalDependencies: ext.Metadata.ExternalDependencies,
		},
	}
	if ext.ChartURL != "" {
		extensionVersion.Spec.ChartURL = ext.ChartURL
		resources = append(resources, extensionVersion)
	} else {
		configmap := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("extension-%s-%s-chart", ext.Metadata.Name, ext.Metadata.Version),
				Namespace: kubeSphereSystem,
			},
			BinaryData: map[string][]byte{
				configMapDataKey: ext.ChartData,
			},
		}
		extensionVersion.Spec.ChartDataRef = &corev1alpha1.ConfigMapKeyRef{
			Namespace: configmap.Namespace,
			ConfigMapKeySelector: corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configmap.Name,
				},
				Key: configMapDataKey,
			},
		}
		resources = append(resources, extensionVersion, configmap)
	}
	return resources
}
