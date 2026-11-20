package vercelqueue

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPublisherSendsOnlyJobIDAndRequiresOIDC(t *testing.T) {
	id := uuid.NewString()
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.String() != "https://iad1.vercel-queue.com/api/v3/topic/factory" {
			t.Errorf("unexpected endpoint: %s", r.URL)
		}
		if r.Header.Get("Authorization") != "Bearer token-test" || r.Header.Get("Vqs-Idempotency-Key") != id {
			t.Error("missing authentication or idempotency key")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"job_id":"`+id+`"}` {
			t.Errorf("unexpected body: %s", body)
		}
		return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header)}, nil
	})}
	p, err := NewPublisher("iad1", "factory", client)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Publish(context.Background(), "", id); err == nil {
		t.Fatal("missing OIDC accepted")
	}
	if err := p.Publish(context.Background(), "token-test", "invalid"); err == nil {
		t.Fatal("invalid job accepted")
	}
	if calls != 0 {
		t.Fatal("request sent before validation")
	}
	if err := p.Publish(context.Background(), "token-test", id); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("requests: %d", calls)
	}
}

func TestPublisherFailureIsReportedWithoutResponseBody(t *testing.T) {
	if _, err := NewPublisher("iad1.evil", "factory", nil); err == nil {
		t.Fatal("invalid region accepted")
	}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("sensitive-body")), Header: make(http.Header)}, nil
	})}
	p, _ := NewPublisher("iad1", "factory", client)
	err := p.Publish(context.Background(), "secret-token", uuid.NewString())
	if err == nil || strings.Contains(err.Error(), "sensitive-body") || strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("unsafe error: %v", err)
	}
}
