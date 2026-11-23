package httpapi

import (
	"context"
	"github.com/youenchene/EnglishCraft/internal/curriculum/adapters/content"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeTutoringStore struct{}

func (fakeTutoringStore) Enabled(context.Context, string) (bool, error)  { return false, nil }
func (fakeTutoringStore) SetEnabled(context.Context, string, bool) error { return nil }
func (fakeTutoringStore) Reserve(context.Context, tutor.Allowance) (string, error) {
	return "", tutor.ErrBudget
}
func (fakeTutoringStore) Record(context.Context, string, tutor.Usage, string) error { return nil }
func (fakeTutoringStore) Spending(context.Context, string, time.Time) ([]tutor.Spending, error) {
	return []tutor.Spending{}, nil
}
func (fakeTutoringStore) PutImport(context.Context, string, tutor.SchoolImport) error { return nil }
func (fakeTutoringStore) Imports(context.Context, string, bool) ([]tutor.SchoolImport, error) {
	return []tutor.SchoolImport{}, nil
}
func (fakeTutoringStore) ApproveImport(context.Context, string, string) error { return nil }
func TestTutoringAuthorizationAndServerRewards(t *testing.T) {
	catalog, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	repo := &coreMemory{state: core.NewState()}
	learning := core.New(catalog, repo, httpClock{})
	sessions := identity.NewSessions(&authMemory{sessions: map[string]identity.Principal{}}, passwords{}, httpClock{})
	handler := NewCore(learning, sessions, false)
	mux := http.NewServeMux()
	handler.Register(mux)
	service := tutor.New(tutor.Dependencies{Learning: learning, Budget: fakeTutoringStore{}, Clock: httpClock{}})
	handler.RegisterTutoring(mux, service, tutor.NewImports(fakeTutoringStore{}, learning))
	request := func(method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		SameOrigin(mux).ServeHTTP(w, r)
		return w
	}
	if request("POST", "/api/v1/tutor/lost-diamond/text", `{"message":"hello"}`, nil).Code != 401 {
		t.Fatal("anonymous AI access")
	}
	child := request("POST", "/api/v1/auth/login", `{"username":"alex","password":"password"}`, nil).Result().Cookies()[0]
	parent := request("POST", "/api/v1/auth/login", `{"username":"parent","password":"password"}`, nil).Result().Cookies()[0]
	if request("GET", "/api/v1/tutor/lost-diamond/intro", "", nil).Code != 401 {
		t.Fatal("anonymous intro access")
	}
	if request("GET", "/api/v1/tutor/lost-diamond/intro", "", parent).Code != 403 {
		t.Fatal("parent used child intro")
	}
	if request("GET", "/api/v1/tutor/lost-diamond/voice", "", nil).Code != 401 {
		t.Fatal("anonymous voice availability access")
	}
	if request("GET", "/api/v1/tutor/lost-diamond/voice", "", parent).Code != 403 {
		t.Fatal("parent used child voice availability")
	}
	voice := request("GET", "/api/v1/tutor/lost-diamond/voice", "", child)
	if voice.Code != 503 || !strings.Contains(voice.Body.String(), `"code":"voice_disabled"`) {
		t.Fatalf("missing voice configuration reason: %d %s", voice.Code, voice.Body.String())
	}
	if request("PUT", "/api/v1/parent/ai", `{"enabled":true}`, child).Code != 403 {
		t.Fatal("child enabled paid providers")
	}
	if request("POST", "/api/v1/parent/imports", `{"text":"went"}`, child).Code != 403 {
		t.Fatal("child published curriculum")
	}
	if request("POST", "/api/v1/tutor/lost-diamond/tools", `{"name":"request_hint","arguments":{}}`, parent).Code != 403 {
		t.Fatal("parent used child identity")
	}
	request("POST", "/api/v1/missions/lost-diamond/start", `{}`, child)
	intro := request("GET", "/api/v1/tutor/lost-diamond/intro", "", child)
	if intro.Code != 200 || !strings.Contains(intro.Body.String(), "Speak English") || strings.Contains(intro.Body.String(), "Le diamant") {
		t.Fatal("intro mixes curriculum dialogue into story/rules")
	}
	text := request("POST", "/api/v1/tutor/lost-diamond/text", `{"message":"hello"}`, child)
	if text.Code != 200 || !strings.Contains(text.Body.String(), `"mode":"local"`) {
		t.Fatal(text.Body.String())
	}
	cheat := request("POST", "/api/v1/tutor/lost-diamond/tools", `{"name":"submit_student_answer","arguments":{"student":"other","step":0,"exerciseId":"diamond-go","answer":"went","idempotencyKey":"abcdefghijklmnop"}}`, child)
	if cheat.Code != 400 {
		t.Fatal("tool accepted foreign identity")
	}
	command := `{"name":"submit_student_answer","arguments":{"step":0,"exerciseId":"diamond-go","answer":"went","idempotencyKey":"abcdefghijklmnop"}}`
	first := request("POST", "/api/v1/tutor/lost-diamond/tools", command, child)
	retry := request("POST", "/api/v1/tutor/lost-diamond/tools", command, child)
	if first.Code != 200 || retry.Body.String() != first.Body.String() || len(repo.state.Attempts) != 1 {
		t.Fatal("tool retry duplicated learning evidence")
	}
	if !strings.Contains(first.Body.String(), "AWARD_XP") {
		t.Fatal("server reward event missing")
	}
	if request("GET", "/api/v1/parent/ai", "", parent).Code != 200 {
		t.Fatal("parent usage unavailable")
	}
}
func TestVoiceHeadersPermitOnlySameOriginMicrophone(t *testing.T) {
	w := httptest.NewRecorder()
	SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(w.Header().Get("Permissions-Policy"), "microphone=(self)") {
		t.Fatal("microphone prohibited")
	}
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "media-src 'self' blob:") {
		t.Fatal("WebRTC audio prohibited")
	}
}

func TestVoiceFailureReasons(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
		code   string
	}{
		{tutor.ErrVoiceDisabled, 503, "voice_disabled"},
		{tutor.ErrAIConsent, 403, "ai_consent_required"},
		{tutor.ErrUnavailable, 503, "unavailable"},
	} {
		w := httptest.NewRecorder()
		tutoringError(w, test.err)
		if w.Code != test.status || !strings.Contains(w.Body.String(), `"code":"`+test.code+`"`) {
			t.Fatalf("wrong voice reason: %d %s", w.Code, w.Body.String())
		}
	}
}
