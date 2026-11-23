package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
)

type verifierStub struct{}

func (verifierStub) Verify(context.Context, string) (identity.Principal, error) {
	return identity.Principal{SubjectID: "student-1", Role: identity.RoleChild}, nil
}

func TestDecodeJSONRejectsTrailingContent(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"pastSimple":"went"} trailing`))
	response := httptest.NewRecorder()
	var destination answerRequest

	if err := decodeJSON(response, request, &destination); err == nil {
		t.Fatal("expected trailing content to be rejected")
	}
}

func TestAuthenticateAddsPrincipal(t *testing.T) {
	authenticate := identity.NewAuthenticate(verifierStub{})
	next := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		principal, ok := PrincipalFromContext(request.Context())
		if !ok || principal.SubjectID != "student-1" {
			t.Fatal("principal missing from context")
		}
		response.WriteHeader(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()

	Authenticate(authenticate, next).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("got status %d", response.Code)
	}
}
