// api/rest.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RESTClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewRESTClient(token, baseURL string) *RESTClient {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &RESTClient{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *RESTClient) get(path string, out any) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid token — run 'ghboard' and enter a valid token")
	}
	if resp.StatusCode == 403 {
		return fmt.Errorf("rate limited (resets: %s)", resp.Header.Get("X-RateLimit-Reset"))
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// --- Stars ---

type StarredRepo struct {
	FullName        string `json:"full_name"`
	Description     string `json:"description"`
	Language        string `json:"language"`
	StargazersCount int    `json:"stargazers_count"`
	HTMLURL         string `json:"html_url"`
	UpdatedAt       string `json:"updated_at"`
	Topics          []string `json:"topics"`
}

func (c *RESTClient) ListStars(page int) ([]StarredRepo, error) {
	var repos []StarredRepo
	path := fmt.Sprintf("/user/starred?per_page=100&page=%d", page)
	return repos, c.get(path, &repos)
}

func (c *RESTClient) Unstar(owner, repo string) error {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/user/starred/%s/%s", c.baseURL, owner, repo), nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return fmt.Errorf("unstar failed: %s", resp.Status)
	}
	return nil
}

// --- Notifications ---

type Subject struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

type NotifRepo struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

type Notification struct {
	ID         string    `json:"id"`
	Unread     bool      `json:"unread"`
	Reason     string    `json:"reason"`
	UpdatedAt  string    `json:"updated_at"`
	Subject    Subject   `json:"subject"`
	Repository NotifRepo `json:"repository"`
}

func (c *RESTClient) ListNotifications() ([]Notification, error) {
	var notifs []Notification
	return notifs, c.get("/notifications?all=false&per_page=100", &notifs)
}

func (c *RESTClient) MarkRead(id string) error {
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/notifications/threads/%s", c.baseURL, id), nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *RESTClient) MarkAllRead() error {
	req, _ := http.NewRequest("PUT", c.baseURL+"/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
