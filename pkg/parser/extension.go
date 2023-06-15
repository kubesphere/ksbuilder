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
	DisplayName         string
	Description         string
	README              []byte
	Changelog           []byte
	KSVersion           string
	StaticFileDirectory string
	Screenshots         []string
	Provider            map[string]*chart.Maintainer
	SupportedLanguages  []string
}

type options struct {
	language string
}

// WithLanguage specifies the language to use when parsing the extension.
// The default value is `en`.
func WithLanguage(language string) func(opts *options) {
	return func(opts *options) {
		opts.language = language
	}
}

func ParseExtension(name string, zipFile []byte, opts ...func(*options)) (*Extension, error) {
	o := &options{
		language: extension.LanguageCodeEn,
	}
	for _, f := range opts {
		f(o)
	}

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

	readmeFileName := "README.md"
	if o.language != extension.LanguageCodeEn {
		readmeFileName = fmt.Sprintf("README_%s.md", o.language)
	}
	readmeData := data[path.Join(name, readmeFileName)]
	changelogFileName := "CHANGELOG.md"
	if o.language != extension.LanguageCodeEn {
		changelogFileName = fmt.Sprintf("CHANGELOG_%s.md", o.language)
	}
	changelogData := data[path.Join(name, changelogFileName)]

	displayNameLanguages := sets.NewString(getKeys(metadata.DisplayName)...)
	descriptionLanguages := sets.NewString(getKeys(metadata.Description)...)

	return &Extension{
		ChartMetadata:       chartMetadata,
		DisplayName:         string(metadata.DisplayName[extension.LanguageCode(o.language)]),
		Description:         string(metadata.Description[extension.LanguageCode(o.language)]),
		KSVersion:           metadata.KsVersion,
		README:              readmeData,
		Changelog:           changelogData,
		StaticFileDirectory: metadata.StaticFileDirectory,
		Screenshots:         metadata.Screenshots,
		Provider:            metadata.Provider,
		SupportedLanguages:  displayNameLanguages.Intersection(descriptionLanguages).UnsortedList(),
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

func getKeys(data extension.Locales) []string {
	ret := make([]string, 0, len(data))
	for k := range data {
		ret = append(ret, string(k))
	}
	return ret
}
