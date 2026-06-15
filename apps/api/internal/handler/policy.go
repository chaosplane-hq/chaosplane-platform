package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type PolicyHandler struct {
	svc *service.PolicyService
}

func NewPolicyHandler(svc *service.PolicyService) *PolicyHandler {
	return &PolicyHandler{svc: svc}
}

func (h *PolicyHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writePolicyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PolicyHandler) Get(c *gin.Context) {
	resp, err := h.svc.Get(c.Request.Context(), actorFromContext(c), c.Param("name"))
	if err != nil {
		writePolicyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PolicyHandler) Create(c *gin.Context) {
	var req service.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.Create(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writePolicyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *PolicyHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), actorFromContext(c), c.Param("name")); err != nil {
		writePolicyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writePolicyError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrPolicyNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
