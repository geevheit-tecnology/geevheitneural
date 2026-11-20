package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDurableWorkspaceRejectsMissingPrincipal(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/cards/card/workspace", nil)
	r.SetPathValue("id", "card")
	HandleGetDurableWorkspace(nil)(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
