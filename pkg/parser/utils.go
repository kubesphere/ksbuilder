package parser

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path"
	"path/filepath"
)

func unzip(zipFile []byte, dest string) error {
	gr, err := gzip.NewReader(bytes.NewReader(zipFile))
	if err != nil {
		return err
	}
	defer gr.Close() // nolint

	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		p := path.Join(dest, h.Name)
		if h.FileInfo().IsDir() {
			if err = os.MkdirAll(p, os.ModePerm); err != nil {
				return err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
			return err
		}
		fw, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, h.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer fw.Close() // nolint

		if _, err = io.Copy(fw, tr); err != nil {
			return err
		}
	}
	return nil
}

func readFile(dir, name string) ([]byte, error) {
	data, err := os.ReadFile(path.Join(dir, name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}
