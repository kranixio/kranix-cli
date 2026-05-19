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
	server     string
	apiKey     string
	httpClient *http.Client
}

type CronScheduleSpec struct {
	Schedule          string `json:"schedule,omitempty"`
	Suspended         bool   `json:"suspended,omitempty"`
	TimeZone          string `json:"timeZone,omitempty"`
	ConcurrencyPolicy string `json:"concurrencyPolicy,omitempty"`
}

type WorkloadTags struct {
	Team        string `json:"team,omitempty"`
	Environment string `json:"environment,omitempty"`
	CostCenter  string `json:"costCenter,omitempty"`
}

type WorkloadSpec struct {
	Name         string              `json:"name"`
	Image        string              `json:"image"`
	Namespace    string              `json:"namespace"`
	Replicas     int                 `json:"replicas"`
	Env          map[string]string   `json:"env"`
	Port         int                 `json:"port"`
	CPU          string              `json:"cpu"`
	Memory       string              `json:"memory"`
	Command      string              `json:"command"`
	CronSchedule *CronScheduleSpec   `json:"cronSchedule,omitempty"`
	Tags         *WorkloadTags       `json:"tags,omitempty"`
}

type WorkloadStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	Image     string `json:"image"`
	Namespace string `json:"namespace"`
}

type Pod struct {
	Name   string `json:"name"`
	Ready  string `json:"ready"`
	Status string `json:"status"`
	Age    string `json:"age"`
	Node   string `json:"node"`
}

type LogOptions struct {
	TailLines int    `json:"tail_lines"`
	Follow    bool   `json:"follow"`
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

type DiffResult struct {
	WorkloadName string                 `json:"workload_name"`
	Changes      []DiffChange           `json:"changes"`
	Summary      map[string]interface{} `json:"summary"`
}

type DiffChange struct {
	Field      string      `json:"field"`
	OldValue   interface{} `json:"old_value"`
	NewValue   interface{} `json:"new_value"`
	ChangeType string      `json:"change_type"` // added, modified, removed
}

type DryRunPreview struct {
	Actions []DryRunAction         `json:"actions"`
	Summary map[string]interface{} `json:"summary"`
}

type DryRunAction struct {
	Type        string                 `json:"type"`
	Resource    string                 `json:"resource"`
	Namespace   string                 `json:"namespace"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
}

type AIRequest struct {
	Prompt    string `json:"prompt"`
	Namespace string `json:"namespace"`
	Workload  string `json:"workload,omitempty"`
}

type AIResponse struct {
	Response        string  `json:"response"`
	SuggestedAction string  `json:"suggested_action,omitempty"`
	CodeSnippet     string  `json:"code_snippet,omitempty"`
	Confidence      float64 `json:"confidence"`
}

type WorkloadCost struct {
	WorkloadName string          `json:"workload_name"`
	Namespace    string          `json:"namespace"`
	Duration     string          `json:"duration"`
	TotalCost    float64         `json:"total_cost"`
	ComputeCost  float64         `json:"compute_cost"`
	StorageCost  float64         `json:"storage_cost"`
	NetworkCost  float64         `json:"network_cost"`
	Breakdown    []CostBreakdown `json:"breakdown"`
}

type CostBreakdown struct {
	Resource string  `json:"resource"`
	Cost     float64 `json:"cost"`
	Usage    string  `json:"usage"`
}

type CostSummary struct {
	Namespace        string         `json:"namespace"`
	Duration         string         `json:"duration"`
	TotalCost        float64        `json:"total_cost"`
	WorkloadCount    int            `json:"workload_count"`
	AverageCost      float64        `json:"average_cost"`
	TopCostWorkloads []WorkloadCost `json:"top_cost_workloads"`
}

type Template struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Category    string        `json:"category"`
	Variables   []TemplateVar `json:"variables"`
}

type TemplateVar struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
}

type TemplateContent struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (c *Client) GetDiff(ctx context.Context, name, namespace string, proposedSpec *WorkloadSpec) (*DiffResult, error) {
	body, err := json.Marshal(proposedSpec)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v1/workloads/%s/diff?namespace=%s", name, namespace)
	resp, err := c.doRequest(ctx, "POST", path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var diff DiffResult
	if err := json.NewDecoder(resp.Body).Decode(&diff); err != nil {
		return nil, err
	}

	return &diff, nil
}

func (c *Client) GetDryRunPreview(ctx context.Context, namespace, manifest string) (*DryRunPreview, error) {
	body := []byte(fmt.Sprintf(`{"namespace": "%s", "manifest": %q}`, namespace, manifest))
	resp, err := c.doRequest(ctx, "POST", "/api/v1/dryrun/preview", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var preview DryRunPreview
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		return nil, err
	}

	return &preview, nil
}

func (c *Client) AskAI(ctx context.Context, namespace, prompt string) (*AIResponse, error) {
	req := &AIRequest{
		Prompt:    prompt,
		Namespace: namespace,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/ai/ask", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var aiResp AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, err
	}

	return &aiResp, nil
}

func (c *Client) GetWorkloadCost(ctx context.Context, workloadName, namespace, duration string) (*WorkloadCost, error) {
	path := fmt.Sprintf("/api/v1/workloads/%s/cost?namespace=%s&duration=%s", workloadName, namespace, duration)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var cost WorkloadCost
	if err := json.NewDecoder(resp.Body).Decode(&cost); err != nil {
		return nil, err
	}

	return &cost, nil
}

func (c *Client) GetCostSummary(ctx context.Context, namespace, duration string) (*CostSummary, error) {
	path := fmt.Sprintf("/api/v1/cost/summary?namespace=%s&duration=%s", namespace, duration)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var summary CostSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

func (c *Client) ListTemplates(ctx context.Context) ([]*Template, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/templates", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var templates []*Template
	if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
		return nil, err
	}

	return templates, nil
}

func (c *Client) GetTemplate(ctx context.Context, templateName string, vars map[string]string) (*TemplateContent, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"name": templateName,
		"vars": vars,
	})
	resp, err := c.doRequest(ctx, "POST", "/api/v1/templates/get", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var content TemplateContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, err
	}

	return &content, nil
}

func (c *Client) ListPods(ctx context.Context, workloadName, namespace string) ([]*Pod, error) {
	path := fmt.Sprintf("/api/v1/workloads/%s/pods?namespace=%s", workloadName, namespace)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var pods []*Pod
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		return nil, err
	}

	return pods, nil
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
