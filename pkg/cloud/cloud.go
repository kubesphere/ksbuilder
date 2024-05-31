package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kubesphere/ksbuilder/pkg/config"
)

type Options struct {
	server string
	token  string
}

func WithServer(server string) func(opts *Options) {
	return func(opts *Options) {
		opts.server = server
	}
}

func WithToken(token string) func(opts *Options) {
	return func(opts *Options) {
		opts.token = token
	}
}

type Client struct {
	client *http.Client
	server string
	token  string
	userID string
}

func NewClient(options ...func(*Options)) (*Client, error) {
	c, err := config.Read()
	if err != nil {
		return nil, err
	}
	opts := &Options{
		token:  string(c.Token),
		server: c.Server,
	}
	for _, f := range options {
		f(opts)
	}
	if opts.server == "" {
		opts.server = "https://apis.kubesphere.cloud"
	}

	client := &Client{
		client: http.DefaultClient,
		server: opts.server,
		token:  opts.token,
	}

	data := &userInfo{}
	if err = client.sendRequest(http.MethodGet, "/apis/user/v1/user", nil, nil, data); err != nil {
		return nil, err
	}
	client.userID = data.UserID
	return client, nil
}

func (c *Client) sendRequest(method, path string, body io.Reader, headers map[string]string, respData interface{}) error {
	req, err := http.NewRequest(method, c.server+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		data := errorResponse{}
		if err = json.Unmarshal(responseBody, &data); err != nil {
			return err
		}
		return fmt.Errorf("%s, %s", resp.Status, data.Message)
	}

	if respData == nil {
		return nil
	}
	return json.Unmarshal(responseBody, respData)
}

func (c *Client) UploadFiles(extensionName, extensionVersion, sourceDir string, paths ...string) (*UploadFilesResponse, error) {
	if len(paths) == 0 {
		return &UploadFilesResponse{}, nil
	}

	fileUsage := &fileUsageResponse{}
	if err := c.sendRequest(
		http.MethodGet, fmt.Sprintf("/apis/extension/v1/users/%s/files/usage", c.userID), nil, nil, fileUsage,
	); err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for i, path := range paths {
		if err := func() error {
			f, err := os.Open(filepath.Join(sourceDir, path))
			if err != nil {
				return err
			}
			defer f.Close()

			fileName := filepath.Base(f.Name())
			part, err := writer.CreateFormFile(fmt.Sprintf("file%d", i+1), fileName)
			if err != nil {
				return err
			}
			if _, err = io.Copy(part, f); err != nil {
				return err
			}
			if err = writer.WriteField(
				fmt.Sprintf("file%d_path", i+1),
				filepath.Join(fileUsage.Dir, extensionName, extensionVersion, fileName),
			); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	data := &UploadFilesResponse{}
	if err := c.sendRequest(
		http.MethodPost,
		fmt.Sprintf("/apis/extension/v1/users/%s/files", c.userID),
		body,
		map[string]string{"Content-Type": writer.FormDataContentType()},
		data,
	); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Client) UploadExtension(extensionName, path string) (*UploadExtensionResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileName := filepath.Base(f.Name())
	part, err := writer.CreateFormFile("extension_package", fileName)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(part, f); err != nil {
		return nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}

	data := &UploadExtensionResponse{}
	if err = c.sendRequest(
		http.MethodPost,
		fmt.Sprintf("/apis/extension/v1/users/%s/extensions/%s/package?force=true&create_extension=true", c.userID, extensionName),
		body,
		map[string]string{"Content-Type": writer.FormDataContentType()},
		data,
	); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Client) SubmitExtension(snapshotID string) error {
	body := &bytes.Buffer{}
	body.WriteString(fmt.Sprintf(`{"message": "%s submit for review"}`, time.Now().Format(time.RFC3339)))

	return c.sendRequest(
		http.MethodPost,
		fmt.Sprintf("/apis/extension/v1/users/%s/snapshots/%s/action:submit", c.userID, snapshotID),
		body,
		map[string]string{"Content-Type": "application/json"},
		nil,
	)
}
