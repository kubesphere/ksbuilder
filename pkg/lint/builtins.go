package lint

import (
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

func WithBuiltins(paths []string) error {
	ext, err := extension.Load(paths[0])
	if err != nil {
		return err
	}

	// check images if exists
	if len(ext.Metadata.Images) == 0 {
		fmt.Printf("WARNING: extension %s has no images\n", paths[0])
	}
	chartRequested, err := loader.Load(paths[0])
	if err != nil {
		return err
	}

	// check if value is valid
	valueValidators := []*validator{
		{
			name: "global.imageRegistry",
			key:  rand.String(12),
			valueFunc: func(v *validator) string {
				return fmt.Sprintf("Values.global.imageRegistry=%s", v.key)
			},
			output: make(map[string]string),
		},
		{
			name: "global.nodeSelector",
			key:  rand.String(12),
			valueFunc: func(v *validator) string {
				return fmt.Sprintf("Values.global.nodeSelector=\"kubernetes.io/os: %s\"", v.key)
			},
			output: make(map[string]string),
		},
	}

	valueOpts := &values.Options{}
	for _, vt := range valueValidators {
		valueOpts.Values = append(valueOpts.Values, vt.InitValue())
	}

	p := getter.All(cli.New())
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return err
	}

	if err := chartutil.ProcessDependenciesWithMerge(chartRequested, vals); err != nil {
		return err
	}

	files, err := engine.Render(chartRequested, vals)
	if err != nil {
		return err
	}

	for name, content := range files {
		for _, vt := range valueValidators {
			if err := vt.Validate(name, content); err != nil {
				return err
			}
		}

	}

	for _, vt := range valueValidators {
		vt.Output()
	}

	return nil
}

type validator struct {
	name      string
	key       string
	valueFunc func(v *validator) string
	output    map[string]string
}

func (v *validator) InitValue() string {
	return v.valueFunc(v)
}

func (v *validator) Validate(fileName string, fileData string) error {
	yamlArr := strings.Split(fileData, "---")
	for _, y := range yamlArr {
		if strings.Contains(y, v.key) {
			yamlMap := make(map[string]any)
			if err := yaml.Unmarshal([]byte(y), &yamlMap); err != nil {
				return err
			}
			v.output[fileName] += fmt.Sprintf("name: %s, groupVersion: %s, kind: %s \n", yamlMap["metadata"].(map[string]any)["name"], yamlMap["apiVersion"], yamlMap["kind"])
		}
	}
	return nil
}

func (v *validator) Output() {
	if len(v.output) != 0 {
		fmt.Printf("\"%s\" is valid in: \n", v.name)
		for fileName, keyStr := range v.output {
			fmt.Printf("%s: \n%s", fileName, keyStr)
		}
		fmt.Print("\n")
	}
}
