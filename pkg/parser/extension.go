package parser

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"sigs.k8s.io/yaml"

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
}

func ParseExtension(name string, zipFile []byte) (*Extension, error) {
	data, err := utils.Unzip(zipFile)
	if err != nil {
		return nil, err
	}

	metadata, err := parseMetadata(name, data)
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
	}, nil
}

func parseMetadata(name string, data map[string][]byte) (*extension.Metadata, error) {
	metadata := new(extension.Metadata)
	if err := yaml.Unmarshal(data[path.Join(name, extension.MetadataFilename)], metadata); err != nil {
		return nil, err
	}
	if err := metadata.Validate(); err != nil {
		return nil, fmt.Errorf("validate the extension metadata failed: %s", err.Error())
	}
	if strings.HasPrefix(metadata.Icon, "http://") ||
		strings.HasPrefix(metadata.Icon, "https://") ||
		strings.HasPrefix(metadata.Icon, "data:image") {
		return metadata, nil
	}

	iconData := data[path.Join(name, metadata.Icon)]

	var base64Encoding string
	mimeType := mime.TypeByExtension(path.Ext(metadata.Icon))
	if mimeType == "" {
		mimeType = http.DetectContentType(iconData)
	}

	base64Encoding += "data:" + mimeType + ";base64,"
	base64Encoding += base64.StdEncoding.EncodeToString(iconData)
	metadata.Icon = base64Encoding
	return metadata, nil
}
