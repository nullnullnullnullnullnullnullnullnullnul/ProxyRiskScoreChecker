package ipqs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckIPSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("strictness"); got != "2" {
			t.Errorf("strictness query = %q; want %q", got, "2")
		}
		if !strings.Contains(r.URL.Path, "1.2.3.4") {
			t.Errorf("path %q does not include ip", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Response{Success: true, FraudScore: 0})
	}))
	defer srv.Close()

	c := Client{APIKey: "k", Timeout: 2 * time.Second, BaseURL: srv.URL}
	resp, err := c.CheckIP(context.Background(), "1.2.3.4", 2)
	if err != nil {
		t.Fatalf("CheckIP: %v", err)
	}
	if !resp.Success || resp.FraudScore != 0 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestCheckIPReportsFailureMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Response{Success: false, Message: "bad key"})
	}))
	defer srv.Close()

	c := Client{APIKey: "k", Timeout: time.Second, BaseURL: srv.URL}
	_, err := c.CheckIP(context.Background(), "1.2.3.4", 0)
	if err == nil {
		t.Fatal("expected error from failing response")
	}
	if !strings.Contains(err.Error(), "bad key") {
		t.Errorf("error %q does not include upstream message", err.Error())
	}
}

func TestCheckIPNonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := Client{APIKey: "k", Timeout: time.Second, BaseURL: srv.URL}
	if _, err := c.CheckIP(context.Background(), "1.2.3.4", 0); err == nil {
		t.Fatal("expected error from HTTP 500")
	}
}

func TestCheckIPMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := Client{APIKey: "k", Timeout: time.Second, BaseURL: srv.URL}
	if _, err := c.CheckIP(context.Background(), "1.2.3.4", 0); err == nil {
		t.Fatal("expected error from malformed JSON")
	}
}
