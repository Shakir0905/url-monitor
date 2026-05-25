package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/shakir/url-monitor/internal/gateway/middleware"
	urlpb "github.com/shakir/url-monitor/proto/url"
)

type URLHandler struct {
	client urlpb.URLServiceClient
}

func NewURLHandler(c urlpb.URLServiceClient) *URLHandler {
	return &URLHandler{client: c}
}

type createURLRequest struct {
	URL                  string `json:"url"`
	CheckIntervalSeconds int32  `json:"check_interval_seconds"`
}

type updateURLRequest struct {
	URL                  string `json:"url"`
	CheckIntervalSeconds int32  `json:"check_interval_seconds"`
	IsActive             bool   `json:"is_active"`
}

func (h *URLHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	resp, err := h.client.ListURLs(r.Context(), &urlpb.ListURLsRequest{
		UserId: userID, Limit: 100, Offset: 0,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *URLHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var req createURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	interval := req.CheckIntervalSeconds
	if interval == 0 {
		interval = 60
	}
	resp, err := h.client.CreateURL(r.Context(), &urlpb.CreateURLRequest{
		UserId: userID, Url: req.URL, CheckIntervalSeconds: interval,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *URLHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	resp, err := h.client.GetURL(r.Context(), &urlpb.GetURLRequest{Id: id, UserId: userID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *URLHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req updateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.UpdateURL(r.Context(), &urlpb.UpdateURLRequest{
		Id: id, UserId: userID,
		Url: req.URL, CheckIntervalSeconds: req.CheckIntervalSeconds, IsActive: req.IsActive,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *URLHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	_, err = h.client.DeleteURL(r.Context(), &urlpb.DeleteURLRequest{Id: id, UserId: userID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}
