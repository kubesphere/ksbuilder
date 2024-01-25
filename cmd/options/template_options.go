package options

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/postrender"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/client-go/util/homedir"
)

type TemplateOptions struct {
	Validate    bool
	IncludeCrds bool
	SkipTests   bool
	KubeVersion string
	ExtraAPIs   []string
	ShowFiles   []string
	Client      *action.Install
	ValueOpts   *values.Options
	Settings    *cli.EnvSettings
}

func NewTemplateOptions() *TemplateOptions {
	o := &TemplateOptions{}
	o.Client = action.NewInstall(new(action.Configuration))
	o.ValueOpts = new(values.Options)
	o.Settings = cli.New()
	return o
}

func (o *TemplateOptions) AddFlags(cmd *cobra.Command, f *pflag.FlagSet) {
	addInstallFlags(cmd, f, o.Client, o.ValueOpts)
	f.StringArrayVarP(&o.ShowFiles, "show-only", "s", []string{}, "only show manifests rendered from the given templates")
	f.StringVar(&o.Client.OutputDir, "output-dir", "", "writes the executed templates to files in output-dir instead of stdout")
	f.BoolVar(&o.Validate, "validate", false, "validate your manifests against the Kubernetes cluster you are currently pointing at. This is the same validation performed on an install")
	f.BoolVar(&o.IncludeCrds, "include-crds", false, "include CRDs in the templated output")
	f.BoolVar(&o.SkipTests, "skip-tests", false, "skip tests from templated output")
	f.BoolVar(&o.Client.IsUpgrade, "is-upgrade", false, "set .Release.IsUpgrade instead of .Release.IsInstall")
	f.StringVar(&o.KubeVersion, "kube-version", "", "Kubernetes version used for Capabilities.KubeVersion")
	f.StringSliceVarP(&o.ExtraAPIs, "api-versions", "a", []string{}, "Kubernetes api versions used for Capabilities.APIVersions")
	f.BoolVar(&o.Client.UseReleaseName, "release-name", false, "use release name in the output-dir path.")
	bindPostRenderFlag(cmd, &o.Client.PostRenderer)
}

