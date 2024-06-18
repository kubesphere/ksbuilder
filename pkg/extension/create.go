package extension

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/mholt/archiver/v3"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Config struct {
	Name     string
	Category string
	Author   string
	Email    string
	URL      string
}
type ConfigSimple struct {
	Name string
}
type ConfigApp struct {
	Name           string
	Maintainers    string
	AppName        string
	Version        string
	AppHome        string
	Abstraction    string
	AppVersionName string
	ZipName        string
}

//go:embed templates
var Templates embed.FS

//go:embed templatessimple
var Templatessimple embed.FS

//go:embed templatesapp
var Templatesapp embed.FS

func copyZipFile(srcPath, dstPath string) error {
	fileName := filepath.Base(srcPath)
	dstFilePath := filepath.Join(dstPath, fileName)
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func generateRandomString() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
func CreateApp(chartPath string) error {
	pwd, _ := os.Getwd()
	f, err := os.ReadFile(chartPath)
	if err != nil {
		return err
	}
	chartPack, err := loader.LoadArchive(bytes.NewReader(f))
	if err != nil {
		return err
	}
	root := path.Join(pwd, chartPack.Name())
	appName := fmt.Sprintf("%s-%s", chartPack.Name(), generateRandomString())
	appVersionName := fmt.Sprintf("%s-%s-%s", appName, chartPack.AppVersion(), generateRandomString())
	err = os.MkdirAll(root, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(root, "charts", "base", "files"), 0755)
	if err != nil {
		return err
	}

	err = copyZipFile(chartPath, filepath.Join(root, "charts", "base", "files"))
	if err != nil {
		return err
	}

	extensionConfig := ConfigApp{
		Name:           chartPack.Name(),
		AppName:        appName,
		Version:        chartPack.Metadata.Version,
		AppHome:        chartPack.Metadata.Home,
		Abstraction:    chartPack.Metadata.Description,
		AppVersionName: appVersionName,
		ZipName:        filepath.Base(chartPath),
	}
	if len(chartPack.Metadata.Maintainers) > 0 {
		extensionConfig.Maintainers = chartPack.Metadata.Maintainers[0].Name
	} else {
		extensionConfig.Maintainers = "admin"
	}

	err = Create(root, extensionConfig, Templatesapp, "templatesapp")
	return err
}

func CreateSimple(chartPath string) error {
	pwd, _ := os.Getwd()
	f, err := os.ReadFile(chartPath)
	if err != nil {
		return err
	}
	chartPack, err := loader.LoadArchive(bytes.NewReader(f))
	if err != nil {
		return err
	}
	root := path.Join(pwd, chartPack.Name())
	err = os.MkdirAll(root, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(root, "charts"), 0755)
	if err != nil {
		return err
	}
	err = archiver.Unarchive(chartPath, filepath.Join(root, "charts"))
	if err != nil {
		return err
	}

	extensionConfig := ConfigSimple{
		Name: chartPack.Name(),
	}
	err = Create(root, extensionConfig, Templatessimple, "templatessimple")
	return err
}

func Create(p string, config any, temp embed.FS, trimPrefix string) error {
	return fs.WalkDir(temp, ".", func(templatePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		innerPath := strings.TrimPrefix(templatePath, trimPrefix)
		if d.IsDir() {
			if innerPath != "" { // Ignore the templates parent directory
				if err = os.MkdirAll(filepath.Join(p, innerPath), 0755); err != nil {
					return err
				}
			}
			return nil
		}

		t, err := template.New(path.Base(templatePath)).Delims("[[", "]]").ParseFS(temp, templatePath)
		if err != nil {
			return err
		}
		f, err := os.Create(filepath.Join(p, innerPath))
		if err != nil {
			return err
		}
		defer f.Close() // nolint
		return t.Execute(f, config)
	})
}

func CreateAppChart(p string, name string, chart []byte) error {
	var cmName = fmt.Sprintf("application-%s-chart", name)

	var cm = corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: kubeSphereSystem,
		},
		BinaryData: map[string][]byte{
			configMapDataKey: chart,
		},
	}
	cmByte, err := yaml.Marshal(cm)
	if err != nil {
		return err
	}

	filePath := path.Join(p, "application-package.yaml")
	return os.WriteFile(filePath, cmByte, 0644)
}
