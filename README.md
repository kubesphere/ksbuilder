# Introduction

ksbuilder is a CLI tool to create, publish, and manage KubeSphere extensions.

## Install

Download the [latest ksbuilder release](https://github.com/kubesphere/ksbuilder/releases) and then install it to `/usr/local/bin`:
```shell
tar xvzf ksbuilder_<version>_<arch>.tar.gz -C /usr/local/bin/
```

## Create your first KubeSphere extension

You can use `ksbuilder create` to create a KubeSphere extension interactively.

```
$ cd <project-directory>
$ ksbuilder create

Please input extension name: test
✔ ai-machine-learning
Please input extension author: ia
Please input Email (optional):
Please input author's URL (optional):
Directory: /path/test

The extension charts has been created.
```

The extension directory created looks like below:

```
.
├── README.md
├── README_zh.md
├── applicationclass.yaml
├── charts
│   ├── backend
│   │   ├── Chart.yaml
│   │   ├── templates
│   │   │   ├── NOTES.txt
│   │   │   ├── deployment.yaml
│   │   │   ├── extensions.yaml
│   │   │   ├── helps.tpl
│   │   │   ├── service.yaml
│   │   │   └── tests
│   │   │       └── test-connection.yaml
│   │   └── values.yaml
│   └── frontend
│       ├── Chart.yaml
│       ├── templates
│       │   ├── NOTES.txt
│       │   ├── deployment.yaml
│       │   ├── extensions.yaml
│       │   ├── helps.tpl
│       │   ├── service.yaml
│       │   └── tests
│       │       └── test-connection.yaml
│       └── values.yaml
├── extension.yaml
├── permissions.yaml
├── static
│   ├── favicon.svg
│   └── screenshots
│       └── screenshot.png
└── values.yaml
```

Then you can customize your extension like below:

- Specify the default backend and frontend images

```
frontend:
  enabled: true
  image:
repository:  <YOUR_REPO>/<extension-name>
    tag: latest

backend:
  enabled: true
  image:
    repository: <YOUR_REPO>/<extension-name>
    tag: latest
```

- Add `APIService` definition to the backend `extensions.yaml`

```yaml
apiVersion: extensions.kubesphere.io/v1alpha1
kind: APIService
metadata:
  name: v1alpha1.<extension-name>.kubesphere.io
spec:
  group: <extension-name>.kubesphere.io
  version: v1alpha1                                      
  url: http://<extension-name>-backend.default.svc:8080
status:
  state: Available
```

- Add `JSBundle` definition to the frontend `extensions.yaml`

```yaml
apiVersion: extensions.kubesphere.io/v1alpha1
kind: JSBundle
metadata:
  name: v1alpha1.<extension-name>.kubesphere.io
spec:
  rawFrom:
    url: http://<extension-name>-frontend.default.svc/dist/<extension-name>-frontend/index.js
status:
  state: Available
  link: /dist/<extension-name>-frontend/index.js
```

Please refer to [KubeSphere extension development guide](https://dev-guide.kubesphere.io/extension-dev-guide/zh/development-procedure/) for more details on extension development.

## Publish/Unpublish your KubeSphere extension

You can publish/unpublish KubeSphere extension to KubeSphere cluster once it's ready:

```shell
ksbuilder publish/unpublish <extension-name>
```
