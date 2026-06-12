// api/gitlab.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GitLabClient handles both REST and GraphQL calls to GitLab.
type GitLabClient struct {
	token   string
	baseURL string
	host    string
	http    *http.Client
}

func NewGitLabClient(token, host string) *GitLabClient {
	if host == "" {
		host = "gitlab.com"
	}
	baseURL := "https://" + host
	return &GitLabClient{
		token:   token,
		baseURL: baseURL,
		host:    host,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *GitLabClient) Host() string { return c.host }

func (c *GitLabClient) restGet(path string, out any) error {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v4"+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid GitLab token — update ~/.config/ghboard/config.json")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("GitLab API error: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *GitLabClient) graphql(query string, out any) error {
	body, _ := json.Marshal(map[string]string{"query": query})
	req, err := http.NewRequest("POST", c.baseURL+"/api/graphql", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("GitLab GraphQL error: %s", resp.Status)
	}
	var wrapped struct {
		Data   any `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	wrapped.Data = out
	if err := json.NewDecoder(resp.Body).Decode(&wrapped); err != nil {
		return err
	}
	if len(wrapped.Errors) > 0 {
		return fmt.Errorf("graphql error: %s", wrapped.Errors[0].Message)
	}
	return nil
}

// --- Stars (GitLab: starred projects) ---

func (c *GitLabClient) ListStars(page int) ([]StarredRepo, error) {
	var raw []struct {
		ID                int    `json:"id"`
		PathWithNamespace string `json:"path_with_namespace"`
		Description       string `json:"description"`
		StarCount         int    `json:"star_count"`
		WebURL            string `json:"web_url"`
		LastActivityAt    string `json:"last_activity_at"`
		Topics            []string `json:"topics"`
	}
	path := fmt.Sprintf("/projects?starred=true&per_page=100&page=%d", page)
	if err := c.restGet(path, &raw); err != nil {
		return nil, err
	}
	out := make([]StarredRepo, len(raw))
	for i, r := range raw {
		out[i] = StarredRepo{
			FullName:        r.PathWithNamespace,
			Description:     r.Description,
			StargazersCount: r.StarCount,
			HTMLURL:         r.WebURL,
			UpdatedAt:       r.LastActivityAt,
			Topics:          r.Topics,
		}
	}
	return out, nil
}

func (c *GitLabClient) Unstar(projectID string) error {
	req, _ := http.NewRequest("DELETE",
		fmt.Sprintf("%s/api/v4/projects/%s/star", c.baseURL, projectID), nil)
	req.Header.Set("PRIVATE-TOKEN", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 304 {
		return fmt.Errorf("unstar failed: %s", resp.Status)
	}
	return nil
}

// --- Notifications (GitLab: todos) ---

func (c *GitLabClient) ListTodos() ([]Notification, error) {
	var raw []struct {
		ID        int    `json:"id"`
		State     string `json:"state"`
		TargetType string `json:"target_type"`
		Target    struct {
			Title  string `json:"title"`
			WebURL string `json:"web_url"`
		} `json:"target"`
		Project struct {
			PathWithNamespace string `json:"path_with_namespace"`
			WebURL            string `json:"web_url"`
		} `json:"project"`
		ActionName string `json:"action_name"`
		CreatedAt  string `json:"created_at"`
	}
	if err := c.restGet("/todos?state=pending&per_page=100", &raw); err != nil {
		return nil, err
	}
	out := make([]Notification, len(raw))
	for i, t := range raw {
		subjType := t.TargetType
		switch subjType {
		case "MergeRequest":
			subjType = "PullRequest"
		case "Issue":
			subjType = "Issue"
		default:
			subjType = "CheckSuite"
		}
		out[i] = Notification{
			ID:         fmt.Sprintf("gl-%d", t.ID),
			Unread:     t.State == "pending",
			Reason:     t.ActionName,
			UpdatedAt:  t.CreatedAt,
			Subject: Subject{
				Title: t.Target.Title,
				Type:  subjType,
				URL:   t.Target.WebURL,
			},
			Repository: NotifRepo{
				FullName: t.Project.PathWithNamespace,
				HTMLURL:  t.Project.WebURL,
			},
		}
	}
	return out, nil
}

func (c *GitLabClient) MarkTodoDone(id string) error {
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/api/v4/todos/%s/mark_as_done", c.baseURL, id[3:]), nil)
	req.Header.Set("PRIVATE-TOKEN", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *GitLabClient) MarkAllTodosDone() error {
	req, _ := http.NewRequest("POST",
		c.baseURL+"/api/v4/todos/mark_as_done", nil)
	req.Header.Set("PRIVATE-TOKEN", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// --- Contributions (GitLab GraphQL) ---

func (c *GitLabClient) FetchContributions(year int) (*ContributionData, error) {
	from := fmt.Sprintf("%d-01-01", year)
	to := fmt.Sprintf("%d-12-31", year)

	query := fmt.Sprintf(`{
		currentUser {
			username
			contributions(from: "%s", to: "%s") {
				totalCommitContributions
			}
		}
	}`, from, to)

	var result struct {
		CurrentUser struct {
			Username     string `json:"username"`
			Contributions struct {
				TotalCommitContributions int `json:"totalCommitContributions"`
			} `json:"contributions"`
		} `json:"currentUser"`
	}
	if err := c.graphql(query, &result); err != nil {
		return nil, err
	}
	return &ContributionData{
		Login:              result.CurrentUser.Username,
		TotalContributions: result.CurrentUser.Contributions.TotalCommitContributions,
		Weeks:              nil,
	}, nil
}

// --- ID helper ---
func (c *GitLabClient) ProviderName() string { return "gitlab" }
