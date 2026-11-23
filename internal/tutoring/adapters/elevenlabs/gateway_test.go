package elevenlabs

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
)

type clock struct{}

func (clock) Now() time.Time { return time.Now() }
func sign(key, alg string, expiry time.Time) string {
	header, _ := json.Marshal(map[string]string{"alg": alg, "typ": "JWT"})
	payload, _ := json.Marshal(map[string]any{"iss": "https://api.elevenlabs.io/convai/speech-engine", "sub": "convai_speech_engine_upstream", "iat": time.Now().Unix(), "exp": expiry.Unix()})
	parts := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	secret := sha256.Sum256([]byte(key))
	mac := hmac.New(sha256.New, secret[:])
	_, _ = mac.Write([]byte(parts))
	return parts + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func config() Config {
	return Config{Key: "test-key", EngineID: "seng_test", Clock: clock{}, Admit: func(context.Context, string) (tutor.VoiceRecord, error) {
		return tutor.VoiceRecord{Session: tutor.SpeechSession{Expires: time.Now().Add(time.Minute)}}, nil
	}, Reply: func(context.Context, string, []tutor.Turn) (string, error) { return "Hello!", nil }}
}
func TestProviderJWT(t *testing.T) {
	now := time.Now()
	if !verifyJWT(sign("test-key", "HS256", now.Add(time.Minute)), "test-key", now) {
		t.Fatal("valid JWT rejected")
	}
	if !verifyJWT("Bearer "+sign("test-key", "HS256", now.Add(time.Minute)), "test-key_residency_eu", now) {
		t.Fatal("residency key not normalized")
	}
	for _, token := range []string{"", "x.y.z", sign("other", "HS256", now.Add(time.Minute)), sign("test-key", "none", now.Add(time.Minute)), sign("test-key", "HS256", now.Add(-2*time.Minute))} {
		if verifyJWT(token, "test-key", now) {
			t.Fatal("invalid JWT accepted")
		}
	}
}
func TestCreateChecksProviderSafetyBeforeMintingToken(t *testing.T) {
	recorded := false
	firstMessage := true
	minted := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("xi-api-key") != "test-key" {
			t.Error("key missing")
		}
		if strings.Contains(r.URL.Path, "speech-engine") {
			_ = json.NewEncoder(w).Encode(map[string]any{"overrides": map[string]bool{"first_message": firstMessage}, "conversation": map[string]int{"max_duration_seconds": 60}, "privacy": map[string]any{"record_voice": recorded, "retention_days": 1, "delete_audio": true, "delete_transcript_and_pii": true}})
		} else {
			minted++
			_ = json.NewEncoder(w).Encode(map[string]string{"conversation_id": "conv_test", "token": "ephemeral"})
		}
	}))
	defer server.Close()
	c := config()
	c.BaseURL = server.URL
	gateway, _ := New(c)
	session, err := gateway.Create(t.Context(), tutor.SpeechRequest{MaxSeconds: 60})
	if err != nil || session.Token != "ephemeral" || session.ID != "conv_test" {
		t.Fatalf("%+v %v", session, err)
	}
	recorded = true
	if _, err := gateway.Create(t.Context(), tutor.SpeechRequest{MaxSeconds: 60}); err == nil || minted != 1 {
		t.Fatal("recording accepted")
	}
	recorded = false
	firstMessage = false
	if _, err := gateway.Create(t.Context(), tutor.SpeechRequest{MaxSeconds: 60}); err == nil || minted != 1 {
		t.Fatal("engine cannot let Alex speak first")
	}
	firstMessage = true
	if _, err := gateway.Create(t.Context(), tutor.SpeechRequest{MaxSeconds: 30}); err == nil || minted != 1 {
		t.Fatal("provider duration exceeds admission")
	}
}
func TestCallbackRejectsAnonymousAndUnknownAdmission(t *testing.T) {
	c := config()
	c.Admit = func(context.Context, string) (tutor.VoiceRecord, error) { return tutor.VoiceRecord{}, tutor.ErrInvalid }
	called := false
	c.Reply = func(context.Context, string, []tutor.Turn) (string, error) { called = true; return "", nil }
	gateway, _ := New(c)
	server := httptest.NewServer(gateway)
	defer server.Close()
	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 401 {
		t.Fatal("anonymous callback accepted")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, server.URL, &websocket.DialOptions{HTTPHeader: http.Header{"X-Elevenlabs-Speech-Engine-Authorization": []string{sign("test-key", "HS256", time.Now().Add(time.Minute))}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	_ = write(ctx, conn, map[string]string{"type": "init", "conversation_id": "conv_unknown"})
	if _, _, err := conn.Read(ctx); err == nil {
		t.Fatal("unknown admission kept open")
	}
	if called {
		t.Fatal("unknown admission called LLM")
	}
}
func TestInterruptedReplyIsCancelledAndDuplicatesDoNotCallLLM(t *testing.T) {
	c := config()
	first := make(chan struct{})
	cancelled := make(chan struct{})
	var calls atomic.Int32
	c.Reply = func(ctx context.Context, _ string, _ []tutor.Turn) (string, error) {
		if calls.Add(1) == 1 {
			close(first)
			<-ctx.Done()
			close(cancelled)
			return "stale", nil
		}
		return "new reply", nil
	}
	gateway, _ := New(c)
	server := httptest.NewServer(gateway)
	defer server.Close()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, server.URL, &websocket.DialOptions{HTTPHeader: http.Header{"X-Elevenlabs-Speech-Engine-Authorization": []string{sign("test-key", "HS256", time.Now().Add(time.Minute))}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	_ = write(ctx, conn, map[string]string{"type": "init", "conversation_id": "conv_test"})
	event := func(id int) map[string]any {
		return map[string]any{"type": "user_transcript", "event_id": id, "user_transcript": []map[string]string{{"role": "user", "content": "Hello"}}}
	}
	_ = write(ctx, conn, event(1))
	select {
	case <-first:
	case <-ctx.Done():
		t.Fatal("first call not started")
	}
	_ = write(ctx, conn, event(2))
	select {
	case <-cancelled:
	case <-ctx.Done():
		t.Fatal("old call not cancelled")
	}
	_, raw, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var reply struct {
		Content string `json:"content"`
		ID      int    `json:"event_id"`
	}
	_ = json.Unmarshal(raw, &reply)
	if reply.ID != 2 || reply.Content != "new reply" {
		t.Fatalf("stale output: %s", raw)
	}
	_ = write(ctx, conn, event(2))
	_ = write(ctx, conn, map[string]string{"type": "ping"})
	_, raw, err = conn.Read(ctx)
	if err != nil || !strings.Contains(string(raw), "pong") || calls.Load() != 2 {
		t.Fatal("duplicate response or failed ping")
	}
	if err := gateway.Hangup(ctx, "conv_test"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := conn.Read(ctx); err == nil {
		t.Fatal("server hangup kept callback open")
	}
}
