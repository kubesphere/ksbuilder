package extension

import (
	"bytes"
	"os"
	"strings"
	"text/template"
)

type Config struct {
	Name     string
	Author   string
	Email    string
	Category string
}

const (
	_tplKeyExtensionYaml = iota
	_tplKeyPermissionYaml
	_tplKeyValuesYaml
	_tplKeyHelmignore
	_tplKeyFavicon
	_tplKeyReadmeZh
	_tplKeyReadmeEn
	_tplKeyFrontendChartYaml
	_tplKeyFrontendChartValues
	_tplKeyFrontendDeploymentYaml
	_tplKeyFrontendServiceYaml
	_tplKeyFrontendExtensionsYaml
	_tplKeyFrontendNOTESTxt
	_tplKeyFrontendHelps
	_tplKeyFrontendTestConnection
	_tplKeyBackendChartYaml
	_tplKeyBackendChartValues
	_tplKeyBackendDeploymentYaml
	_tplKeyBackendServiceYaml
	_tplKeyBackendExtensionsYaml
	_tplKeyBackendNOTESTxt
	_tplKeyBackendHelps
	_tplKeyBackendTestConnection
)

var (
	// files key => path
	files = map[int]string{
		_tplKeyExtensionYaml:          "/extension.yaml",
		_tplKeyPermissionYaml:         "/permission.yaml",
		_tplKeyValuesYaml:             "/values.yaml",
		_tplKeyHelmignore:             "/.helmignore",
		_tplKeyFavicon:                "/favicon.svg",
		_tplKeyReadmeZh:               "/README_zh.md",
		_tplKeyReadmeEn:               "/README.md",
		_tplKeyFrontendChartYaml:      "/charts/frontend/Chart.yaml",
		_tplKeyFrontendChartValues:    "/charts/frontend/values.yaml",
		_tplKeyFrontendDeploymentYaml: "/charts/frontend/templates/deployment.yaml",
		_tplKeyFrontendServiceYaml:    "/charts/frontend/templates/service.yaml",
		_tplKeyFrontendExtensionsYaml: "/charts/frontend/templates/extensions.yaml",
		_tplKeyFrontendNOTESTxt:       "/charts/frontend/templates/NOTES.txt",
		_tplKeyFrontendHelps:          "/charts/frontend/templates/helps.tpl",
		_tplKeyFrontendTestConnection: "/charts/frontend/templates/tests/test-connection.yaml",
		_tplKeyBackendChartYaml:       "/charts/backend/Chart.yaml",
		_tplKeyBackendChartValues:     "/charts/backend/values.yaml",
		_tplKeyBackendDeploymentYaml:  "/charts/backend/templates/deployment.yaml",
		_tplKeyBackendServiceYaml:     "/charts/backend/templates/service.yaml",
		_tplKeyBackendExtensionsYaml:  "/charts/backend/templates/extensions.yaml",
		_tplKeyBackendNOTESTxt:        "/charts/backend/templates/NOTES.txt",
		_tplKeyBackendHelps:           "/charts/backend/templates/helps.tpl",
		_tplKeyBackendTestConnection:  "/charts/backend/templates/tests/test-connection.yaml",
	}
	// tpls key => content
	tpls = map[int]string{
		_tplKeyExtensionYaml:          tplExtensionYaml,
		_tplKeyPermissionYaml:         tplPermissionYaml,
		_tplKeyValuesYaml:             tplValuesYaml,
		_tplKeyHelmignore:             tplHelmignore,
		_tplKeyFavicon:                tplFavicon,
		_tplKeyReadmeZh:               tplReadmeZh,
		_tplKeyReadmeEn:               tplReadmeEn,
		_tplKeyFrontendChartYaml:      _tplFrontendChartYaml,
		_tplKeyFrontendChartValues:    _tplFrontendValuesYaml,
		_tplKeyFrontendDeploymentYaml: _tplFrontendDeploymentYaml,
		_tplKeyFrontendServiceYaml:    _tplFrontendServiceYaml,
		_tplKeyFrontendExtensionsYaml: _tplFrontendExtensionsYaml,
		_tplKeyFrontendNOTESTxt:       _tplFrontendNOTESTxt,
		_tplKeyFrontendHelps:          _tplFrontendHelps,
		_tplKeyFrontendTestConnection: _tplFrontendTestConnection,
		_tplKeyBackendChartYaml:       _tplBackendChartYaml,
		_tplKeyBackendChartValues:     _tplBackendValuesYaml,
		_tplKeyBackendDeploymentYaml:  _tplBackendDeploymentYaml,
		_tplKeyBackendServiceYaml:     _tplBackendServiceYaml,
		_tplKeyBackendExtensionsYaml:  _tplBackendExtensionsYaml,
		_tplKeyBackendNOTESTxt:        _tplBackendNOTESTxt,
		_tplKeyBackendHelps:           _tplBackendHelps,
		_tplKeyBackendTestConnection:  _tplBackendTestConnection,
	}
)

func Create(p string, config Config) (err error) {
	if err = os.MkdirAll(p, 0755); err != nil {
		return err
	}

	for t, v := range files {
		i := strings.LastIndex(v, "/")
		if i > 0 {
			dir := v[:i]
			if err = os.MkdirAll(p+dir, 0755); err != nil {
				return
			}
		}
		if err = write(p+v, tpls[t], config); err != nil {
			return
		}
	}
	return
}

func write(name, tpl string, data interface{}) (err error) {
	body, err := parse(tpl, data)
	if err != nil {
		return
	}
	return os.WriteFile(name, body, 0644)
}

func parse(s string, data interface{}) ([]byte, error) {
	t, err := template.New("").Delims("[[", "]]").Parse(s)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
