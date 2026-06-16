package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type ExperimentHandler struct {
	svc          *service.ExperimentService
	notification *service.NotificationService
	billing      *service.BillingService
}

func NewExperimentHandler(svc *service.ExperimentService, notification *service.NotificationService, billing *service.BillingService) *ExperimentHandler {
	return &ExperimentHandler{svc: svc, notification: notification, billing: billing}
}

func (h *ExperimentHandler) Create(c *gin.Context) {
	var req service.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeExperimentError(c, err)
		return
	}

	if tenantID, ok := c.Get("tenant_id"); ok {
		tid := tenantID.(string)
		if h.billing != nil {
			go h.billing.RecordUsage(context.Background(), tid, "experiments", 1)
		}
		if h.notification != nil {
			go h.notification.Dispatch(context.Background(), tid, "experiment.created", map[string]interface{}{
				"name": resp.Name, "namespace": resp.Namespace, "action": resp.Action.Type,
			})
		}
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *ExperimentHandler) List(c *gin.Context) {
	limit, offset := parsePagination(c)
	statusFilter := c.Query("status")
	actionFilter := c.Query("action")

	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c), statusFilter, actionFilter, limit, offset)
	if err != nil {
		writeExperimentError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ExperimentHandler) Get(c *gin.Context) {
	resp, err := h.svc.Get(c.Request.Context(), actorFromContext(c), c.Param("name"))
	if err != nil {
		writeExperimentError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ExperimentHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), actorFromContext(c), c.Param("name")); err != nil {
		writeExperimentError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ExperimentHandler) Abort(c *gin.Context) {
	resp, err := h.svc.Abort(c.Request.Context(), actorFromContext(c), c.Param("name"))
	if err != nil {
		writeExperimentError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ExperimentHandler) FaultCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"groups": service.FaultCatalog()})
}

func writeExperimentError(c *gin.Context, err error) {
	var verr *service.ValidationError
	if errors.As(err, &verr) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, service.ErrExperimentNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func parsePagination(c *gin.Context) (limit, offset int) {
	limit = 20
	offset = 0

	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 100 {
		limit = 100
	}

	if v := c.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
