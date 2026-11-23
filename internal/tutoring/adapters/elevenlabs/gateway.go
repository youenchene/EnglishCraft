// Package elevenlabs owns the provider HTTP/WebSocket protocol. No provider
// token, transcript role or callback can choose a student or grant progression.
package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
)

var conversationID = regexp.MustCompile(`^conv_[A-Za-z0-9_-]{1,180}$`)
var engineID = regexp.MustCompile(`^seng_[A-Za-z0-9_-]{1,180}$`)

type Config struct {
	Key, EngineID, BaseURL string
	Clock                  tutor.Clock
	Admit                  func(context.Context, string) (tutor.VoiceRecord, error)
	Reply                  func(context.Context, string, []tutor.Turn) (string, error)
}
type connection struct {
	conn   *websocket.Conn
	cancel context.CancelFunc
}
type Gateway struct {
	config Config
	http   *http.Client
	mu     sync.Mutex
	active map[string]*connection
}

func New(config Config) (*Gateway, error) {
	if config.Key == "" || !engineID.MatchString(config.EngineID) || config.Clock == nil || config.Admit == nil || config.Reply == nil {
		return nil, tutor.ErrInvalid
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.elevenlabs.io"
	}
	u, err := url.Parse(config.BaseURL)
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost"))) {
		return nil, tutor.ErrInvalid
	}
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")
	return &Gateway{config: config, http: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, active: map[string]*connection{}}, nil
}

func (g *Gateway) request(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		raw, err := json.Marshal(input)
		if err != nil {
			return tutor.ErrInvalid
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, g.config.BaseURL+path, body)
	if err != nil {
		return tutor.ErrUnavailable
	}
	req.Header.Set("xi-api-key", g.config.Key)
	req.Header.Set("Content-Type", "application/json")
	response, err := g.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return tutor.ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return tutor.ErrUnavailable
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 128001))
	if err != nil || len(raw) > 128000 || json.Unmarshal(raw, output) != nil {
		return tutor.ErrUnavailable
	}
	return nil
}

func (g *Gateway) Create(ctx context.Context, input tutor.SpeechRequest) (tutor.SpeechSession, error) {
	// Fail closed if the externally configured engine would exceed admission's
	// duration or use the provider's default indefinite recording/retention.
	var engine struct {
		Overrides struct {
			FirstMessage bool `json:"first_message"`
		} `json:"overrides"`
		Conversation struct {
			MaxSeconds int `json:"max_duration_seconds"`
		} `json:"conversation"`
		Privacy struct {
			Record           bool `json:"record_voice"`
			Days             int  `json:"retention_days"`
			DeleteAudio      bool `json:"delete_audio"`
			DeleteTranscript bool `json:"delete_transcript_and_pii"`
		} `json:"privacy"`
	}
	if err := g.request(ctx, "GET", "/v1/speech-engine/"+g.config.EngineID, nil, &engine); err != nil {
		return tutor.SpeechSession{}, err
	}
	if !engine.Overrides.FirstMessage || engine.Conversation.MaxSeconds < 1 || engine.Conversation.MaxSeconds > input.MaxSeconds || engine.Privacy.Record || engine.Privacy.Days != 1 || !engine.Privacy.DeleteAudio || !engine.Privacy.DeleteTranscript {
		return tutor.SpeechSession{}, tutor.ErrUnavailable
	}
	var result struct {
		ID    string `json:"conversation_id"`
		Token string `json:"token"`
	}
	if err := g.request(ctx, "GET", "/v1/convai/conversation/token?agent_id="+url.QueryEscape(g.config.EngineID), nil, &result); err != nil {
		return tutor.SpeechSession{}, err
	}
	if !conversationID.MatchString(result.ID) || result.Token == "" || len(result.Token) > 12000 {
		return tutor.SpeechSession{}, tutor.ErrUnavailable
	}
	return tutor.SpeechSession{ID: result.ID, Token: result.Token, Provider: "elevenlabs"}, nil
}

func (g *Gateway) Hangup(_ context.Context, id string) error {
	if !conversationID.MatchString(id) {
		return tutor.ErrInvalid
	}
	g.mu.Lock()
	active := g.active[id]
	g.mu.Unlock()
	if active != nil {
		active.cancel()
		_ = active.conn.Close(websocket.StatusNormalClosure, "Session ended")
	}
	return nil
}

// ServeHTTP is mounted separately from browser auth and verifies the provider
// JWT before upgrading. Admission still comes exclusively from persistent Go state.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !verifyJWT(r.Header.Get("X-Elevenlabs-Speech-Engine-Authorization"), g.config.Key, g.config.Clock.Now()) {
		http.Error(w, "Unauthorized", 401)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true}) // Origin is not browser auth here; provider JWT is mandatory.
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(24000)
	initCtx, initCancel := context.WithTimeout(r.Context(), 10*time.Second)
	_, raw, err := conn.Read(initCtx)
	initCancel()
	var initial struct {
		Type string `json:"type"`
		ID   string `json:"conversation_id"`
	}
	if err != nil || json.Unmarshal(raw, &initial) != nil || initial.Type != "init" || !conversationID.MatchString(initial.ID) {
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	current := &connection{conn: conn, cancel: cancel}
	g.mu.Lock()
	if g.active[initial.ID] != nil {
		g.mu.Unlock()
		return
	}
	g.active[initial.ID] = current
	g.mu.Unlock()
	defer func() {
		g.mu.Lock()
		if g.active[initial.ID] == current {
			delete(g.active, initial.ID)
		}
		g.mu.Unlock()
	}()
	record, err := g.config.Admit(ctx, initial.ID)
	if err != nil {
		return
	}
	ctx, deadlineCancel := context.WithDeadline(ctx, record.Session.Expires)
	defer deadlineCancel()
	var callCancel context.CancelFunc
	defer func() {
		if callCancel != nil {
			callCancel()
		}
	}()
	lastEvent := -1
	turns := 0
	for {
		_, raw, err = conn.Read(ctx)
		if err != nil {
			return
		}
		var event struct {
			Type       string `json:"type"`
			ID         int    `json:"event_id"`
			Transcript []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"user_transcript"`
		}
		if json.Unmarshal(raw, &event) != nil {
			return
		}
		switch event.Type {
		case "ping":
			if write(ctx, conn, map[string]string{"type": "pong"}) != nil {
				return
			}
		case "close", "error":
			return
		case "user_transcript":
			if event.ID < 0 || event.ID <= lastEvent {
				continue
			}
			lastEvent = event.ID
			turns++
			if turns > 50 || len(event.Transcript) == 0 || len(event.Transcript) > 100 {
				return
			}
			if callCancel != nil {
				callCancel()
			}
			callCtx, abort := context.WithCancel(ctx)
			callCancel = abort
			history := []tutor.Turn{}
			start := max(0, len(event.Transcript)-12)
			for _, message := range event.Transcript[start:] {
				role := message.Role
				if role == "agent" {
					role = "assistant"
				}
				history = append(history, tutor.Turn{Role: role, Text: message.Content})
			}
			go func(eventID int, history []tutor.Turn) {
				text, err := g.config.Reply(callCtx, initial.ID, history)
				if callCtx.Err() != nil {
					return
				}
				if err != nil {
					cancel()
					_ = conn.Close(websocket.StatusPolicyViolation, "Tutoring unavailable")
					return
				}
				_ = write(callCtx, conn, map[string]any{"type": "agent_response", "event_id": eventID, "content": text, "is_final": true})
			}(event.ID, history)
		}
	}
}
func write(ctx context.Context, conn *websocket.Conn, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, raw)
}
