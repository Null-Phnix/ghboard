// api/rest_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListStars(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/starred" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]StarredRepo{
			{FullName: "owner/repo", Description: "test", Language: "Go", StargazersCount: 42},
		})
	}))
	defer srv.Close()

	client := NewRESTClient("test-token", srv.URL)
	repos, err := client.ListStars(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || repos[0].FullName != "owner/repo" {
		t.Errorf("unexpected repos: %v", repos)
	}
}

func TestListNotifications(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Notification{
			{ID: "1", Unread: true, Reason: "mention", Subject: Subject{Title: "Test PR", Type: "PullRequest"}},
		})
	}))
	defer srv.Close()

	client := NewRESTClient("test-token", srv.URL)
	notifs, err := client.ListNotifications()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifs) != 1 || notifs[0].ID != "1" {
		t.Errorf("unexpected notifications: %v", notifs)
	}
}
