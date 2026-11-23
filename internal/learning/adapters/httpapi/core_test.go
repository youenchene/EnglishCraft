package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/youenchene/EnglishCraft/internal/curriculum/adapters/content"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
)

type coreMemory struct{ state core.StudentState }

func (m *coreMemory) Read(context.Context, string) (core.StudentState, error) { return m.state, nil }
func (m *coreMemory) Transact(_ context.Context, _ string, fn func(*core.StudentState) error) error {
	data, err := json.Marshal(m.state)
	if err != nil {
		return err
	}
	copy := core.NewState()
	if err := json.Unmarshal(data, &copy); err != nil {
		return err
	}
	if err := fn(&copy); err != nil {
		return err
	}
	m.state = copy
	return nil
}

type authMemory struct{ sessions map[string]identity.Principal }

func (m *authMemory) Login(_ context.Context, username string, fn func(*identity.Account) (identity.Session, error)) error {
	role := identity.RoleChild
	if username == "parent" {
		role = identity.RoleParent
	}
	account := identity.Account{ID: username, Role: role, PasswordHash: "password"}
	s, err := fn(&account)
	if err == nil {
		m.sessions[s.Hash] = identity.Principal{SubjectID: username, Role: role}
	}
	return err
}
func (m *authMemory) Resolve(_ context.Context, hash string, _ time.Time) (identity.Principal, error) {
	p, ok := m.sessions[hash]
	if !ok {
		return p, identity.ErrInvalidCredentials
	}
	return p, nil
}
func (m *authMemory) Revoke(_ context.Context, hash string) error {
	delete(m.sessions, hash)
	return nil
}
func (m *authMemory) ChildForParent(context.Context, string) (string, error) { return "alex", nil }
func (m *authMemory) Provision(context.Context, identity.Family) error       { return nil }

type passwords struct{}

func (passwords) Hash(p string) (string, error) { return p, nil }
func (passwords) Verify(hash, p string) bool    { return hash == p }

type httpClock struct{}

func (httpClock) Now() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) }

func TestCookieAuthRolesDTOsAndMissionJourney(t *testing.T) {
	catalog, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	repo := &coreMemory{state: core.NewState()}
	sessions := identity.NewSessions(&authMemory{sessions: map[string]identity.Principal{}}, passwords{}, httpClock{})
	mux := http.NewServeMux()
	NewCore(core.New(catalog, repo, httpClock{}), sessions, true).Register(mux)
	handler := SameOrigin(mux)
	request := func(method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	if got := request("GET", "/api/v1/missions", "", nil).Code; got != 401 {
		t.Fatalf("anonymous got %d", got)
	}
	login := request("POST", "/api/v1/auth/login", `{"username":"alex","password":"password"}`, nil)
	if login.Code != 200 {
		t.Fatal(login.Body.String())
	}
	cookie := login.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatal("insecure session cookie")
	}
	if got := request("GET", "/api/v1/parent/results", "", cookie).Code; got != 403 {
		t.Fatal("child accessed parent results")
	}
	start := request("POST", "/api/v1/missions/lost-diamond/start", `{}`, cookie)
	if start.Code != 200 || strings.Contains(start.Body.String(), "accepted") {
		t.Fatal("start failed or correction leaked")
	}
	cheat := request("POST", "/api/v1/answers", `{"xp":999,"exerciseId":"diamond-go","answer":"went","idempotencyKey":"abcdefghijklmnop"}`, cookie)
	if cheat.Code != 400 {
		t.Fatal("client reward field accepted")
	}
	answer := `{"missionId":"lost-diamond","step":0,"exerciseId":"diamond-go","answer":"went","idempotencyKey":"abcdefghijklmnop"}`
	first := request("POST", "/api/v1/answers", answer, cookie)
	retry := request("POST", "/api/v1/answers", answer, cookie)
	if first.Code != 200 || retry.Body.String() != first.Body.String() || len(repo.state.Attempts) != 1 {
		t.Fatal("HTTP retry duplicated reward")
	}
	parentLogin := request("POST", "/api/v1/auth/login", `{"username":"parent","password":"password"}`, nil)
	parentCookie := parentLogin.Result().Cookies()[0]
	if request("GET", "/api/v1/parent/results", "", parentCookie).Code != 200 {
		t.Fatal("parent cannot view results")
	}
	if request("POST", "/api/v1/missions/lost-diamond/start", `{}`, parentCookie).Code != 403 {
		t.Fatal("parent submitted as child")
	}
	if request("POST", "/api/v1/auth/logout", `{}`, cookie).Code != 204 || request("GET", "/api/v1/auth/me", "", cookie).Code != 401 {
		t.Fatal("logout did not revoke session")
	}
}
func TestCookieMutationsRejectCrossOrigin(t *testing.T) {
	h := SameOrigin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	r := httptest.NewRequest("POST", "http://localhost/api/v1/auth/login", nil)
	r.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross-origin mutation accepted")
	}
}
