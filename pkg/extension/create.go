package extension

import (
	"embed"
	"fmt"
	"io/fs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"text/template"
)

type Config struct {
	Name     string
	Category string
	Author   string
	Email    string
	URL      string
}

//go:embed templates
var templates embed.FS

func Create(p string, config Config) error {
	return fs.WalkDir(templates, ".", func(templatePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		innerPath := strings.TrimPrefix(templatePath, "templates")
		if d.IsDir() {
			if innerPath != "" { // Ignore the templates parent directory
				if err = os.MkdirAll(filepath.Join(p, innerPath), 0755); err != nil {
					return err
				}
			}
			return nil
		}

		t, err := template.New(path.Base(templatePath)).Delims("[[", "]]").ParseFS(templates, templatePath)
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(p, innerPath))
		if err != nil {
			return err
		}
		defer f.Close() // nolint
		if err = t.Execute(f, config); err != nil {
			return err
		}
		return nil
	})
}

func CreateAppChart(p string, name string, chart []byte) error {
	var cmName = fmt.Sprintf("application-%s-chart", name)

	var cm = v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: "kubesphere-system",
		},
		BinaryData: map[string][]byte{
			"chart.tgz": chart,
		},
	}
	cmByte, err := yaml.Marshal(cm)
	if err != nil {
		return err
	}

	filePath := path.Join(p, "application-package.yaml")
	if err = os.WriteFile(filePath, cmByte, 0644); err != nil {
		return err
	}
	return nil
}
