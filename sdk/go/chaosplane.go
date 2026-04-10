// Package chaosplane provides a Go client for the ChaosPlane Platform API.
package chaosplane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const defaultBaseURL = "http://localhost:8080"

// APIError represents an error response from the ChaosPlane API.
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("chaosplane: %d %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("chaosplane: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type AuthUser struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	EmailVerified bool       `json:"emailVerified"`
	LastLoginAt   *time.Time `json:"lastLoginAt,omitempty"`
}

type AuthTenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type AuthResponse struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	ExpiresIn    int64      `json:"expiresIn"`
	User         AuthUser   `json:"user"`
	Tenant       AuthTenant `json:"tenant"`
}

type ActionRequest struct {
	Type       string          `json:"type"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}

type TargetRequest struct {
	Kind          string            `json:"kind"`
	Namespace     string            `json:"namespace,omitempty"`
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
	Names         []string          `json:"names,omitempty"`
}

type CreateExperimentRequest struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Action    ActionRequest `json:"action"`
	Target    TargetRequest `json:"target"`
	Duration  string        `json:"duration"`
}

type ExperimentResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Action    string `json:"action"`
	Phase     string `json:"phase"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

type PaginatedExperiments struct {
	Items      []ExperimentResponse `json:"items"`
	TotalCount int                  `json:"totalCount"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
}

type Organization struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

type Workspace struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenantId"`
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Description    string `json:"description,omitempty"`
}

type Team struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

type Project struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

type Environment struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	ProjectID   string `json:"projectId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	AgentStatus string `json:"agentStatus"`
}

type HierarchyResponse struct {
	Organizations []Organization `json:"organizations"`
	Workspaces    []Workspace    `json:"workspaces"`
	Teams         []Team         `json:"teams"`
	Projects      []Project      `json:"projects"`
	Environments  []Environment  `json:"environments"`
}

type Subscription struct {
	ID                 string     `json:"id"`
	TenantID           string     `json:"tenantId"`
	Plan               string     `json:"plan"`
	Status             string     `json:"status"`
	Gateway            string     `json:"gateway"`
	CurrentPeriodStart *time.Time `json:"currentPeriodStart,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"currentPeriodEnd,omitempty"`
	TrialEndsAt        *time.Time `json:"trialEndsAt,omitempty"`
	CancelledAt        *time.Time `json:"cancelledAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type UsageStats struct {
	Experiments int64 `json:"experiments"`
	Agents      int64 `json:"agents"`
	APICalls    int64 `json:"apiCalls"`
}

type PlanLimits struct {
	MaxExperiments int64 `json:"maxExperiments"`
	MaxAgents      int64 `json:"maxAgents"`
	MaxAPICalls    int64 `json:"maxApiCalls"`
}

type BillingStatusResponse struct {
	Subscription *Subscription `json:"subscription"`
	Usage        *UsageStats   `json:"usage"`
	Limits       *PlanLimits   `json:"limits"`
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets a custom API base URL.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// WithAPIKey sets the X-API-Key header for API-key authenticated endpoints.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

// Client is the ChaosPlane Platform API client.
type Client struct {
	baseURL      string
	apiKey       string
	accessToken  string
	refreshToken string
	http         *http.Client
	mu           sync.RWMutex
}

// New creates a new ChaosPlane client.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// SetTokens manually sets the Bearer access and refresh tokens.
func (c *Client) SetTokens(access, refresh string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = access
	c.refreshToken = refresh
}

// Login authenticates with email/password and stores the returned tokens.
func (c *Client) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	var resp AuthResponse
	if err := c.do(ctx, http.MethodPost, "/auth/login", req, &resp); err != nil {
		return nil, err
	}
	c.SetTokens(resp.AccessToken, resp.RefreshToken)
	return &resp, nil
}

// Refresh exchanges a refresh token for new tokens.
func (c *Client) Refresh(ctx context.Context) (*AuthResponse, error) {
	c.mu.RLock()
	rt := c.refreshToken
	c.mu.RUnlock()

	var resp AuthResponse
	if err := c.do(ctx, http.MethodPost, "/auth/refresh", &RefreshRequest{RefreshToken: rt}, &resp); err != nil {
		return nil, err
	}
	c.SetTokens(resp.AccessToken, resp.RefreshToken)
	return &resp, nil
}

// ListExperiments returns a paginated list of experiments.
func (c *Client) ListExperiments(ctx context.Context, limit, offset int) (*PaginatedExperiments, error) {
	path := fmt.Sprintf("/api/v1/experiments?limit=%s&offset=%s",
		url.QueryEscape(strconv.Itoa(limit)),
		url.QueryEscape(strconv.Itoa(offset)),
	)
	var resp PaginatedExperiments
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateExperiment creates a new chaos experiment.
func (c *Client) CreateExperiment(ctx context.Context, req *CreateExperimentRequest) (*ExperimentResponse, error) {
	var resp ExperimentResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/experiments", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetExperiment retrieves an experiment by name.
func (c *Client) GetExperiment(ctx context.Context, name string) (*ExperimentResponse, error) {
	var resp ExperimentResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/experiments/"+url.PathEscape(name), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteExperiment deletes an experiment by name.
func (c *Client) DeleteExperiment(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/experiments/"+url.PathEscape(name), nil, nil)
}

// AbortExperiment aborts a running experiment by name.
func (c *Client) AbortExperiment(ctx context.Context, name string) (*ExperimentResponse, error) {
	var resp ExperimentResponse
	if err := c.do(ctx, http.MethodPost, "/api/v1/experiments/"+url.PathEscape(name)+"/abort", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListHierarchy returns the full tenant hierarchy.
func (c *Client) ListHierarchy(ctx context.Context) (*HierarchyResponse, error) {
	var resp HierarchyResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/hierarchy", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBillingStatus returns the current billing status.
func (c *Client) GetBillingStatus(ctx context.Context) (*BillingStatusResponse, error) {
	var resp BillingStatusResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/billing", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) do(ctx context.Context, method, path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("chaosplane: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("chaosplane: build request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	} else {
		c.mu.RLock()
		token := c.accessToken
		c.mu.RUnlock()
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("chaosplane: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("chaosplane: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: respBody}
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			apiErr.Message = errResp.Error
		}
		return apiErr
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("chaosplane: decode response: %w", err)
		}
	}

	return nil
}
