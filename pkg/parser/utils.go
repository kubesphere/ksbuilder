package parser

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
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
