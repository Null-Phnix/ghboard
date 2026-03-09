// api/graphql.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GraphQLClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewGraphQLClient(token, baseURL string) *GraphQLClient {
	if baseURL == "" {
		baseURL = "https://api.github.com/graphql"
	}
	return &GraphQLClient{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

type ContributionDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
	Weekday           int    `json:"weekday"`
}

type ContributionWeek struct {
	ContributionDays []ContributionDay `json:"contributionDays"`
}

type ContributionData struct {
	Login              string
	TotalContributions int
	Weeks              []ContributionWeek
}

func (c *GraphQLClient) FetchContributions(year int) (*ContributionData, error) {
	from := fmt.Sprintf("%d-01-01T00:00:00Z", year)
	to := fmt.Sprintf("%d-12-31T23:59:59Z", year)

	query := fmt.Sprintf(`{
		user: viewer {
			login
			contributionsCollection(from: "%s", to: "%s") {
				totalCommitContributions
				contributionCalendar {
					totalContributions
					weeks {
						contributionDays {
							date
							contributionCount
							weekday
						}
					}
				}
			}
		}
	}`, from, to)

	body, _ := json.Marshal(map[string]string{"query": query})
	req, err := http.NewRequest("POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			User struct {
				Login                   string `json:"login"`
				ContributionsCollection struct {
					ContributionCalendar struct {
						TotalContributions int                `json:"totalContributions"`
						Weeks              []ContributionWeek `json:"weeks"`
					} `json:"contributionCalendar"`
				} `json:"contributionsCollection"`
			} `json:"user"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	u := result.Data.User
	return &ContributionData{
		Login:              u.Login,
		TotalContributions: u.ContributionsCollection.ContributionCalendar.TotalContributions,
		Weeks:              u.ContributionsCollection.ContributionCalendar.Weeks,
	}, nil
}
