package helm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"helm.sh/helm/v3/pkg/lint/rules"
	"helm.sh/helm/v3/pkg/lint/support"
	"k8s.io/apimachinery/pkg/api/validation"
	apipath "k8s.io/apimachinery/pkg/api/validation/path"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apiserver/pkg/endpoints/deprecation"
	kscheme "k8s.io/client-go/kubernetes/scheme"
)

// Lint executes 'helm Lint' against the given chart.
func Lint(l *action.Lint, paths []string, vals map[string]interface{}, metadata *chart.Metadata) *action.LintResult {
	lowestTolerance := support.ErrorSev
	if l.Strict {
		lowestTolerance = support.WarningSev
	}
	result := &action.LintResult{}
	for _, path := range paths {
		linter := lintAll(path, vals, l.Namespace, l.Strict, metadata)

		result.Messages = append(result.Messages, linter.Messages...)
		result.TotalChartsLinted++
		for _, msg := range linter.Messages {
			if msg.Severity >= lowestTolerance {
				result.Errors = append(result.Errors, msg.Err)
			}
		}
	}
	return result
}

// lintAll runs all of the available linters on the given base directory.
func lintAll(basedir string, values map[string]interface{}, namespace string, strict bool, metadata *chart.Metadata) support.Linter {
	// Using abs path to get directory context
	chartDir, _ := filepath.Abs(basedir)

	linter := support.Linter{ChartDir: chartDir}
	// For ks-extension it's not exist.
	lintChartfile(&linter, chartDir, *metadata)
	rules.ValuesWithOverrides(&linter, values)
	lintTemplates(&linter, values, namespace, strict, *metadata)
	lintDependencies(&linter, *metadata)
	return linter
}

