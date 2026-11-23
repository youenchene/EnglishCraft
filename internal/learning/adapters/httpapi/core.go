package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
	"github.com/youenchene/EnglishCraft/internal/learning/application/core"
)

type CoreHandler struct {
	service  *core.Service
	sessions *identity.Sessions
	secure   bool
}

func NewCore(service *core.Service, sessions *identity.Sessions, secure bool) *CoreHandler {
	return &CoreHandler{service: service, sessions: sessions, secure: secure}
}
func (h *CoreHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
	mux.Handle("GET /api/v1/auth/me", h.auth(false, http.HandlerFunc(h.me)))
	mux.Handle("GET /api/v1/skills", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, h.service.Skills()) })))
	mux.Handle("GET /api/v1/verbs", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, h.service.Verbs()) })))
	mux.Handle("GET /api/v1/missions", h.auth(false, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, h.service.Missions()) })))
	mux.Handle("POST /api/v1/missions/{id}/start", h.auth(true, http.HandlerFunc(h.start)))
	mux.Handle("GET /api/v1/missions/{id}/current", h.auth(true, http.HandlerFunc(h.current)))
	mux.Handle("POST /api/v1/answers", h.auth(true, http.HandlerFunc(h.answer)))
	mux.Handle("GET /api/v1/training", h.auth(true, http.HandlerFunc(h.training)))
	mux.Handle("GET /api/v1/progress", h.auth(true, http.HandlerFunc(h.progress)))
	mux.Handle("GET /api/v1/reviews", h.auth(true, http.HandlerFunc(h.progress)))
	mux.Handle("GET /api/v1/parent/results", h.auth(false, http.HandlerFunc(h.parent)))
	mux.Handle("GET /api/v1/parent/children", h.auth(false, http.HandlerFunc(h.children)))
	mux.Handle("POST /api/v1/parent/children", h.auth(false, http.HandlerFunc(h.addChild)))
}

// SameSite plus an Origin check protects cookie-authenticated mutations,
// including login CSRF. Reverse proxies must preserve the original Host.
func SameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" && r.Method != "HEAD" {
			origin := r.Header.Get("Origin")
			if origin != "" {
				u, err := url.Parse(origin)
				if err != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
					writeError(w, 403, "forbidden", "Cross-origin request rejected.")
					return
				}
			}
			if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				writeError(w, 403, "forbidden", "Cross-site request rejected.")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (h *CoreHandler) auth(child bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("englishcraft_session")
		if err != nil {
			writeError(w, 401, "unauthorized", "Connecte-toi pour continuer.")
			return
		}
		p, err := h.sessions.Resolve(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, 401, "unauthorized", "Session expirée.")
			return
		}
		if child && p.Role != identity.RoleChild {
			writeError(w, 403, "forbidden", "Espace élève uniquement.")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalContextKey{}, p)))
	})
}
func (h *CoreHandler) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, 400, "invalid_request", "Requête invalide.")
		return
	}
	token, p, err := h.sessions.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		if !errors.Is(err, identity.ErrInvalidCredentials) {
			coreError(w, err)
			return
		}
		writeError(w, 401, "unauthorized", "Identifiants invalides ou compte temporairement verrouillé.")
		return
	}
	h.cookie(w, token, 12*60*60)
	writeJSON(w, 200, map[string]string{"id": p.SubjectID, "role": string(p.Role)})
}
func (h *CoreHandler) cookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: "englishcraft_session", Value: token, Path: "/", HttpOnly: true,
		Secure: h.secure, SameSite: http.SameSiteStrictMode, MaxAge: maxAge})
}
func (h *CoreHandler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("englishcraft_session"); err == nil {
		if err := h.sessions.Logout(r.Context(), cookie.Value); err != nil {
			coreError(w, err)
			return
		}
	}
	h.cookie(w, "", -1)
	w.WriteHeader(204)
}
func (h *CoreHandler) me(w http.ResponseWriter, r *http.Request) {
	p, _ := PrincipalFromContext(r.Context())
	writeJSON(w, 200, map[string]string{"id": p.SubjectID, "role": string(p.Role)})
}
func (h *CoreHandler) start(w http.ResponseWriter, r *http.Request) {
	p, _ := PrincipalFromContext(r.Context())
	result, err := h.service.Start(r.Context(), p.SubjectID, r.PathValue("id"))
	if err != nil {
		coreError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
func (h *CoreHandler) current(w http.ResponseWriter, r *http.Request) {
	p, _ := PrincipalFromContext(r.Context())
	result, err := h.service.Current(r.Context(), p.SubjectID, r.PathValue("id"))
	if err != nil {
		coreError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
func (h *CoreHandler) answer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MissionID  string `json:"missionId"`
		Step       int    `json:"step"`
		ExerciseID string `json:"exerciseId"`
		Answer     string `json:"answer"`
		Key        string `json:"idempotencyKey"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, 400, "invalid_request", "Réponse invalide.")
		return
	}
	p, _ := PrincipalFromContext(r.Context())
	result, err := h.service.Submit(r.Context(), core.Submit{Student: p.SubjectID, MissionID: input.MissionID,
		Step: input.Step, ExerciseID: input.ExerciseID, Answer: input.Answer, Key: input.Key})
	if err != nil {
		coreError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
func (h *CoreHandler) training(w http.ResponseWriter, r *http.Request) {
	p, _ := PrincipalFromContext(r.Context())
	result, err := h.service.Training(r.Context(), p.SubjectID)
	if err != nil {
		coreError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
func (h *CoreHandler) progress(w http.ResponseWriter, r *http.Request) {
	p, _ := PrincipalFromContext(r.Context())
	result, err := h.service.Progress(r.Context(), p.SubjectID)
	if err != nil {
		coreError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
func (h *CoreHandler) parent(w http.ResponseWriter, r *http.Request) {
	child, ok := h.parentChild(w, r)
	if !ok {
		return
	}
	result, err := h.service.Progress(r.Context(), child)
	if err != nil {
		coreError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
func coreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrInvalidCredentials):
		writeError(w, 403, "forbidden", "Espace parent uniquement.")
	case errors.Is(err, identity.ErrInvalidChild):
		writeError(w, 400, "invalid_request", "Identifiant de 3 à 80 caractères (lettres, chiffres, _, ., -) et mot de passe de 12 à 72 octets requis.")
	case errors.Is(err, identity.ErrUsernameTaken):
		writeError(w, 409, "conflict", "Cet identifiant est déjà utilisé.")
	case errors.Is(err, core.ErrNotFound):
		writeError(w, 404, "not_found", "Contenu indisponible.")
	case errors.Is(err, core.ErrInvalid):
		writeError(w, 400, "invalid_request", "Réponse invalide.")
	case errors.Is(err, core.ErrConflict):
		writeError(w, 409, "conflict", "Cette étape a changé. Recharge la mission.")
	default:
		writeError(w, 500, "internal_error", "Une erreur est survenue.")
	}
}
