package extension

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/otiai10/copy"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"
)

const MetadataFilename = "extension.yaml"

func LoadMetadata(path string) (*Metadata, error) {
	content, err := os.ReadFile(path + "/" + MetadataFilename)
	if err != nil {
		return nil, err
	}
	metadata := new(Metadata)
	if err = yaml.Unmarshal(content, metadata); err != nil {
		return nil, err
	}
	if err = metadata.Validate(); err != nil {
		return nil, err
	}
	if err = metadata.Init(path); err != nil {
		return nil, err
	}
	return metadata, nil
}

func LoadApplicationClass(name, p, tempDir string) error {
	var b bytes.Buffer
	defer func() {
		b.Reset()
	}()

	content, err := os.ReadFile(p + "/applicationclass.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var appClass ApplicationClass
	err = yaml.Unmarshal(content, &appClass)
	if err != nil {
		return err
	}

	// Validate
	if len(appClass.Name) != 0 {
		filePath := path.Join(tempDir, "charts/applicationclass")
		err = os.MkdirAll(filePath, 0644)
		if err != nil {
			return err
		}

		c := &chart.Metadata{
			APIVersion: "v1",
			Name:       appClass.Name,
			Version:    appClass.PackageVersion,
			AppVersion: appClass.AppVersion,
		}
		appClassChart, err := yaml.Marshal(c)
		if err != nil {
			return err
		}

		err = os.WriteFile(filePath+"/Chart.yaml", appClassChart, 0644)
		if err != nil {
			return err
		}

		err = os.MkdirAll(filePath+"/templates", 0644)
		if err != nil {
			return err
		}

		if appClass.Provisioner == "kubesphere.io/helm-application" {
			var cmName = fmt.Sprintf("application-%s-chart", name)
			appClass.Parameters = make(map[string]string)
			appClass.Parameters["configmap"] = cmName
			appClass.Parameters["namespace"] = "kubesphere-system"

			err = copy.Copy(tempDir+"/application-package.yaml", filePath+"/templates/application-package.yaml")
			if err != nil {
				return err
			}
		}

		err = applicationClassTmpl.Execute(&b, appClass)
		if err != nil {
			return err
		}
		err = os.WriteFile(filePath+"/templates/applicationclass.yaml", b.Bytes(), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func Load(path string) (*Extension, error) {
	metadata, err := LoadMetadata(path)
	if err != nil {
		return nil, err
	}

	var extension Extension
	extension.Metadata = metadata
	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir) // nolint

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

	err = LoadApplicationClass(metadata.Name, path, tempDir)
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
