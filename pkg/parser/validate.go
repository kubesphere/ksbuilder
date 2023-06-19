package parser

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"path"

	"sigs.k8s.io/yaml"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

func ValidateExtension(name string, zipFile []byte) error {
	gr, err := gzip.NewReader(bytes.NewReader(zipFile))
	if err != nil {
		return fmt.Errorf("gzip read file failed: %s", err.Error())
	}
	defer gr.Close() // nolint

	metadataFilename := path.Join(name, extension.MetadataFilename)
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if h.FileInfo().IsDir() {
			continue
		}
		if h.Name != metadataFilename {
			continue
		}
		buffer := make([]byte, h.Size)
		if _, err = tr.Read(buffer); err != nil && err != io.EOF {
			return fmt.Errorf("read tar file failed: %s", err.Error())
		}
		if err = yaml.Unmarshal(buffer, &extension.Metadata{}); err != nil {
			return fmt.Errorf("validate the extension metadata failed: %s", err.Error())
		}
		return nil
	}
	return errors.New("unable to find the extension metadata file")
}
