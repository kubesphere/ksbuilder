package extension

const (
	_tplChartYaml = `apiVersion: v1
name: [[ .Name ]]
description: [[ .Desc ]]
type: application
version: 0.0.1
appVersion: "0.0.1"
keywords:
  - [[ .Category ]]
home: https://kubesphere.io
sources:
  - https://github.com/kubesphere
dependencies:
  - name: frontend
    condition: frontend.enabled
  - name: backend
    condition: backend.enabled
maintainers:
  - name: [[ .Author ]]
    email: [[ .Email ]]
icon: 'data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBzdGFuZGFsb25lPSJubyI/PjwhRE9DVFlQRSBzdmcgUFVCTElDICItLy9XM0MvL0RURCBTVkcgMS4xLy9FTiIgImh0dHA6Ly93d3cudzMub3JnL0dyYXBoaWNzL1NWRy8xLjEvRFREL3N2ZzExLmR0ZCI+PHN2ZyB0PSIxNjQ2OTg2MjkyOTA5IiBjbGFzcz0iaWNvbiIgdmlld0JveD0iMCAwIDEwMjQgMTAyNCIgdmVyc2lvbj0iMS4xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHAtaWQ9IjIyMTQiIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB3aWR0aD0iMjU2IiBoZWlnaHQ9IjI1NiI+PGRlZnM+PHN0eWxlIHR5cGU9InRleHQvY3NzIj48L3N0eWxlPjwvZGVmcz48cGF0aCBkPSJNNTEyIDBDMjMwLjQgMCAwIDIzMC40IDAgNTEyczIzMC40IDUxMiA1MTIgNTEyIDUxMi0yMzAuNCA1MTItNTEyUzc5My42IDAgNTEyIDB6IG0xNzkuMiAzMDcuMmM0MC45NiAwIDc2LjggMzUuODQgNzYuOCA3Ni44cy0zNS44NCA3Ni44LTc2LjggNzYuOFM2MTQuNCA0MjQuOTYgNjE0LjQgMzg0IDY1MC4yNCAzMDcuMiA2OTEuMiAzMDcuMnogbS0zNTguNCAwYzQwLjk2IDAgNzYuOCAzNS44NCA3Ni44IDc2LjhzLTM1Ljg0IDc2LjgtNzYuOCA3Ni44LTc2LjgtMzUuODQtNzYuOC03Ni44UzI5MS44NCAzMDcuMiAzMzIuOCAzMDcuMnpNNTEyIDgxOS4yYy0xMzMuMTIgMC0yNDUuNzYtODcuMDQtMjg2LjcyLTIwNC44aDU3My40NGMtNDAuOTYgMTE3Ljc2LTE1My42IDIwNC44LTI4Ni43MiAyMDQuOHoiIHAtaWQ9IjIyMTUiPjwvcGF0aD48L3N2Zz4='
annotations:
  extensions.kubesphere.io/foo: bar
`
	_tplValuesYaml = `frontend:
  enabled: true
  image:
    repository: 
    tag: latest

backend:
  enabled: true
  image:
    repository: 
    tag: latest
`
	_tplHelmignore = `# Patterns to ignore when building packages.
# This supports shell glob matching, relative path matching, and
# negation (prefixed with !). Only one pattern per line.
.DS_Store
# Common VCS dirs
.git/
.gitignore
.bzr/
.bzrignore
.hg/
.hgignore
.svn/
# Common backup files
*.swp
*.bak
*.tmp
*.orig
*~
# Various IDEs
.project
.idea/
*.tmproj
.vscode/
`
)
