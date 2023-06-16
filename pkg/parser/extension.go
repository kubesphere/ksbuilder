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
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

type Extension struct {
	ChartMetadata       *chart.Metadata
	DisplayName         extension.Locales
	Description         extension.Locales
	README              extension.Locales
	Changelog           extension.Locales
	Category            string
	KSVersion           string
	StaticFileDirectory string
	Screenshots         []string
	Provider            map[extension.LanguageCode]*chart.Maintainer
	SupportedLanguages  []extension.LanguageCode
}

func ParseExtension(name string, zipFile []byte) (*Extension, error) {
	data, err := unzip(zipFile)
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

	readmeData := extension.Locales{}
	for _, lang := range supportedLanguages {
		readmeFileName := "README.md"
		if lang != extension.LanguageCodeEn {
			readmeFileName = fmt.Sprintf("README_%s.md", lang)
		}
		readmeData[lang] = string(data[path.Join(name, readmeFileName)])
	}
	changelogData := extension.Locales{}
	for _, lang := range supportedLanguages {
		changelogFileName := "CHANGELOG.md"
		if lang != extension.LanguageCodeEn {
			changelogFileName = fmt.Sprintf("CHANGELOG_%s.md", lang)
		}
		changelogData[lang] = string(data[path.Join(name, changelogFileName)])
	}

	return &Extension{
		ChartMetadata:       chartMetadata,
		DisplayName:         metadata.DisplayName,
		Description:         metadata.Description,
		KSVersion:           metadata.KsVersion,
		README:              readmeData,
		Changelog:           changelogData,
		Category:            metadata.Category,
		StaticFileDirectory: metadata.StaticFileDirectory,
		Screenshots:         metadata.Screenshots,
		Provider:            metadata.Provider,
		SupportedLanguages:  supportedLanguages,
	}, nil
}

func parseMetadata(name string, data map[string][]byte) (*extension.Metadata, error) {
	metadata := new(extension.Metadata)
	if err := yaml.Unmarshal(data[path.Join(name, extension.MetadataFilename)], metadata); err != nil {
		return nil, err
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
