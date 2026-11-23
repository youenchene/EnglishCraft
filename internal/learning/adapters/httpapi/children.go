package httpapi

import "net/http"

func (h *CoreHandler) children(w http.ResponseWriter, r *http.Request) {
	p, _ := PrincipalFromContext(r.Context())
	children, err := h.sessions.Children(r.Context(), p)
	if err != nil {
		coreError(w, err)
		return
	}
	writeJSON(w, 200, children)
}
func (h *CoreHandler) addChild(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if decodeJSON(w, r, &input) != nil {
		writeError(w, 400, "invalid_request", "Requête invalide.")
		return
	}
	p, _ := PrincipalFromContext(r.Context())
	child, err := h.sessions.AddChild(r.Context(), p, input.Username, input.Password)
	if err != nil {
		coreError(w, err)
		return
	}
	writeJSON(w, 201, child)
}
