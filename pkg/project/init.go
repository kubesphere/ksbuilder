package project

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

const (
	_tplTypeDockerfile = iota
	_tplTypeDockerignore
)

var (
	// files type => path
	files = map[int]string{
		_tplTypeDockerfile:   "/Dockerfile",
		_tplTypeDockerignore: "/.dockerignore",
	}
	// tpls type => content
	tpls = map[int]string{
		_tplTypeDockerfile:   _tplDockerfile,
		_tplTypeDockerignore: _tplDockerignore,
	}
)

func Create(p string) (err error) {
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
		if err = write(p+v, tpls[t], p); err != nil {
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
	return ioutil.WriteFile(name, body, 0644)
}

func parse(s string, data interface{}) ([]byte, error) {
	t, err := template.New("").Parse(s)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
