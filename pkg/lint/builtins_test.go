package lint

import (
	"os"
	"testing"
	"text/template"

	"k8s.io/klog/v2"
)

var templatesFileTmpl = template.Must(template.New("templates").Parse(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      {{- if .HasNodeSelector }}
      {{ "{{- if .Values.global.nodeSelector }}" }}
      nodeSelector: {{ "{{- include \"common.tplvalues.render\" ( dict \"value\" .Values.global.nodeSelector \"context\" $) | nindent 8 }}" }}
      {{ "{{- end }}" }}
      {{- end }}
      containers:
        - name: nginx
          {{- if .HasImageRegistry }}
          image: {{ "{{ .Values.global.imageRegistry }}" }}/nginx:latest
 		  {{- else }}	
          image: nginx:latest
          {{- end }}
          ports:
            - containerPort: 80
`))

var extensionFileTmpl = template.Must(template.New("extension").Parse(`apiVersion: v1
name: lint-test
version: 0.1.0
displayName: 
  en: lint-test
  zh: lint测试
description:
  en: lint-test
  zh: 这是针对lint命令的测试chart
category: test
provider:
  en:
    name: ksbuilder
  zh:
    name: ksbuilder
icon: ./test.svg
{{- if .Images }}
images: {{ .Images }}
{{- end }}
`))

func TestWithBuiltins(t *testing.T) {
	testcases := []struct {
		name             string
		templatesFileVal map[string]any
		extensionFileVal map[string]any
	}{
		{
			name: "images in extension is empty",
			templatesFileVal: map[string]any{
				"HasNodeSelector":  true,
				"HasImageRegistry": true,
			},
			extensionFileVal: map[string]any{},
		},
		{
			name: "global.imageRegistry and global.nodeSelector is not empty",
			templatesFileVal: map[string]any{
				"HasNodeSelector":  true,
				"HasImageRegistry": true,
			},
			extensionFileVal: map[string]any{
				"Images": []string{"a"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// create templates file
			tf, err := os.Create("testdata/templates/deployment.yaml")
			if err != nil {
				klog.Error(err)
				return
			}
			defer os.Remove("testdata/templates/deployment.yaml")

			if err := templatesFileTmpl.Execute(tf, tc.templatesFileVal); err != nil {
				klog.Error(err)
				return
			}
			// create extension file
			ef, err := os.Create("testdata/extension.yaml")
			if err != nil {
				klog.Error(err)
				return
			}
			defer os.Remove("testdata/extension.yaml")

			if err := extensionFileTmpl.Execute(ef, tc.extensionFileVal); err != nil {
				klog.Error(err)
				return
			}
			if err := WithBuiltins([]string{"testdata"}); err != nil {
				t.Fatal(err)
			}

		})
	}

}
