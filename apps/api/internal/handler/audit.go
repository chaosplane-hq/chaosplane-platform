package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type AuditHandler struct {
	svc *service.AuditService
}

func NewAuditHandler(svc *service.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(c *gin.Context) {
	limit := 50
	offset := 0
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	if v := c.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c), &service.ListAuditLogsRequest{
		Action:       c.Query("action"),
		ResourceType: c.Query("resourceType"),
		UserID:       c.Query("userId"),
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
