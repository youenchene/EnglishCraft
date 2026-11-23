package models_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/youenchene/EnglishCraft/internal/tutoring/adapters/models"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRealHTTPWireContracts(t *testing.T) {
	for _, provider := range []string{"openai", "anthropic", "mistral", "compatible"} {
		t.Run(provider, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var input map[string]any
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					t.Error(err)
				}
				if input["model"] != "configured-model" {
					t.Error("model selection lost")
				}
				if provider == "mistral" && input["reasoning_effort"] != "none" {
					t.Error("Mistral must return final JSON, not thinking chunks")
				}
				text := `{"text":"Well done! Try a short sentence.","mood":"encouraging","hint":"Think about yesterday."}`
				if provider == "anthropic" {
					if r.URL.Path != "/messages" || r.Header.Get("x-api-key") != "test-secret" || input["system"] == nil {
						t.Error("invalid Anthropic request")
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"content": []map[string]string{{"type": "text", "text": text}}, "stop_reason": "end_turn", "usage": map[string]int{"input_tokens": 12, "output_tokens": 15}})
					return
				}
				if r.URL.Path != "/chat/completions" || r.Header.Get("Authorization") != "Bearer test-secret" {
					t.Error("invalid chat request")
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": text}, "finish_reason": "stop"}}, "usage": map[string]int{"prompt_tokens": 12, "completion_tokens": 15}})
			}))
			defer server.Close()
			client, err := models.New(models.Config{Provider: provider, Model: "configured-model", Key: "test-secret", BaseURL: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			reply, usage, err := client.Reply(t.Context(), tutor.ConversationRequest{Message: "Hello"})
			if err != nil || reply.Mood != "encouraging" || usage.InputTokens != 12 {
				t.Fatalf("reply=%+v usage=%+v err=%v", reply, usage, err)
			}
		})
	}
}
func TestRejectUntrustedStructuredReplies(t *testing.T) {
	for _, test := range []struct{ name, text string }{
		{name: "reward injection", text: `{"text":"hello","mood":"happy","hint":"","xp":100}`},
		{name: "invalid mood", text: `{"text":"hello","mood":"evil","hint":""}`},
		{name: "empty reply", text: `{"text":"","mood":"happy","hint":""}`},
		{name: "trailing JSON", text: `{"text":"hello","mood":"happy","hint":""} {}`},
		{name: "null", text: `null`},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": test.text}, "finish_reason": "stop"}}})
			}))
			defer server.Close()
			client, _ := models.New(models.Config{Provider: "openai", Model: "test", Key: "never-expose", BaseURL: server.URL})
			_, _, err := client.Reply(t.Context(), tutor.ConversationRequest{Message: "test"})
			if !errors.Is(err, tutor.ErrInvalid) {
				t.Fatalf("got %v", err)
			}
			if strings.Contains(err.Error(), "never-expose") {
				t.Fatal("secret exposed")
			}
		})
	}
}
func TestBoundedTransientRetryAndCancellation(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(429) }))
	defer server.Close()
	client, _ := models.New(models.Config{Provider: "mistral", Model: "test", Key: "secret", BaseURL: server.URL})
	_, _, err := client.Reply(t.Context(), tutor.ConversationRequest{Message: "hi"})
	if !errors.Is(err, tutor.ErrUnavailable) || calls.Load() != 2 {
		t.Fatalf("calls=%d err=%v", calls.Load(), err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, err = client.Reply(ctx, tutor.ConversationRequest{Message: "hi"})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestProviderTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(30 * time.Millisecond); w.WriteHeader(200) }))
	defer server.Close()
	client, _ := models.New(models.Config{Provider: "openai", Model: "test", Key: "secret", BaseURL: server.URL, Timeout: 10 * time.Millisecond})
	_, _, err := client.Reply(t.Context(), tutor.ConversationRequest{Message: "hi"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}
func TestVoiceCreationAndServerHangup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/realtime/calls/rtc_test/hangup" {
			w.WriteHeader(200)
			return
		}
		if err := r.ParseMultipartForm(100000); err != nil {
			t.Fatal(err)
		}
		var session map[string]any
		if json.Unmarshal([]byte(r.FormValue("session")), &session) != nil || session["model"] != "voice-model" || session["type"] != "realtime" {
			t.Error("wrong realtime session")
		}
		if r.FormValue("sdp") != "v=0\r\n" {
			t.Error("SDP lost")
		}
		w.Header().Set("Location", "/v1/realtime/calls/rtc_test")
		w.WriteHeader(201)
		_, _ = w.Write([]byte("v=0\r\nanswer"))
	}))
	defer server.Close()
	client, err := models.NewSpeech(models.Config{Key: "secret", Model: "voice-model", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	session, err := client.Create(t.Context(), tutor.SpeechRequest{SDP: "v=0\r\n", PushToTalk: true})
	if err != nil || session.ID != "rtc_test" {
		t.Fatalf("%+v %v", session, err)
	}
	if err := client.Hangup(t.Context(), session.ID); err != nil {
		t.Fatal(err)
	}
	if err := client.Hangup(t.Context(), "../escape"); !errors.Is(err, tutor.ErrInvalid) {
		t.Fatal(err)
	}
}
