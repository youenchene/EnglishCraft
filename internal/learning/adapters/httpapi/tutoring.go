package httpapi

import (
	"errors"
	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
	tutor "github.com/youenchene/EnglishCraft/internal/tutoring/application"
	"io"
	"net/http"
)

// Reuse the existing session boundary, not a second authentication system.
func (h *CoreHandler) RegisterTutoring(mux *http.ServeMux, service *tutor.Service, imports *tutor.Imports) {
	mux.Handle("GET /api/v1/tutor/{mission}/intro", h.auth(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())
		text, err := service.Introduction(r.Context(), p.SubjectID, r.PathValue("mission"))
		if err != nil {
			tutoringError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, map[string]string{"text": text})
	})))
	mux.Handle("POST /api/v1/tutor/{mission}/text", h.auth(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Message string `json:"message"`
		}
		if decodeJSON(w, r, &input) != nil {
			tutoringError(w, tutor.ErrInvalid)
			return
		}
		p, _ := PrincipalFromContext(r.Context())
		result, err := service.Converse(r.Context(), p.SubjectID, r.PathValue("mission"), input.Message)
		if err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 200, result)
	})))
	mux.Handle("POST /api/v1/tutor/{mission}/tools", h.auth(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input tutor.ToolRequest
		if decodeJSON(w, r, &input) != nil {
			tutoringError(w, tutor.ErrInvalid)
			return
		}
		p, _ := PrincipalFromContext(r.Context())
		result, err := service.Tool(r.Context(), p.SubjectID, r.PathValue("mission"), input)
		if err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 200, result)
	})))
	mux.Handle("GET /api/v1/tutor/{mission}/voice", h.auth(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())
		if err := service.VoiceReady(r.Context(), p.SubjectID, r.PathValue("mission")); err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"available": true, "provider": service.VoiceProvider()})
	})))
	mux.Handle("POST /api/v1/tutor/{mission}/voice", h.auth(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SDP needs a larger envelope than ordinary learning commands.
		var input struct {
			SDP        string `json:"sdp"`
			PushToTalk bool   `json:"pushToTalk"`
		}
		if decodeSizedJSON(w, r, &input, 66000) != nil {
			tutoringError(w, tutor.ErrInvalid)
			return
		}
		p, _ := PrincipalFromContext(r.Context())
		result, err := service.StartVoice(r.Context(), p.SubjectID, r.PathValue("mission"), input.SDP, input.PushToTalk)
		if err != nil {
			tutoringError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 201, result)
	})))
	mux.Handle("DELETE /api/v1/tutor/voice/{id}", h.auth(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())
		if err := service.StopVoice(r.Context(), p.SubjectID, r.PathValue("id")); err != nil {
			tutoringError(w, err)
			return
		}
		w.WriteHeader(204)
	})))
	mux.Handle("GET /api/v1/plan", h.auth(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())
		value, err := service.Plan(r.Context(), p.SubjectID)
		if err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 200, value)
	})))
	mux.Handle("GET /api/v1/parent/ai", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		child, ok := h.parentChild(w, r)
		if !ok {
			return
		}
		value, err := service.Spending(r.Context(), child)
		if err != nil {
			tutoringError(w, err)
			return
		}
		enabled, err := service.AIEnabled(r.Context(), child)
		if err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"enabled": enabled, "spending": value})
	})))
	mux.Handle("PUT /api/v1/parent/ai", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		child, ok := h.parentChild(w, r)
		if !ok {
			return
		}
		var input struct {
			Enabled *bool `json:"enabled"`
		}
		if decodeJSON(w, r, &input) != nil || input.Enabled == nil {
			tutoringError(w, tutor.ErrInvalid)
			return
		}
		if err := service.SetAIEnabled(r.Context(), child, *input.Enabled); err != nil {
			tutoringError(w, err)
			return
		}
		w.WriteHeader(204)
	})))
	mux.Handle("GET /api/v1/parent/plan", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		child, ok := h.parentChild(w, r)
		if !ok {
			return
		}
		value, err := service.Plan(r.Context(), child)
		if err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 200, value)
	})))
	mux.Handle("POST /api/v1/parent/imports", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		child, ok := h.parentChild(w, r)
		if !ok {
			return
		}
		var input struct {
			Text string `json:"text"`
		}
		if decodeSizedJSON(w, r, &input, 64000) != nil {
			tutoringError(w, tutor.ErrInvalid)
			return
		}
		value, err := imports.ImportText(r.Context(), child, input.Text)
		if err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 201, value)
	})))
	mux.Handle("GET /api/v1/parent/imports", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		child, ok := h.parentChild(w, r)
		if !ok {
			return
		}
		value, err := imports.List(r.Context(), child, false)
		if err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 200, value)
	})))
	mux.Handle("POST /api/v1/parent/imports/{id}/approve", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		child, ok := h.parentChild(w, r)
		if !ok {
			return
		}
		if err := imports.Approve(r.Context(), child, r.PathValue("id")); err != nil {
			tutoringError(w, err)
			return
		}
		w.WriteHeader(204)
	})))
	mux.Handle("GET /api/v1/school-lessons", h.auth(true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())
		value, err := imports.List(r.Context(), p.SubjectID, true)
		if err != nil {
			tutoringError(w, err)
			return
		}
		writeJSON(w, 200, value)
	})))
}
func decodeSizedJSON(w http.ResponseWriter, r *http.Request, destination any, limit int64) error {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
	if err != nil {
		return tutor.ErrInvalid
	}
	return tutor.StrictJSON(raw, destination)
}
func (h *CoreHandler) RegisterDeletion(mux *http.ServeMux, service *identity.Deletion, tutoring *tutor.Service) {
	mux.Handle("DELETE /api/v1/parent/child", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		child, ok := h.parentChild(w, r)
		if !ok {
			return
		}
		var input struct {
			Confirmation string `json:"confirmation"`
		}
		if decodeJSON(w, r, &input) != nil || input.Confirmation != "DELETE STUDENT" {
			tutoringError(w, tutor.ErrInvalid)
			return
		}
		if err := tutoring.CloseStudentVoice(r.Context(), child); err != nil {
			tutoringError(w, err)
			return
		}
		p, _ := PrincipalFromContext(r.Context())
		if err := service.DeleteSelected(r.Context(), p, child); err != nil {
			tutoringError(w, err)
			return
		}
		w.WriteHeader(204)
	})))
}
func (h *CoreHandler) parentChild(w http.ResponseWriter, r *http.Request) (string, bool) {
	p, _ := PrincipalFromContext(r.Context())
	if p.Role != identity.RoleParent {
		writeError(w, 403, "forbidden", "Espace parent uniquement.")
		return "", false
	}
	child, err := h.sessions.SelectChild(r.Context(), p, r.URL.Query().Get("child"))
	if err != nil {
		if errors.Is(err, identity.ErrInvalidCredentials) {
			writeError(w, 403, "forbidden", "Profil enfant inaccessible.")
			return "", false
		}
		coreError(w, err)
		return "", false
	}
	return child, true
}
func tutoringError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tutor.ErrInvalid):
		writeError(w, 400, "invalid_request", "Paramètres du professeur invalides.")
	case errors.Is(err, tutor.ErrBudget):
		writeError(w, 429, "budget", "La limite quotidienne de budget réservé ou de durée vocale est atteinte. Les réservations ne sont pas la facture réelle. Réessaie après 00:00 UTC, ou demande à ton parent de revoir le plafond serveur.")
	case errors.Is(err, tutor.ErrVoiceDisabled):
		writeError(w, 503, "voice_disabled", "Conversation micro non configurée sur le serveur. Un administrateur doit activer le fournisseur vocal. Le mode texte et la lecture d’Alex restent disponibles.")
	case errors.Is(err, tutor.ErrAIConsent):
		writeError(w, 403, "ai_consent_required", "La conversation micro nécessite l’autorisation IA du parent dans l’espace parent. Le mode texte reste disponible.")
	case errors.Is(err, tutor.ErrUnavailable):
		writeError(w, 503, "unavailable", "Conversation vocale indisponible. Le mode texte reste disponible.")
	default:
		coreError(w, err)
	}
}