func lintChartfile(linter *support.Linter, chartFilePath string, chartFile chart.Metadata) {
	var chartFileName = "Chart.yaml"
	if _, err := os.Stat(chartFilePath + "/" + "Chart.yaml"); os.IsNotExist(err) {
		chartFileName = "extension.yaml"
	}

	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartName(chartFile))

	// Chart metadata
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartAPIVersion(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartVersion(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartMaintainer(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartSources(chartFile))
	linter.RunLinterRule(support.InfoSev, chartFileName, validateChartIconPresence(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartIconURL(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartType(chartFile))
	linter.RunLinterRule(support.ErrorSev, chartFileName, validateChartDependencies(chartFile))
}

func lintTemplates(linter *support.Linter, values map[string]interface{}, namespace string, strict bool, metadata chart.Metadata) {
	fpath := "templates/"
	templatesPath := filepath.Join(linter.ChartDir, fpath)

	templatesDirExist := linter.RunLinterRule(support.WarningSev, fpath, validateTemplatesDir(templatesPath))

	// Templates directory is optional for now
	if !templatesDirExist {
		return
	}

	// Load chart and parse templates
	//chart, err := LoadHelmCharts(linter.ChartDir)
	chart, err := Load(linter.ChartDir, &metadata)

	chartLoaded := linter.RunLinterRule(support.ErrorSev, fpath, err)

	if !chartLoaded {
		return
	}

	options := chartutil.ReleaseOptions{
		Name:      "test-release",
		Namespace: namespace,
	}

	// lint ignores import-values
	// See https://github.com/helm/helm/issues/9658
	if err := chartutil.ProcessDependenciesWithMerge(chart, values); err != nil {
		return
	}

	cvals, err := chartutil.CoalesceValues(chart, values)
	if err != nil {
		return
	}
	valuesToRender, err := chartutil.ToRenderValues(chart, cvals, options, nil)
	if err != nil {
		linter.RunLinterRule(support.ErrorSev, fpath, err)
		return
	}
	var e engine.Engine
	e.LintMode = true
	renderedContentMap, err := e.Render(chart, valuesToRender)

	renderOk := linter.RunLinterRule(support.ErrorSev, fpath, err)

	if !renderOk {
		return
	}

	/* Iterate over all the templates to check:
	- It is a .yaml file
	- All the values in the template file is defined
	- {{}} include | quote
	- Generated content is a valid Yaml file
	- Metadata.Namespace is not set
	*/
	for _, template := range chart.Templates {
		fileName, data := template.Name, template.Data
		fpath = fileName

		linter.RunLinterRule(support.ErrorSev, fpath, validateAllowedExtension(fileName))
		// These are v3 specific checks to make sure and warn people if their
		// chart is not compatible with v3
		linter.RunLinterRule(support.WarningSev, fpath, validateNoCRDHooks(data))
		linter.RunLinterRule(support.ErrorSev, fpath, validateNoReleaseTime(data))

		// We only apply the following lint rules to yaml files
		if filepath.Ext(fileName) != ".yaml" || filepath.Ext(fileName) == ".yml" {
			continue
		}

		// NOTE: disabled for now, Refs https://github.com/helm/helm/issues/1463
		// Check that all the templates have a matching value
		// linter.RunLinterRule(support.WarningSev, fpath, validateNoMissingValues(templatesPath, valuesToRender, preExecutedTemplate))

		// NOTE: disabled for now, Refs https://github.com/helm/helm/issues/1037
		// linter.RunLinterRule(support.WarningSev, fpath, validateQuotes(string(preExecutedTemplate)))

		renderedContent := renderedContentMap[path.Join(chart.Name(), fileName)]
		if strings.TrimSpace(renderedContent) != "" {
			linter.RunLinterRule(support.WarningSev, fpath, validateTopIndentLevel(renderedContent))

			decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(renderedContent), 4096)

			// Lint all resources if the file contains multiple documents separated by ---
			for {
				// Even though K8sYamlStruct only defines a few fields, an error in any other
				// key will be raised as well
				var yamlStruct *K8sYamlStruct

				err := decoder.Decode(&yamlStruct)
				if err == io.EOF {
					break
				}

				//  If YAML linting fails here, it will always fail in the next block as well, so we should return here.
				// fix https://github.com/helm/helm/issues/11391
				if !linter.RunLinterRule(support.ErrorSev, fpath, validateYamlContent(err)) {
					return
				}
				if yamlStruct != nil {
					// NOTE: set to warnings to allow users to support out-of-date kubernetes
					// Refs https://github.com/helm/helm/issues/8596
					linter.RunLinterRule(support.WarningSev, fpath, validateMetadataName(yamlStruct))
					linter.RunLinterRule(support.WarningSev, fpath, validateNoDeprecations(yamlStruct))

					linter.RunLinterRule(support.ErrorSev, fpath, validateMatchSelector(yamlStruct, renderedContent))
					linter.RunLinterRule(support.ErrorSev, fpath, validateListAnnotations(yamlStruct, renderedContent))
				}
			}
		}
	}
}

// Dependencies runs lints against a chart's dependencies
//
// See https://github.com/helm/helm/issues/7910
func lintDependencies(linter *support.Linter, metadata chart.Metadata) {
	c, err := Load(linter.ChartDir, &metadata)
	if !linter.RunLinterRule(support.ErrorSev, "", validateChartFormat(err)) {
		return
	}

	linter.RunLinterRule(support.ErrorSev, linter.ChartDir, validateDependencyInMetadata(c))
	linter.RunLinterRule(support.WarningSev, linter.ChartDir, validateDependencyInChartsDir(c))
}

// Validation functions
func validateTemplatesDir(templatesPath string) error {
	if fi, err := os.Stat(templatesPath); err == nil {
		if !fi.IsDir() {
			return errors.New("not a directory")
		}
	}
	return nil
}

func validateAllowedExtension(fileName string) error {
	ext := filepath.Ext(fileName)
	validExtensions := []string{".yaml", ".yml", ".tpl", ".txt"}

	for _, b := range validExtensions {
		if b == ext {
			return nil
		}
	}

	return errors.Errorf("file extension '%s' not valid. Valid extensions are .yaml, .yml, .tpl, or .txt", ext)
}

var (
	crdHookSearch     = regexp.MustCompile(`"?helm\.sh/hook"?:\s+crd-install`)
	releaseTimeSearch = regexp.MustCompile(`\.Release\.Time`)
)

func validateNoCRDHooks(manifest []byte) error {
	if crdHookSearch.Match(manifest) {
		return errors.New("manifest is a crd-install hook. This hook is no longer supported in v3 and all CRDs should also exist the crds/ directory at the top level of the chart")
	}
	return nil
}

func validateNoReleaseTime(manifest []byte) error {
	if releaseTimeSearch.Match(manifest) {
		return errors.New(".Release.Time has been removed in v3, please replace with the `now` function in your templates")
	}
	return nil
}

// validateTopIndentLevel checks that the content does not start with an indent level > 0.
//
// This error can occur when a template accidentally inserts space. It can cause
// unpredictable errors depending on whether the text is normalized before being passed
// into the YAML parser. So we trap it here.
//
// See https://github.com/helm/helm/issues/8467
func validateTopIndentLevel(content string) error {
	// Read lines until we get to a non-empty one
	scanner := bufio.NewScanner(bytes.NewBufferString(content))
	for scanner.Scan() {
		line := scanner.Text()
		// If line is empty, skip
		if strings.TrimSpace(line) == "" {
			continue
		}
		// If it starts with one or more spaces, this is an error
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			return fmt.Errorf("document starts with an illegal indent: %q, which may cause parsing problems", line)
		}
		// Any other condition passes.
		return nil
	}
	return scanner.Err()
}

func validateYamlContent(err error) error {
	return errors.Wrap(err, "unable to parse YAML")
}

// validateMetadataName uses the correct validation function for the object
// Kind, or if not set, defaults to the standard definition of a subdomain in
// DNS (RFC 1123), used by most resources.
func validateMetadataName(obj *K8sYamlStruct) error {
	fn := validateMetadataNameFunc(obj)
	allErrs := field.ErrorList{}
	for _, msg := range fn(obj.Metadata.Name, false) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("metadata").Child("name"), obj.Metadata.Name, msg))
	}
	if len(allErrs) > 0 {
		return errors.Wrapf(allErrs.ToAggregate(), "object name does not conform to Kubernetes naming requirements: %q", obj.Metadata.Name)
	}
	return nil
}

