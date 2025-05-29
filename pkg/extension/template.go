package extension

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"

	"github.com/kubesphere/ksbuilder/cmd/options"
	"github.com/kubesphere/ksbuilder/pkg/api"
	"github.com/kubesphere/ksbuilder/pkg/helm"
)

func PrintTemplate(args []string, o *options.TemplateOptions, out io.Writer) error {
	name, chart, err := o.Client.NameAndChart(args)
	if err != nil {
		return err
	}
	o.Client.ReleaseName = name
	cp, err := o.Client.LocateChart(chart, o.Settings)
	if err != nil {
		return err
	}

	// set metadata
	metadata, err := api.LoadMetadata(cp)
	if err != nil {
		return err
	}

	rel, err := helm.Template(args, o, cp, metadata.ToChartYaml(), out)
	if err != nil && !o.Settings.Debug {
		if rel != nil {
			return fmt.Errorf("%w\n\nUse --debug flag to render out invalid YAML", err)
		}
		return err
	}

	// We ignore a potential error here because, when the --debug flag was specified,
	// we always want to print the YAML, even if it is not valid. The error is still returned afterwards.
	if rel != nil {
		var manifests bytes.Buffer
		_, _ = fmt.Fprintln(&manifests, strings.TrimSpace(rel.Manifest))
		if !o.Client.DisableHooks {
			fileWritten := make(map[string]bool)
			for _, m := range rel.Hooks {
				if o.SkipTests && isTestHook(m) {
					continue
				}
				if o.Client.OutputDir == "" {
					_, _ = fmt.Fprintf(&manifests, "---\n# Source: %s\n%s\n", m.Path, m.Manifest)
				} else {
					newDir := o.Client.OutputDir
					if o.Client.UseReleaseName {
						newDir = filepath.Join(o.Client.OutputDir, o.Client.ReleaseName)
					}
					_, err := os.Stat(filepath.Join(newDir, m.Path))
					if err == nil {
						fileWritten[m.Path] = true
					}

					err = writeToFile(newDir, m.Path, m.Manifest, fileWritten[m.Path])
					if err != nil {
						return err
					}
				}

			}
		}

		// if we have a list of files to render, then check that each of the
		// provided files exists in the chart.
		if len(o.ShowFiles) > 0 {
			// This is necessary to ensure consistent manifest ordering when using --show-only
			// with globs or directory names.
			splitManifests := releaseutil.SplitManifests(manifests.String())
			manifestsKeys := make([]string, 0, len(splitManifests))
			for k := range splitManifests {
				manifestsKeys = append(manifestsKeys, k)
			}
			sort.Sort(releaseutil.BySplitManifestsOrder(manifestsKeys))

			manifestNameRegex := regexp.MustCompile("# Source: [^/]+/(.+)")
			var manifestsToRender []string
			for _, f := range o.ShowFiles {
				missing := true
				// Use linux-style filepath separators to unify user's input path
				f = filepath.ToSlash(f)
				for _, manifestKey := range manifestsKeys {
					manifest := splitManifests[manifestKey]
					submatch := manifestNameRegex.FindStringSubmatch(manifest)
					if len(submatch) == 0 {
						continue
					}
					manifestName := submatch[1]
					// manifest.Name is rendered using linux-style filepath separators on Windows as
					// well as macOS/linux.
					manifestPathSplit := strings.Split(manifestName, "/")
					// manifest.Path is connected using linux-style filepath separators on Windows as
					// well as macOS/linux
					manifestPath := strings.Join(manifestPathSplit, "/")

					// if the filepath provided matches a manifest path in the
					// chart, render that manifest
					if matched, _ := filepath.Match(f, manifestPath); !matched {
						continue
					}
					manifestsToRender = append(manifestsToRender, manifest)
					missing = false
				}
				if missing {
					return fmt.Errorf("could not find template %s in chart", f)
				}
			}
			for _, m := range manifestsToRender {
				_, _ = fmt.Fprintf(out, "---\n%s\n", m)
			}
		} else {
			_, _ = fmt.Fprintf(out, "%s", manifests.String())
		}
	}

	return err
}

func isTestHook(h *release.Hook) bool {
	for _, e := range h.Events {
		if e == release.HookTest {
			return true
		}
	}
	return false
}

const defaultDirectoryPermission = 0755

// write the <data> to <output-dir>/<name>. <append> controls if the file is created or content will be appended
func writeToFile(outputDir string, name string, data string, append bool) error {
	outfileName := strings.Join([]string{outputDir, name}, string(filepath.Separator))

	err := ensureDirectoryForFile(outfileName)
	if err != nil {
		return err
	}

	f, err := createOrOpenFile(outfileName, append)
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = fmt.Fprintf(f, "---\n# Source: %s\n%s\n", name, data)
	if err != nil {
		return err
	}

	fmt.Printf("wrote %s\n", outfileName)
	return nil
}
func createOrOpenFile(filename string, append bool) (*os.File, error) {
	if append {
		return os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	}
	return os.Create(filename)
}

// check if the directory exists to create file. creates if don't exists
func ensureDirectoryForFile(file string) error {
	baseDir := path.Dir(file)
	_, err := os.Stat(baseDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(baseDir, defaultDirectoryPermission)
}