func addInstallFlags(cmd *cobra.Command, f *pflag.FlagSet, client *action.Install, valueOpts *values.Options) {
	f.BoolVar(&client.CreateNamespace, "create-namespace", false, "create the release namespace if not present")
	// --dry-run options with expected outcome:
	// - Not set means no dry run and server is contacted.
	// - Set with no value, a value of client, or a value of true and the server is not contacted
	// - Set with a value of false, none, or false and the server is contacted
	// The true/false part is meant to reflect some legacy behavior while none is equal to "".
	f.StringVar(&client.DryRunOption, "dry-run", "", "simulate an install. If --dry-run is set with no option being specified or as '--dry-run=client', it will not attempt cluster connections. Setting '--dry-run=server' allows attempting cluster connections.")
	f.Lookup("dry-run").NoOptDefVal = "client"
	f.BoolVar(&client.Force, "force", false, "force resource updates through a replacement strategy")
	f.BoolVar(&client.DisableHooks, "no-hooks", false, "prevent hooks from running during install")
	f.BoolVar(&client.Replace, "replace", false, "re-use the given name, only if that name is a deleted release which remains in the history. This is unsafe in production")
	f.DurationVar(&client.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&client.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVar(&client.WaitForJobs, "wait-for-jobs", false, "if set and --wait enabled, will wait until all Jobs have been completed before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVarP(&client.GenerateName, "generate-name", "g", false, "generate the name (and omit the NAME parameter)")
	f.StringVar(&client.NameTemplate, "name-template", "", "specify template used to name the release")
	f.StringVar(&client.Description, "description", "", "add a custom description")
	f.BoolVar(&client.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&client.DependencyUpdate, "dependency-update", false, "update dependencies if they are missing before installing the chart")
	f.BoolVar(&client.DisableOpenAPIValidation, "disable-openapi-validation", false, "if set, the installation process will not validate rendered templates against the Kubernetes OpenAPI Schema")
	f.BoolVar(&client.Atomic, "atomic", false, "if set, the installation process deletes the installation on failure. The --wait flag will be set automatically if --atomic is used")
	f.BoolVar(&client.SkipCRDs, "skip-crds", false, "if set, no CRDs will be installed. By default, CRDs are installed if not already present")
	f.BoolVar(&client.SubNotes, "render-subchart-notes", false, "if set, render subchart notes along with the parent")
	f.StringToStringVarP(&client.Labels, "labels", "l", nil, "Labels that would be added to release metadata. Should be divided by comma.")
	f.BoolVar(&client.EnableDNS, "enable-dns", false, "enable DNS lookups when rendering templates")
	addValueOptionsFlags(f, valueOpts)
	addChartPathOptionsFlags(f, &client.ChartPathOptions)

	err := cmd.RegisterFlagCompletionFunc("version", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		requiredArgs := 2
		if client.GenerateName {
			requiredArgs = 1
		}
		if len(args) != requiredArgs {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return compVersionFlag(args[requiredArgs-1], toComplete)
	})

	if err != nil {
		log.Fatal(err)
	}
}

func addValueOptionsFlags(f *pflag.FlagSet, v *values.Options) {
	f.StringSliceVarP(&v.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	f.StringArrayVar(&v.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&v.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
	f.StringArrayVar(&v.JSONValues, "set-json", []string{}, "set JSON values on the command line (can specify multiple or separate values with commas: key1=jsonval1,key2=jsonval2)")
	f.StringArrayVar(&v.LiteralValues, "set-literal", []string{}, "set a literal STRING value on the command line")
}

func addChartPathOptionsFlags(f *pflag.FlagSet, c *action.ChartPathOptions) {
	f.StringVar(&c.Version, "version", "", "specify a version constraint for the chart version to use. This constraint can be a specific tag (e.g. 1.1.1) or it may reference a valid range (e.g. ^2.0.0). If this is not specified, the latest version is used")
	f.BoolVar(&c.Verify, "verify", false, "verify the package before using it")
	f.StringVar(&c.Keyring, "keyring", defaultKeyring(), "location of public keys used for verification")
	f.StringVar(&c.RepoURL, "repo", "", "chart repository url where to locate the requested chart")
	f.StringVar(&c.Username, "username", "", "chart repository username where to locate the requested chart")
	f.StringVar(&c.Password, "password", "", "chart repository password where to locate the requested chart")
	f.StringVar(&c.CertFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&c.KeyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.BoolVar(&c.InsecureSkipTLSverify, "insecure-skip-tls-verify", false, "skip tls certificate checks for the chart download")
	f.BoolVar(&c.PlainHTTP, "plain-http", false, "use insecure HTTP connections for the chart download")
	f.StringVar(&c.CaFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
	f.BoolVar(&c.PassCredentialsAll, "pass-credentials", false, "pass credentials to all domains")
}

const (
	outputFlag         = "output"
	postRenderFlag     = "post-renderer"
	postRenderArgsFlag = "post-renderer-args"
)

type postRendererOptions struct {
	renderer   *postrender.PostRenderer
	binaryPath string
	args       []string
}

type postRendererString struct {
	options *postRendererOptions
}

func (p *postRendererString) String() string {
	return p.options.binaryPath
}

func (p *postRendererString) Type() string {
	return "postRendererString"
}

func (p *postRendererString) Set(val string) error {
	if val == "" {
		return nil
	}
	p.options.binaryPath = val
	pr, err := postrender.NewExec(p.options.binaryPath, p.options.args...)
	if err != nil {
		return err
	}
	*p.options.renderer = pr
	return nil
}

type postRendererArgsSlice struct {
	options *postRendererOptions
}

func (p *postRendererArgsSlice) String() string {
	return "[" + strings.Join(p.options.args, ",") + "]"
}

func (p *postRendererArgsSlice) Type() string {
	return "postRendererArgsSlice"
}

func (p *postRendererArgsSlice) Set(val string) error {

	// a post-renderer defined by a user may accept empty arguments
	p.options.args = append(p.options.args, val)

	if p.options.binaryPath == "" {
		return nil
	}
	// overwrite if already create PostRenderer by `post-renderer` flags
	pr, err := postrender.NewExec(p.options.binaryPath, p.options.args...)
	if err != nil {
		return err
	}
	*p.options.renderer = pr
	return nil
}

func (p *postRendererArgsSlice) Append(val string) error {
	p.options.args = append(p.options.args, val)
	return nil
}

func (p *postRendererArgsSlice) Replace(val []string) error {
	p.options.args = val
	return nil
}

func (p *postRendererArgsSlice) GetSlice() []string {
	return p.options.args
}

func bindPostRenderFlag(cmd *cobra.Command, varRef *postrender.PostRenderer) {
	p := &postRendererOptions{varRef, "", []string{}}
	cmd.Flags().Var(&postRendererString{p}, postRenderFlag, "the path to an executable to be used for post rendering. If it exists in $PATH, the binary will be used, otherwise it will try to look for the executable at the given path")
	cmd.Flags().Var(&postRendererArgsSlice{p}, postRenderArgsFlag, "an argument to the post-renderer (can specify multiple)")
}

// defaultKeyring returns the expanded path to the default keyring.
func defaultKeyring() string {
	if v, ok := os.LookupEnv("GNUPGHOME"); ok {
		return filepath.Join(v, "pubring.gpg")
	}
	return filepath.Join(homedir.HomeDir(), ".gnupg", "pubring.gpg")
}

var settings = cli.New()

func compVersionFlag(chartRef string, toComplete string) ([]string, cobra.ShellCompDirective) {
	chartInfo := strings.Split(chartRef, "/")
	if len(chartInfo) != 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	repoName := chartInfo[0]
	chartName := chartInfo[1]

	path := filepath.Join(settings.RepositoryCache, helmpath.CacheIndexFile(repoName))

	var versions []string
	if indexFile, err := repo.LoadIndexFile(path); err == nil {
		for _, details := range indexFile.Entries[chartName] {
			appVersion := details.Metadata.AppVersion
			appVersionDesc := ""
			if appVersion != "" {
				appVersionDesc = fmt.Sprintf("App: %s, ", appVersion)
			}
			created := details.Created.Format("January 2, 2006")
			createdDesc := ""
			if created != "" {
				createdDesc = fmt.Sprintf("Created: %s ", created)
			}
			deprecated := ""
			if details.Metadata.Deprecated {
				deprecated = "(deprecated)"
			}
			versions = append(versions, fmt.Sprintf("%s\t%s%s%s", details.Metadata.Version, appVersionDesc, createdDesc, deprecated))
		}
	}

	return versions, cobra.ShellCompDirectiveNoFileComp
}
