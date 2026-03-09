// api/graphql_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchContributions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": map[string]any{
				"user": map[string]any{
					"login": "Null-Phnix",
					"contributionsCollection": map[string]any{
						"totalCommitContributions": 150,
						"contributionCalendar": map[string]any{
							"totalContributions": 200,
							"weeks": []map[string]any{
								{
									"contributionDays": []map[string]any{
										{"date": "2026-01-01", "contributionCount": 5, "weekday": 4},
									},
								},
							},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := NewGraphQLClient("test-token", srv.URL)
	data, err := client.FetchContributions(2026)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Login != "Null-Phnix" {
		t.Errorf("expected login 'Null-Phnix', got %q", data.Login)
	}
	if data.TotalContributions != 200 {
		t.Errorf("expected 200 total, got %d", data.TotalContributions)
	}
}
