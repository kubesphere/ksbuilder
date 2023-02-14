package extension

import (
	"os"

	"github.com/otiai10/copy"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
)

const MetadataFilename = "extension.yaml"

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

func LoadPermission(path string) (*rbacv1.ClusterRole, error) {
	_, err := os.Stat(path + "/" + "permission.yaml")
	if err == nil {
		content, err := os.ReadFile(path + "/" + "permission.yaml")
		if err != nil {
			return nil, err
		}
		var clusterRole rbacv1.ClusterRole
		err = yaml.Unmarshal(content, &clusterRole)
		if err != nil {
			return nil, err
		}
		return &clusterRole, nil
	}
	return nil, nil
}

func Load(path string) (*Extension, error) {
	metadata, err := LoadMetadata(path)
	if err != nil {
		return nil, err
	}
	clusterRole, err := LoadPermission(path)
	if err != nil {
		return nil, err
	}
	var extension Extension
	extension.Metadata = metadata
	extension.ClusterRole = clusterRole
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
