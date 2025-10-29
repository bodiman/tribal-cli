package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

type Config struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

type Graph struct {
	ID          uuid.UUID              `json:"id"`
	Title       string                 `json:"title"`
	Description *string                `json:"description"`
	Nodes       []Node                 `json:"nodes"`
	Edges       []Edge                 `json:"edges"`
	Metadata    map[string]interface{} `json:"metadata"`
	OwnerID     uuid.UUID              `json:"owner_id"`
	Owner       User                   `json:"owner"`
	IsPublic    bool                   `json:"is_public"`
	Version     int                    `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type Node struct {
	ID       string     `json:"id"`
	Label    string     `json:"label"`
	Markup   *string    `json:"markup,omitempty"`
	Position Position   `json:"position"`
	Size     *Size      `json:"size,omitempty"`
}

type Edge struct {
	ID       string  `json:"id"`
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Directed bool    `json:"directed"`
	Label    *string `json:"label,omitempty"`
	Markup   *string `json:"markup,omitempty"`
	Size     *Size   `json:"size,omitempty"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type CreateGraphRequest struct {
	Title       string                 `json:"title"`
	Description *string                `json:"description"`
	Nodes       []Node                 `json:"nodes"`
	Edges       []Edge                 `json:"edges"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsPublic    bool                   `json:"is_public"`
}

type UpdateGraphRequest struct {
	Title       *string                `json:"title,omitempty"`
	Description *string                `json:"description,omitempty"`
	Nodes       *[]Node                `json:"nodes,omitempty"`
	Edges       *[]Edge                `json:"edges,omitempty"`
	Metadata    *map[string]interface{} `json:"metadata,omitempty"`
	IsPublic    *bool                  `json:"is_public,omitempty"`
	Message     string                 `json:"message"`
}

type SearchRequest struct {
	Query     string     `json:"query"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	IsPublic  *bool      `json:"is_public,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

type SearchResponse struct {
	Graphs  []Graph `json:"graphs"`
	Total   int     `json:"total"`
	HasMore bool    `json:"has_more"`
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, errResp.Error, errResp.Message)
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Auth methods
func (c *Client) Register(username, email, password string) (*AuthResponse, error) {
	body := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}

	resp, err := c.doRequest("POST", "/api/v1/auth/register", body)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := c.handleResponse(resp, &authResp); err != nil {
		return nil, err
	}

	return &authResp, nil
}

func (c *Client) Login(username, password string) (*AuthResponse, error) {
	body := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := c.doRequest("POST", "/api/v1/auth/login", body)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := c.handleResponse(resp, &authResp); err != nil {
		return nil, err
	}

	return &authResp, nil
}

func (c *Client) GetMe() (*User, error) {
	resp, err := c.doRequest("GET", "/api/v1/auth/me", nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := c.handleResponse(resp, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Graph methods
func (c *Client) CreateGraph(req CreateGraphRequest) (*Graph, error) {
	resp, err := c.doRequest("POST", "/api/v1/graphs", req)
	if err != nil {
		return nil, err
	}

	var graph Graph
	if err := c.handleResponse(resp, &graph); err != nil {
		return nil, err
	}

	return &graph, nil
}

func (c *Client) GetGraphs(limit, offset int) ([]Graph, int, bool, error) {
	path := fmt.Sprintf("/api/v1/graphs?limit=%d&offset=%d", limit, offset)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, 0, false, err
	}

	var result struct {
		Graphs  []Graph `json:"graphs"`
		Total   int     `json:"total"`
		HasMore bool    `json:"has_more"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, 0, false, err
	}

	return result.Graphs, result.Total, result.HasMore, nil
}

func (c *Client) GetGraph(id string) (*Graph, error) {
	resp, err := c.doRequest("GET", "/api/v1/graphs/"+id, nil)
	if err != nil {
		return nil, err
	}

	var graph Graph
	if err := c.handleResponse(resp, &graph); err != nil {
		return nil, err
	}

	return &graph, nil
}

func (c *Client) UpdateGraph(id string, req UpdateGraphRequest) (*Graph, error) {
	resp, err := c.doRequest("PUT", "/api/v1/graphs/"+id, req)
	if err != nil {
		return nil, err
	}

	var graph Graph
	if err := c.handleResponse(resp, &graph); err != nil {
		return nil, err
	}

	return &graph, nil
}

func (c *Client) DeleteGraph(id string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/graphs/"+id, nil)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}

func (c *Client) SearchGraphs(req SearchRequest) (*SearchResponse, error) {
	resp, err := c.doRequest("POST", "/api/v1/search/graphs", req)
	if err != nil {
		return nil, err
	}

	var searchResp SearchResponse
	if err := c.handleResponse(resp, &searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}

func (c *Client) GetPublicGraphs(limit, offset int) ([]Graph, int, bool, error) {
	path := fmt.Sprintf("/api/v1/public/graphs?limit=%d&offset=%d", limit, offset)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, 0, false, err
	}

	var result struct {
		Graphs  []Graph `json:"graphs"`
		Total   int     `json:"total"`
		HasMore bool    `json:"has_more"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, 0, false, err
	}

	return result.Graphs, result.Total, result.HasMore, nil
}

// Health check
func (c *Client) HealthCheck() error {
	resp, err := c.doRequest("GET", "/health", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// URL validation
func (c *Client) ValidateURL() error {
	parsedURL, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("base URL must use http or https scheme")
	}

	return nil
}