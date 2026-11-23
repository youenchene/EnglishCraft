//go:build external

package external_test

import (
	"github.com/youenchene/EnglishCraft/internal/tutoring/adapters/models"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
	"os"
	"testing"
	"time"
)

// Explicit opt-in, billable test. Requires an actual SDP offer from a browser;
// a dummy SDP cannot establish or validate a real WebRTC voice connection.
func TestOpenAIRealtimeCreateAndHangup(t *testing.T) {
	path := os.Getenv("TEST_REALTIME_SDP_FILE")
	if os.Getenv("ALLOW_BILLABLE_TESTS") != "true" || os.Getenv("OPENAI_API_KEY") == "" || path == "" {
		t.Skip("billable test disabled; key and browser SDP required")
	}
	sdp, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gateway, err := models.NewSpeech(models.Config{Key: os.Getenv("OPENAI_API_KEY"), Model: os.Getenv("AI_VOICE_MODEL"), Timeout: 10 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	session, err := gateway.Create(t.Context(), tutor.SpeechRequest{SDP: string(sdp)})
	if err != nil {
		t.Fatal(err)
	}
	if err := gateway.Hangup(t.Context(), session.ID); err != nil {
		t.Fatal(err)
	}
}