// validateMetadataNameFunc will return a name validation function for the
// object kind, if defined below.
//
// Rules should match those set in the various api validations:
// https://github.com/kubernetes/kubernetes/blob/v1.20.0/pkg/apis/core/validation/validation.go#L205-L274
// https://github.com/kubernetes/kubernetes/blob/v1.20.0/pkg/apis/apps/validation/validation.go#L39
// ...
//
// Implementing here to avoid importing k/k.
//
// If no mapping is defined, returns NameIsDNSSubdomain.  This is used by object
// kinds that don't have special requirements, so is the most likely to work if
// new kinds are added.
func validateMetadataNameFunc(obj *K8sYamlStruct) validation.ValidateNameFunc {
	switch strings.ToLower(obj.Kind) {
	case "pod", "node", "secret", "endpoints", "resourcequota", // core
		"controllerrevision", "daemonset", "deployment", "replicaset", "statefulset", // apps
		"autoscaler",     // autoscaler
		"cronjob", "job", // batch
		"lease",                    // coordination
		"endpointslice",            // discovery
		"networkpolicy", "ingress", // networking
		"podsecuritypolicy",                           // policy
		"priorityclass",                               // scheduling
		"podpreset",                                   // settings
		"storageclass", "volumeattachment", "csinode": // storage
		return validation.NameIsDNSSubdomain
	case "service":
		return validation.NameIsDNS1035Label
	case "namespace":
		return validation.ValidateNamespaceName
	case "serviceaccount":
		return validation.ValidateServiceAccountName
	case "certificatesigningrequest":
		// No validation.
		// https://github.com/kubernetes/kubernetes/blob/v1.20.0/pkg/apis/certificates/validation/validation.go#L137-L140
		return func(name string, prefix bool) []string { return nil }
	case "role", "clusterrole", "rolebinding", "clusterrolebinding":
		// https://github.com/kubernetes/kubernetes/blob/v1.20.0/pkg/apis/rbac/validation/validation.go#L32-L34
		return func(name string, prefix bool) []string {
			return apipath.IsValidPathSegmentName(name)
		}
	default:
		return validation.NameIsDNSSubdomain
	}
}

