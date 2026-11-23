package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	identity "github.com/youenchene/EnglishCraft/internal/identity/application"
	practice "github.com/youenchene/EnglishCraft/internal/learning/application/practice"
)

type Handler struct{ practice *practice.Service }

func NewHandler(service *practice.Service) *Handler { return &Handler{practice: service} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/practice/irregular-verbs/next", h.nextExercise)
	mux.HandleFunc("POST /api/v1/practice/irregular-verbs/{verbID}/answers", h.submitAnswer)
}

type exerciseResponse struct {
	VerbID      string `json:"verbId"`
	BaseForm    string `json:"baseForm"`
	Translation string `json:"translation"`
	SkillID     string `json:"skillId"`
}

func (h *Handler) nextExercise(response http.ResponseWriter, request *http.Request) {
	principal, ok := PrincipalFromContext(request.Context())
	if !ok || principal.Role != identity.RoleChild {
		writeError(response, http.StatusForbidden, "forbidden", "This activity is available to the child account.")
		return
	}
	exercise, err := h.practice.NextExercise(request.Context(), principal.SubjectID)
	if err != nil {
		writeServiceError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, exerciseResponse(exercise))
}

type answerRequest struct {
	PastSimple     string `json:"pastSimple"`
	PastParticiple string `json:"pastParticiple"`
}

type answerResponse struct {
	Correct                bool      `json:"correct"`
	PastSimpleCorrect      bool      `json:"pastSimpleCorrect"`
	PastParticipleCorrect  bool      `json:"pastParticipleCorrect"`
	ExpectedPastSimple     string    `json:"expectedPastSimple"`
	ExpectedPastParticiple string    `json:"expectedPastParticiple"`
	Feedback               string    `json:"feedback"`
	ValidationLevel        string    `json:"validationLevel"`
	CorrectAnswers         int       `json:"correctAnswers"`
	TotalAnswers           int       `json:"totalAnswers"`
	NextReviewAt           time.Time `json:"nextReviewAt"`
}

func (h *Handler) submitAnswer(response http.ResponseWriter, request *http.Request) {
	principal, ok := PrincipalFromContext(request.Context())
	if !ok || principal.Role != identity.RoleChild {
		writeError(response, http.StatusForbidden, "forbidden", "This activity is available to the child account.")
		return
	}
	var input answerRequest
	if err := decodeJSON(response, request, &input); err != nil {
		writeError(response, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	result, err := h.practice.SubmitAnswer(request.Context(), practice.SubmitAnswerCommand{
		StudentID:      principal.SubjectID,
		VerbID:         request.PathValue("verbID"),
		PastSimple:     input.PastSimple,
		PastParticiple: input.PastParticiple,
	})
	if err != nil {
		writeServiceError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, answerResponse{
		Correct:                result.Correct,
		PastSimpleCorrect:      result.PastSimpleCorrect,
		PastParticipleCorrect:  result.PastParticipleCorrect,
		ExpectedPastSimple:     result.ExpectedPastSimple,
		ExpectedPastParticiple: result.ExpectedPastParticiple,
		Feedback:               result.Feedback,
		ValidationLevel:        string(result.ValidationLevel),
		CorrectAnswers:         result.CorrectAnswers,
		TotalAnswers:           result.TotalAnswers,
		NextReviewAt:           result.NextReviewAt,
	})
}

func decodeJSON(response http.ResponseWriter, request *http.Request, destination any) error {
	request.Body = http.MaxBytesReader(response, request.Body, 16*1024)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("the request body must contain a single JSON object")
	}
	return nil
}

func writeServiceError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, practice.ErrInvalidCommand):
		writeError(response, http.StatusBadRequest, "invalid_request", "The submitted answer is incomplete.")
	case errors.Is(err, practice.ErrVerbNotFound):
		writeError(response, http.StatusNotFound, "exercise_not_found", "This exercise no longer exists.")
	default:
		writeError(response, http.StatusInternalServerError, "internal_error", "Something went wrong. Please try again.")
	}
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func writeError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func bearerToken(header string) string {
	prefix, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(prefix, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}
