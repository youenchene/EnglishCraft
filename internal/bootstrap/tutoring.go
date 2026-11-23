package bootstrap

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
	"github.com/youenchene/EnglishCraft/internal/platform/system"
	"github.com/youenchene/EnglishCraft/internal/tutoring/adapters/elevenlabs"
	"github.com/youenchene/EnglishCraft/internal/tutoring/adapters/models"
	postgres "github.com/youenchene/EnglishCraft/internal/tutoring/adapters/postgres"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
	"net/http"
	"os"
	"strconv"
	"time"
)

func tutoring(pool *pgxpool.Pool, learning *core.Service) (*tutor.Service, *tutor.Imports, http.Handler, error) {
	policy := tutor.Policy{Provider: os.Getenv("AI_PROVIDER"), VoiceProvider: os.Getenv("VOICE_PROVIDER"), VoiceMaxSeconds: 300, VoiceDailySeconds: 900}
	values := []struct {
		name   string
		target *int64
	}{
		{name: "AI_PROVIDER_DAILY_MICROS", target: &policy.DailyMicros},
		{name: "AI_STUDENT_DAILY_MICROS", target: &policy.StudentDailyMicros},
		{name: "AI_TEXT_RESERVE_MICROS", target: &policy.TextReserveMicros},
		{name: "AI_VOICE_RESERVE_MICROS", target: &policy.VoiceReserveMicros},
	}
	for _, value := range values {
		if raw := os.Getenv(value.name); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || parsed < 0 || parsed > 1_000_000_000 {
				return nil, nil, nil, fmt.Errorf("invalid %s", value.name)
			}
			*value.target = parsed
		}
	}
	for _, value := range []struct {
		name   string
		target *int
		max    int
	}{
		{name: "VOICE_MAX_SECONDS", target: &policy.VoiceMaxSeconds, max: 1200},
		{name: "VOICE_DAILY_SECONDS", target: &policy.VoiceDailySeconds, max: 3600},
	} {
		if raw := os.Getenv(value.name); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > value.max {
				return nil, nil, nil, fmt.Errorf("invalid %s", value.name)
			}
			*value.target = parsed
		}
	}
	var conversation tutor.TutorConversation
	if policy.Provider != "" && policy.Provider != "disabled" {
		keyName := map[string]string{"openai": "OPENAI_API_KEY", "anthropic": "ANTHROPIC_API_KEY", "mistral": "MISTRAL_API_KEY", "compatible": "AI_API_KEY"}[policy.Provider]
		client, err := models.New(models.Config{Provider: policy.Provider, Key: os.Getenv(keyName), Model: os.Getenv("AI_TEXT_MODEL"), BaseURL: os.Getenv("AI_BASE_URL"), Timeout: 10 * time.Second})
		if err != nil {
			return nil, nil, nil, err
		}
		conversation = client
	}
	var speech tutor.SpeechSessionGateway
	var upstream http.Handler
	var service *tutor.Service
	if os.Getenv("VOICE_ENABLED") == "true" {
		switch policy.VoiceProvider {
		case "elevenlabs":
			if conversation == nil {
				return nil, nil, nil, fmt.Errorf("ElevenLabs requires a text tutor provider")
			}
			client, err := elevenlabs.New(elevenlabs.Config{Key: os.Getenv("ELEVENLABS_API_KEY"), EngineID: os.Getenv("ELEVENLABS_ENGINE_ID"), Clock: system.Clock{},
				Admit: func(ctx context.Context, id string) (tutor.VoiceRecord, error) { return service.AdmitVoice(ctx, id) },
				Reply: func(ctx context.Context, id string, history []tutor.Turn) (string, error) {
					return service.VoiceReply(ctx, id, history)
				}})
			if err != nil {
				return nil, nil, nil, err
			}
			speech = client
			upstream = client
		case "", "openai":
			client, err := models.NewSpeech(models.Config{Key: os.Getenv("OPENAI_API_KEY"), Model: os.Getenv("AI_VOICE_MODEL"), Timeout: 10 * time.Second})
			if err != nil {
				return nil, nil, nil, err
			}
			speech = client
		default:
			return nil, nil, nil, fmt.Errorf("unsupported VOICE_PROVIDER")
		}
	}
	store := postgres.New(pool)
	service = tutor.New(tutor.Dependencies{Learning: learning, Tutor: conversation, Speech: speech, Budget: store, Voices: store, Clock: system.Clock{}, Policy: policy})
	return service, tutor.NewImports(store, learning), upstream, nil
}
