package handlers

import (
	"net/http"
	"strconv"

	"github.com/shakir/url-monitor/internal/gateway/middleware"
	analyticspb "github.com/shakir/url-monitor/proto/analytics"
)

type AnalyticsHandler struct {
	client analyticspb.AnalyticsServiceClient
}

func NewAnalyticsHandler(c analyticspb.AnalyticsServiceClient) *AnalyticsHandler {
	return &AnalyticsHandler{client: c}
}

func (h *AnalyticsHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	resp, err := h.client.GetUserDashboard(r.Context(), &analyticspb.GetUserDashboardRequest{UserId: userID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AnalyticsHandler) URLStats(w http.ResponseWriter, r *http.Request) {
	urlID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	resp, err := h.client.GetURLStats(r.Context(), &analyticspb.GetURLStatsRequest{UrlId: urlID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
