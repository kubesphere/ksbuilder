package extension

import (
	"os"

	"github.com/otiai10/copy"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"
)

const (
	MetadataFilename    = "extension.yaml"
	PermissionsFilename = "permissions.yaml"
)

func LoadMetadata(path string) (*Metadata, error) {
	content, err := os.ReadFile(path + "/" + MetadataFilename)
	if err != nil {
		return nil, err
	}
	var metadata Metadata
	err = yaml.Unmarshal(content, &metadata)
	if err != nil {
		return nil, err
	}
	err = metadata.Validate(path)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func Load(path string) (*Extension, error) {
	metadata, err := LoadMetadata(path)
	if err != nil {
		return nil, err
	}

	permissions, err := os.ReadFile(path + "/" + PermissionsFilename)
	if err != nil {
		return nil, err
	}
	var extension Extension
	extension.Metadata = metadata
	extension.Permissions = permissions

	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return nil, err
	}

	err = copy.Copy(path, tempDir)
	if err != nil {
		return nil, err
	}

	chartYaml, err := metadata.ToChartYaml()
	if err != nil {
		return nil, err
	}

	chartMetadata, err := yaml.Marshal(chartYaml)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(tempDir+"/Chart.yaml", chartMetadata, 0644)
	if err != nil {
		return nil, err
	}

	ch, err := loader.LoadDir(tempDir)
	if err != nil {
		return nil, err
	}
	chartFilename, err := chartutil.Save(ch, tempDir)
	if err != nil {
		return nil, err
	}
	chartContent, err := os.ReadFile(chartFilename)
	if err != nil {
		return nil, err
	}

	extension.ChartData = chartContent
	return &extension, nil
}
