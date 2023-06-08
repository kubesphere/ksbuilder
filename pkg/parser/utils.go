package parser

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path"
)

func unzip(zipFile []byte) (map[string][]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(zipFile))
	if err != nil {
		return nil, err
	}
	defer gr.Close() // nolint

	data := make(map[string][]byte)
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if h.FileInfo().IsDir() {
			continue
		}

		buffer := make([]byte, h.Size)
		if _, err = tr.Read(buffer); err != nil && err != io.EOF {
			return nil, err
		}
		data[h.Name] = buffer
	}
	return data, nil
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
