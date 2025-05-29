package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/api"
	"github.com/kubesphere/ksbuilder/pkg/cloud"
	"github.com/kubesphere/ksbuilder/pkg/extension"
)

type pushOptions struct{}

func pushCmd() *cobra.Command {
	o := pushOptions{}

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push and submit extensions to KubeSphere Cloud for review",
		Long: `NOTE: We will upload static files such as icons and screenshots in the extension to the KubeSphere Cloud separately
and delete the static file directory in the original package to reduce the size of the entire chart`,
		Args: cobra.ExactArgs(1),
		RunE: o.push,
	}
	return cmd
}

func (o *pushOptions) push(_ *cobra.Command, args []string) error {
	fmt.Printf("push extension %s\n", args[0])

	client, err := cloud.NewClient()
	if err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return err
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Printf("remove %s failed: %v", path, err)
		}
	}(tempDir)
	if err = extension.WriteFilesToTempDir(args[0], tempDir); err != nil {
		return err
	}

	metadata, err := api.LoadMetadata(tempDir, api.WithEncodeIcon(false))
	if err != nil {
		return err
	}
	// upload images to cloud
	if api.IsLocalFile(metadata.Icon) {
		resp, err := client.UploadFiles(metadata.Name, metadata.Version, tempDir, metadata.Icon)
		if err != nil {
			return err
		}
		metadata.Icon = resp.Files[0].URL
	}
	screenshots := make([]string, 0)
	localScreenshots := make([]string, 0)
	for _, p := range metadata.Screenshots {
		if api.IsLocalFile(p) {
			localScreenshots = append(localScreenshots, p)
		} else {
			screenshots = append(screenshots, p)
		}
	}
	if len(localScreenshots) > 0 {
		resp, err := client.UploadFiles(metadata.Name, metadata.Version, tempDir, localScreenshots...)
		if err != nil {
			return err
		}
		for _, f := range resp.Files {
			screenshots = append(screenshots, f.URL)
		}
		metadata.Screenshots = screenshots
	}
	// delete static directory to reduce chart package size
	if metadata.StaticFileDirectory != "" {
		if err = os.RemoveAll(filepath.Join(tempDir, metadata.StaticFileDirectory)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	data, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}
	if err = os.WriteFile(tempDir+"/"+api.MetadataFilename, data, 0644); err != nil {
		return err
	}

	chartMetadata, err := yaml.Marshal(metadata.ToChartYaml())
	if err != nil {
		return err
	}
	if err = os.WriteFile(tempDir+"/Chart.yaml", chartMetadata, 0644); err != nil {
		return err
	}

	ch, err := loader.LoadDir(tempDir)
	if err != nil {
		return err
	}
	chartFilename, err := chartutil.Save(ch, tempDir)
	if err != nil {
		return err
	}

	uploadExtensionResp, err := client.UploadExtension(metadata.Name, chartFilename)
	if err != nil {
		return err
	}
	if err = client.SubmitExtension(uploadExtensionResp.Snapshot.SnapshotID); err != nil {
		return err
	}
	fmt.Println("Extension pushed and submitted to KubeSphere Cloud, waiting for review")
	return nil
}
