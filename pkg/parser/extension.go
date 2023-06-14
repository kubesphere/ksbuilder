package parser

import (
	"fmt"
	"os"
	"path"

	"helm.sh/helm/v3/pkg/chart"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

type Extension struct {
	ChartMetadata *chart.Metadata
	DisplayName   string
	README        []byte
	Changelog     []byte
	KSVersion     string
	Vendor        *chart.Maintainer
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
		language: "en",
	}
	for _, f := range opts {
		f(o)
	}

	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir) // nolint

	if err = unzip(zipFile, tempDir); err != nil {
		return nil, err
	}

	chartDir := path.Join(tempDir, name)
	metadata, err := extension.LoadMetadata(chartDir)
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
	readmeData, err := readFile(chartDir, readmeFileName)
	if err != nil {
		return nil, err
	}
	changelogFileName := "CHANGELOG.md"
	if o.language != extension.LanguageCodeEn {
		changelogFileName = fmt.Sprintf("CHANGELOG_%s.md", o.language)
	}
	changelogData, err := readFile(tempDir, changelogFileName)
	if err != nil {
		return nil, err
	}

	return &Extension{
		ChartMetadata: chartMetadata,
		DisplayName:   string(metadata.DisplayName[extension.LanguageCode(o.language)]),
		KSVersion:     metadata.KsVersion,
		Vendor:        metadata.Vendor,
		README:        readmeData,
		Changelog:     changelogData,
	}, nil
}
