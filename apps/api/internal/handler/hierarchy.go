package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type HierarchyHandler struct {
	svc *service.HierarchyService
}

func NewHierarchyHandler(svc *service.HierarchyService) *HierarchyHandler {
	return &HierarchyHandler{svc: svc}
}

func (h *HierarchyHandler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HierarchyHandler) CreateOrganization(c *gin.Context) {
	var req service.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.CreateOrganization(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *HierarchyHandler) UpdateOrganization(c *gin.Context) {
	var req service.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.UpdateOrganization(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HierarchyHandler) CreateWorkspace(c *gin.Context) {
	var req service.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.CreateWorkspace(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *HierarchyHandler) UpdateWorkspace(c *gin.Context) {
	var req service.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.UpdateWorkspace(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HierarchyHandler) CreateTeam(c *gin.Context) {
	var req service.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.CreateTeam(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *HierarchyHandler) UpdateTeam(c *gin.Context) {
	var req service.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.UpdateTeam(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HierarchyHandler) CreateProject(c *gin.Context) {
	var req service.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.CreateProject(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *HierarchyHandler) UpdateProject(c *gin.Context) {
	var req service.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.UpdateProject(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *HierarchyHandler) CreateEnvironment(c *gin.Context) {
	var req service.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.CreateEnvironment(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *HierarchyHandler) UpdateEnvironment(c *gin.Context) {
	var req service.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.UpdateEnvironment(c.Request.Context(), actorFromContext(c), c.Param("id"), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func actorFromContext(c *gin.Context) service.ActorContext {
	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")
	return service.ActorContext{UserID: userID.(string), TenantID: tenantID.(string)}
}

func writeHierarchyError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	switch err {
	case service.ErrHierarchyNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case service.ErrInvalidCredentials:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