var (
	// This should be set in the Makefile based on the version of client-go being imported.
	// These constants will be overwritten with LDFLAGS. The version components must be
	// strings in order for LDFLAGS to set them.
	k8sVersionMajor = "1"
	k8sVersionMinor = "20"
)

// deprecatedAPIError indicates than an API is deprecated in Kubernetes
type deprecatedAPIError struct {
	Deprecated string
	Message    string
}

func (e deprecatedAPIError) Error() string {
	msg := e.Message
	return msg
}

func validateNoDeprecations(resource *K8sYamlStruct) error {
	// if `resource` does not have an APIVersion or Kind, we cannot test it for deprecation
	if resource.APIVersion == "" {
		return nil
	}
	if resource.Kind == "" {
		return nil
	}

	runtimeObject, err := resourceToRuntimeObject(resource)
	if err != nil {
		// do not error for non-kubernetes resources
		if runtime.IsNotRegisteredError(err) {
			return nil
		}
		return err
	}
	maj, err := strconv.Atoi(k8sVersionMajor)
	if err != nil {
		return err
	}
	min, err := strconv.Atoi(k8sVersionMinor)
	if err != nil {
		return err
	}

	if !deprecation.IsDeprecated(runtimeObject, maj, min) {
		return nil
	}
	gvk := fmt.Sprintf("%s %s", resource.APIVersion, resource.Kind)
	return deprecatedAPIError{
		Deprecated: gvk,
		Message:    deprecation.WarningMessage(runtimeObject),
	}
}

func resourceToRuntimeObject(resource *K8sYamlStruct) (runtime.Object, error) {
	scheme := runtime.NewScheme()
	err := kscheme.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	gvk := schema.FromAPIVersionAndKind(resource.APIVersion, resource.Kind)
	out, err := scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	out.GetObjectKind().SetGroupVersionKind(gvk)
	return out, nil
}

// validateMatchSelector ensures that template specs have a selector declared.
// See https://github.com/helm/helm/issues/1990
func validateMatchSelector(yamlStruct *K8sYamlStruct, manifest string) error {
	switch yamlStruct.Kind {
	case "Deployment", "ReplicaSet", "DaemonSet", "StatefulSet":
		// verify that matchLabels or matchExpressions is present
		if !(strings.Contains(manifest, "matchLabels") || strings.Contains(manifest, "matchExpressions")) {
			return fmt.Errorf("a %s must contain matchLabels or matchExpressions, and %q does not", yamlStruct.Kind, yamlStruct.Metadata.Name)
		}
	}
	return nil
}

func validateListAnnotations(yamlStruct *K8sYamlStruct, manifest string) error {
	if yamlStruct.Kind == "List" {
		m := struct {
			Items []struct {
				Metadata struct {
					Annotations map[string]string
				}
			}
		}{}

		if err := yaml.Unmarshal([]byte(manifest), &m); err != nil {
			return validateYamlContent(err)
		}

		for _, i := range m.Items {
			if _, ok := i.Metadata.Annotations["helm.sh/resource-policy"]; ok {
				return errors.New("Annotation 'helm.sh/resource-policy' within List objects are ignored")
			}
		}
	}
	return nil
}

func validateChartFormat(chartError error) error {
	if chartError != nil {
		return errors.Errorf("unable to load chart\n\t%s", chartError)
	}
	return nil
}

func validateDependencyInMetadata(c *chart.Chart) (err error) {
	dependencies := map[string]struct{}{}
	missing := []string{}
	for _, dep := range c.Metadata.Dependencies {
		dependencies[dep.Name] = struct{}{}
	}
	for _, dep := range c.Dependencies() {
		if _, ok := dependencies[dep.Metadata.Name]; !ok {
			missing = append(missing, dep.Metadata.Name)
		}
	}
	if len(missing) > 0 {
		err = fmt.Errorf("chart metadata is missing these dependencies: %s", strings.Join(missing, ","))
	}
	return err
}

