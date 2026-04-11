package handler

import (
	"context"
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

	resp, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tenantID, ok := c.Get("tenant_id"); ok {
		tid := tenantID.(string)
		if h.billing != nil {
			go h.billing.RecordUsage(context.Background(), tid, "experiments", 1)
		}
		if h.notification != nil {
			go h.notification.Dispatch(context.Background(), tid, "experiment.created", map[string]interface{}{
				"name": resp.Name, "namespace": resp.Namespace, "action": resp.Action,
			})
		}
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *ExperimentHandler) List(c *gin.Context) {
	limit, offset := parsePagination(c)
	namespace := c.Query("namespace")

	resp, err := h.svc.List(c.Request.Context(), namespace, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ExperimentHandler) Get(c *gin.Context) {
	name := c.Param("name")
	namespace := c.DefaultQuery("namespace", "default")

	resp, err := h.svc.Get(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ExperimentHandler) Delete(c *gin.Context) {
	name := c.Param("name")
	namespace := c.DefaultQuery("namespace", "default")

	if err := h.svc.Delete(c.Request.Context(), namespace, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ExperimentHandler) Abort(c *gin.Context) {
	name := c.Param("name")
	namespace := c.DefaultQuery("namespace", "default")

	resp, err := h.svc.Abort(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
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
