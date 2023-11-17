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
)

var Categories = []string{
	"ai-machine-learning", "database", "integration-delivery", "monitoring-logging", "networking", "security",
	"storage", "streaming-messaging",
}

type Metadata struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	// The name of the chart. Required.
	Name             string                                               `json:"name" validate:"required"`
	Version          string                                               `json:"version" validate:"required"`
	DisplayName      corev1alpha1.Locales                                 `json:"displayName" validate:"required"`
	Description      corev1alpha1.Locales                                 `json:"description" validate:"required"`
	Category         string                                               `json:"category" validate:"required"`
	Keywords         []string                                             `json:"keywords,omitempty"`
	Home             string                                               `json:"home,omitempty"`
	Sources          []string                                             `json:"sources,omitempty"`
	KubeVersion      string                                               `json:"kubeVersion,omitempty"`
	KSVersion        string                                               `json:"ksVersion,omitempty"`
	Maintainers      []*chart.Maintainer                                  `json:"maintainers,omitempty"`
	Provider         map[corev1alpha1.LanguageCode]*corev1alpha1.Provider `json:"provider" validate:"required"`
	Icon             string                                               `json:"icon" validate:"required"`
	Screenshots      []string                                             `json:"screenshots,omitempty"`
	Dependencies     []*chart.Dependency                                  `json:"dependencies,omitempty"`
	InstallationMode corev1alpha1.InstallationMode                        `json:"installationMode,omitempty"`
}

func (md *Metadata) validateLanguageCode() error {
	m := make(map[corev1alpha1.LanguageCode]int)
	for k := range md.DisplayName {
		m[k]++
	}
	for k := range md.Description {
		m[k]++
	}
	for k := range md.Provider {
		m[k]++
	}

	for k, v := range m {
		if v != 3 {
			return fmt.Errorf("validate language code failed: only some multi-language sections include the %s language.\nIf you want to support a language, make sure all multi-language sections(displayName, description and provider etc.) include that language", k)
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

func (md *Metadata) Init(p string) error {
	if err := md.LoadIcon(p); err != nil {
		return err
	}

	if md.InstallationMode == "" {
		md.InstallationMode = corev1alpha1.InstallationModeHostOnly
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
		Description:  string(md.Description[corev1alpha1.DefaultLanguageCode]),
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
	kubeSphereSystem = "kubesphere-system"
	configMapDataKey = "chart.tgz"
)

func (ext *Extension) ToKubernetesResources() []runtimeclient.Object {
	cmName := fmt.Sprintf("extension-%s-%s-chart", ext.Metadata.Name, ext.Metadata.Version)

	return []runtimeclient.Object{
		&corev1alpha1.Extension{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "kubesphere.io/v1alpha1",
				Kind:       "Extension",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: ext.Metadata.Name,
				Labels: map[string]string{
					corev1alpha1.CategoryLabel: ext.Metadata.Category,
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
		},
		&corev1alpha1.ExtensionVersion{
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
			},
			Spec: corev1alpha1.ExtensionVersionSpec{
				InstallationMode: ext.Metadata.InstallationMode,
				ChartDataRef: &corev1alpha1.ConfigMapKeyRef{
					Namespace: kubeSphereSystem,
					ConfigMapKeySelector: corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cmName,
						},
						Key: configMapDataKey,
					},
				},
				ExtensionInfo: corev1alpha1.ExtensionInfo{
					Description: ext.Metadata.Description,
					DisplayName: ext.Metadata.DisplayName,
					Icon:        ext.Metadata.Icon,
					Provider:    ext.Metadata.Provider,
					Created:     metav1.Now(),
				},
				Home:        ext.Metadata.Home,
				Keywords:    ext.Metadata.Keywords,
				KSVersion:   ext.Metadata.KSVersion,
				KubeVersion: ext.Metadata.KubeVersion,
				Sources:     ext.Metadata.Sources,
				Version:     ext.Metadata.Version,
				Category:    ext.Metadata.Category,
				Screenshots: ext.Metadata.Screenshots,
			},
		},
		&corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: kubeSphereSystem,
			},
			BinaryData: map[string][]byte{
				configMapDataKey: ext.ChartData,
			},
		},
	}
}
