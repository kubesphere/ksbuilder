package api

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	KubeSphereSystem  = "kubesphere-system"
	ConfigMapDataKey  = "chart.tgz"
	KubeSphereManaged = "kubesphere.io/managed"
)

type Extension struct {
	Metadata *Metadata
	// ChartURL valid when the chart source online.
	ChartURL string
	// ChartData valid when the chart source local.
	ChartData []byte
}

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
					KubeSphereManaged:          "true",
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
				Namespace: KubeSphereSystem,
			},
			BinaryData: map[string][]byte{
				ConfigMapDataKey: ext.ChartData,
			},
		}
		extensionVersion.Spec.ChartDataRef = &corev1alpha1.ConfigMapKeyRef{
			Namespace: configmap.Namespace,
			ConfigMapKeySelector: corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configmap.Name,
				},
				Key: ConfigMapDataKey,
			},
		}
		resources = append(resources, extensionVersion, configmap)
	}
	return resources
}
