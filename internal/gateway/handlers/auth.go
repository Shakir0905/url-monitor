package handlers

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/status"

	authpb "github.com/shakir/url-monitor/proto/auth"
)

type AuthHandler struct {
	client authpb.AuthServiceClient
}

func NewAuthHandler(c authpb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{client: c}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.Register(r.Context(), &authpb.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user_id": resp.GetUserId(),
		"token":   resp.GetToken(),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.Login(r.Context(), &authpb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": resp.GetUserId(),
		"token":   resp.GetToken(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeGRPCError(w http.ResponseWriter, err error) {
	st, _ := status.FromError(err)
	httpStatus := grpcToHTTPCode(st.Code().String())
	writeJSON(w, httpStatus, map[string]string{"error": st.Message()})
}

func grpcToHTTPCode(code string) int {
	switch code {
	case "NotFound":
		return http.StatusNotFound
	case "AlreadyExists":
		return http.StatusConflict
	case "InvalidArgument":
		return http.StatusBadRequest
	case "Unauthenticated":
		return http.StatusUnauthorized
	case "PermissionDenied":
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
