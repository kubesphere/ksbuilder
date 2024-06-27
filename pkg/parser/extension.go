package parser

import (
	"fmt"
	"path"

	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"

	"github.com/kubesphere/ksbuilder/pkg/extension"
	"github.com/kubesphere/ksbuilder/pkg/utils"
)

type Extension struct {
	ChartMetadata      *chart.Metadata
	DisplayName        corev1alpha1.Locales
	Description        corev1alpha1.Locales
	README             corev1alpha1.Locales
	Changelog          corev1alpha1.Locales
	Category           string
	KSVersion          string
	Screenshots        []string
	Provider           map[corev1alpha1.LanguageCode]*corev1alpha1.Provider
	SupportedLanguages []corev1alpha1.LanguageCode
	Docs               string
}

func ParseExtension(name string, zipFile []byte) (*Extension, error) {
	data, err := utils.Unzip(zipFile)
	if err != nil {
		return nil, err
	}

	metadata, err := extension.ParseMetadata(data[path.Join(name, extension.MetadataFilename)])
	if err != nil {
		return nil, err
	}
	chartMetadata, err := metadata.ToChartYaml()
	if err != nil {
		return nil, err
	}

	displayNameLanguages := sets.KeySet(metadata.DisplayName)
	descriptionLanguages := sets.KeySet(metadata.Description)
	supportedLanguages := displayNameLanguages.Intersection(descriptionLanguages).UnsortedList()

	readmeData := corev1alpha1.Locales{}
	for _, lang := range supportedLanguages {
		readmeFileName := "README.md"
		if lang != corev1alpha1.LanguageCodeEn {
			readmeFileName = fmt.Sprintf("README_%s.md", lang)
		}
		readmeData[lang] = corev1alpha1.LocaleString(data[path.Join(name, readmeFileName)])
	}
	changelogData := corev1alpha1.Locales{}
	for _, lang := range supportedLanguages {
		changelogFileName := "CHANGELOG.md"
		if lang != corev1alpha1.LanguageCodeEn {
			changelogFileName = fmt.Sprintf("CHANGELOG_%s.md", lang)
		}
		changelogData[lang] = corev1alpha1.LocaleString(data[path.Join(name, changelogFileName)])
	}

	return &Extension{
		ChartMetadata:      chartMetadata,
		DisplayName:        metadata.DisplayName,
		Description:        metadata.Description,
		KSVersion:          metadata.KSVersion,
		README:             readmeData,
		Changelog:          changelogData,
		Category:           metadata.Category,
		Screenshots:        metadata.Screenshots,
		Provider:           metadata.Provider,
		SupportedLanguages: supportedLanguages,
		Docs:               metadata.Docs,
	}, nil
}