func validateDependencyInChartsDir(c *chart.Chart) (err error) {
	dependencies := map[string]struct{}{}
	missing := []string{}
	for _, dep := range c.Dependencies() {
		dependencies[dep.Metadata.Name] = struct{}{}
	}
	for _, dep := range c.Metadata.Dependencies {
		if _, ok := dependencies[dep.Name]; !ok {
			missing = append(missing, dep.Name)
		}
	}
	if len(missing) > 0 {
		err = fmt.Errorf("chart directory is missing these dependencies: %s", strings.Join(missing, ","))
	}
	return err
}

// K8sYamlStruct stubs a Kubernetes YAML file.
//
// DEPRECATED: In Helm 4, this will be made a private type, as it is for use only within
// the rules package.
type K8sYamlStruct struct {
	APIVersion string `json:"apiVersion"`
	Kind       string
	Metadata   k8sYamlMetadata
}

type k8sYamlMetadata struct {
	Namespace string
	Name      string
}

// Validate chartfile
func validateChartName(cf chart.Metadata) error {
	if cf.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func validateChartAPIVersion(cf chart.Metadata) error {
	if cf.APIVersion == "" {
		return errors.New("apiVersion is required. The value must be either \"v1\" or \"v2\"")
	}

	if cf.APIVersion != chart.APIVersionV1 && cf.APIVersion != chart.APIVersionV2 {
		return fmt.Errorf("apiVersion '%s' is not valid. The value must be either \"v1\" or \"v2\"", cf.APIVersion)
	}

	return nil
}

func validateChartVersion(cf chart.Metadata) error {
	if cf.Version == "" {
		return errors.New("version is required")
	}

	version, err := semver.NewVersion(cf.Version)

	if err != nil {
		return errors.Errorf("version '%s' is not a valid SemVer", cf.Version)
	}

	c, err := semver.NewConstraint(">0.0.0-0")
	if err != nil {
		return err
	}
	valid, msg := c.Validate(version)

	if !valid && len(msg) > 0 {
		return errors.Errorf("version %v", msg[0])
	}

	return nil
}

func validateChartMaintainer(cf chart.Metadata) error {
	for _, maintainer := range cf.Maintainers {
		if maintainer.Name == "" {
			return errors.New("each maintainer requires a name")
		} else if maintainer.Email != "" && !govalidator.IsEmail(maintainer.Email) {
			return errors.Errorf("invalid email '%s' for maintainer '%s'", maintainer.Email, maintainer.Name)
		} else if maintainer.URL != "" && !govalidator.IsURL(maintainer.URL) {
			return errors.Errorf("invalid url '%s' for maintainer '%s'", maintainer.URL, maintainer.Name)
		}
	}
	return nil
}

func validateChartSources(cf chart.Metadata) error {
	for _, source := range cf.Sources {
		if source == "" || !govalidator.IsRequestURL(source) {
			return errors.Errorf("invalid source URL '%s'", source)
		}
	}
	return nil
}

func validateChartIconPresence(cf chart.Metadata) error {
	if cf.Icon == "" {
		return errors.New("icon is recommended")
	}
	return nil
}

func validateChartIconURL(cf chart.Metadata) error {
	if cf.Icon != "" && !govalidator.IsRequestURL(cf.Icon) {
		return errors.Errorf("invalid icon URL '%s'", cf.Icon)
	}
	return nil
}

func validateChartType(cf chart.Metadata) error {
	if len(cf.Type) > 0 && cf.APIVersion != chart.APIVersionV2 {
		return fmt.Errorf("chart type is not valid in apiVersion '%s'. It is valid in apiVersion '%s'", cf.APIVersion, chart.APIVersionV2)
	}
	return nil
}

func validateChartDependencies(cf chart.Metadata) error {
	if len(cf.Dependencies) > 0 && cf.APIVersion != chart.APIVersionV2 {
		return fmt.Errorf("dependencies are not valid in the Chart file with apiVersion '%s'. They are valid in apiVersion '%s'", cf.APIVersion, chart.APIVersionV2)
	}
	return nil
}
