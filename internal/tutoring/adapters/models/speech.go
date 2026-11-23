package models

import (
	"bytes"
	"context"
	"encoding/json"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
)

var callIDPattern = regexp.MustCompile(`^rtc_[A-Za-z0-9_-]{1,180}$`)

type Speech struct{ client *Client }

func NewSpeech(config Config) (*Speech, error) {
	config.Provider = "openai"
	client, err := New(config)
	if err != nil {
		return nil, err
	}
	return &Speech{client: client}, nil
}
func voiceTools() []map[string]any {
	tools := []map[string]any{}
	for _, name := range []string{"get_current_objective", "get_current_challenge", "request_hint", "get_review_items", "complete_mission_step"} {
		tools = append(tools, map[string]any{"type": "function", "name": name,
			"description": "Read authorised learning state. Completion never skips a challenge.",
			"parameters":  map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}})
	}
	tools = append(tools, map[string]any{"type": "function", "name": "submit_student_answer",
		"description": "Submit only the student's actual answer to the current challenge. Never invent an answer. Server determines correction and rewards.",
		"parameters": map[string]any{"type": "object", "additionalProperties": false,
			"required": []string{"step", "exerciseId", "answer", "idempotencyKey"},
			"properties": map[string]any{"step": map[string]string{"type": "integer"},
				"exerciseId": map[string]string{"type": "string"}, "answer": map[string]string{"type": "string"},
				"idempotencyKey": map[string]string{"type": "string", "description": "Unique key of 16 to 100 characters; reuse on retry."}}}})
	return tools
}
func (s *Speech) Create(ctx context.Context, input tutor.SpeechRequest) (tutor.SpeechSession, error) {
	var detection any = map[string]any{"type": "server_vad", "interrupt_response": true, "create_response": true}
	if input.PushToTalk {
		detection = nil
	}
	session := map[string]any{"type": "realtime", "model": s.client.config.Model,
		"instructions": Instructions(input.Context), "tools": voiceTools(), "max_output_tokens": 600,
		"audio": map[string]any{"input": map[string]any{"turn_detection": detection,
			"transcription": map[string]string{"model": "gpt-4o-mini-transcribe", "language": "en"}},
			"output": map[string]string{"voice": "marin"}}}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	config, _ := json.Marshal(session)
	if err := writer.WriteField("sdp", input.SDP); err != nil {
		return tutor.SpeechSession{}, err
	}
	if err := writer.WriteField("session", string(config)); err != nil {
		return tutor.SpeechSession{}, err
	}
	if err := writer.Close(); err != nil {
		return tutor.SpeechSession{}, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", s.client.config.BaseURL+"/realtime/calls", &body)
	if err != nil {
		return tutor.SpeechSession{}, tutor.ErrUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+s.client.config.Key)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := s.client.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return tutor.SpeechSession{}, ctx.Err()
		}
		return tutor.SpeechSession{}, tutor.ErrUnavailable
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 64001))
	id := response.Header.Get("Location")
	id = id[strings.LastIndex(id, "/")+1:]
	if err != nil || response.StatusCode != 201 || len(raw) > 64000 || !callIDPattern.MatchString(id) || !bytes.HasPrefix(raw, []byte("v=0")) {
		return tutor.SpeechSession{}, tutor.ErrUnavailable
	}
	return tutor.SpeechSession{ID: id, SDP: string(raw)}, nil
}
func (s *Speech) Hangup(ctx context.Context, id string) error {
	if !callIDPattern.MatchString(id) {
		return tutor.ErrInvalid
	}
	req, err := http.NewRequestWithContext(ctx, "POST", s.client.config.BaseURL+"/realtime/calls/"+id+"/hangup", nil)
	if err != nil {
		return tutor.ErrUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+s.client.config.Key)
	response, err := s.client.http.Do(req)
	if err != nil {
		return tutor.ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 || response.StatusCode == 404 {
		return nil
	}
	return tutor.ErrUnavailable
}
