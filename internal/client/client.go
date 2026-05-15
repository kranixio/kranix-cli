package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kranix-io/kranix-cli/internal/auth"
)

type Client struct {
	server    string
	apiKey    string
	httpClient *http.Client
}

type WorkloadSpec struct {
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Namespace string            `json:"namespace"`
	Replicas  int               `json:"replicas"`
	Env       map[string]string `json:"env"`
	Port      int               `json:"port"`
	CPU       string            `json:"cpu"`
	Memory    string            `json:"memory"`
	Command   string            `json:"command"`
}

type WorkloadStatus struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
	Image string `json:"image"`
}

type LogOptions struct {
	TailLines int  `json:"tail_lines"`
	Follow    bool `json:"follow"`
	Since     string `json:"since"`
}

type Namespace struct {
	Name string `json:"name"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details"`
}

func New(server, apiKey string) *Client {
	return &Client{
		server: server,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.server+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", auth.GetAuthHeader(c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *Client) Deploy(ctx context.Context, spec *WorkloadSpec) (*WorkloadStatus, error) {
	body, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/workloads", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var status WorkloadStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (c *Client) GetStatus(ctx context.Context, name, namespace string) (*WorkloadStatus, error) {
	path := fmt.Sprintf("/api/v1/workloads/%s?namespace=%s", name, namespace)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var status WorkloadStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (c *Client) ListWorkloads(ctx context.Context, namespace string) ([]*WorkloadStatus, error) {
	path := fmt.Sprintf("/api/v1/workloads?namespace=%s", namespace)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var statuses []*WorkloadStatus
	if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
		return nil, err
	}

	return statuses, nil
}

func (c *Client) Restart(ctx context.Context, name, namespace string) error {
	path := fmt.Sprintf("/api/v1/workloads/%s/restart?namespace=%s", name, namespace)
	resp, err := c.doRequest(ctx, "POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, name, namespace string) error {
	path := fmt.Sprintf("/api/v1/workloads/%s?namespace=%s", name, namespace)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

func (c *Client) StreamLogs(ctx context.Context, name, namespace string, opts *LogOptions) (io.ReadCloser, error) {
	path := fmt.Sprintf("/api/v1/workloads/%s/logs?namespace=%s&tail=%d&follow=%t", name, namespace, opts.TailLines, opts.Follow)
	if opts.Since != "" {
		path += fmt.Sprintf("&since=%s", opts.Since)
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, c.parseError(resp)
	}

	return resp.Body, nil
}

func (c *Client) CreateNamespace(ctx context.Context, name string) error {
	body, _ := json.Marshal(map[string]string{"name": name})
	resp, err := c.doRequest(ctx, "POST", "/api/v1/namespaces", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return c.parseError(resp)
	}

	return nil
}

func (c *Client) ListNamespaces(ctx context.Context) ([]*Namespace, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/namespaces", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var namespaces []*Namespace
	if err := json.NewDecoder(resp.Body).Decode(&namespaces); err != nil {
		return nil, err
	}

	return namespaces, nil
}

func (c *Client) DeleteNamespace(ctx context.Context, name string) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s", name)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.parseError(resp)
	}

	return nil
}

func (c *Client) parseError(resp *http.Response) error {
	var errResp ErrorResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &errResp)

	if errResp.Error != "" {
		return fmt.Errorf("%s: %s", errResp.Code, errResp.Error)
	}

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}
