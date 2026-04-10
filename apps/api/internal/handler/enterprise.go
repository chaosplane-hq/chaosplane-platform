package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

type EnterpriseHandler struct {
	saml     *service.SAMLService
	abac     *service.ABACService
	mfa      *service.MFAService
	sessions *service.SessionManagementService
	deletion *service.AccountDeletionService
	emailChg *service.EmailChangeService
	auditExp *service.AuditExportService
}

func NewEnterpriseHandler(
	saml *service.SAMLService,
	abac *service.ABACService,
	mfa *service.MFAService,
	sessions *service.SessionManagementService,
	deletion *service.AccountDeletionService,
	emailChg *service.EmailChangeService,
	auditExp *service.AuditExportService,
) *EnterpriseHandler {
	return &EnterpriseHandler{saml: saml, abac: abac, mfa: mfa, sessions: sessions, deletion: deletion, emailChg: emailChg, auditExp: auditExp}
}

func (h *EnterpriseHandler) ListSAMLProviders(c *gin.Context) {
	resp, err := h.saml.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnterpriseHandler) CreateSAMLProvider(c *gin.Context) {
	var req service.CreateSAMLProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.saml.Create(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *EnterpriseHandler) DeleteSAMLProvider(c *gin.Context) {
	if err := h.saml.Delete(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseHandler) ListABACPolicies(c *gin.Context) {
	resp, err := h.abac.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnterpriseHandler) CreateABACPolicy(c *gin.Context) {
	var req service.CreateABACPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.abac.Create(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *EnterpriseHandler) DeleteABACPolicy(c *gin.Context) {
	if err := h.abac.Delete(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseHandler) EvaluateABAC(c *gin.Context) {
	var req service.EvaluateABACRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.abac.Evaluate(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnterpriseHandler) GenerateRecoveryCodes(c *gin.Context) {
	resp, err := h.mfa.GenerateRecoveryCodes(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnterpriseHandler) GetRecoveryCodeCount(c *gin.Context) {
	count, err := h.mfa.GetRemainingCount(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"remaining": count})
}

func (h *EnterpriseHandler) ListSessions(c *gin.Context) {
	resp, err := h.sessions.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnterpriseHandler) RevokeSession(c *gin.Context) {
	if err := h.sessions.Revoke(c.Request.Context(), actorFromContext(c), c.Param("id")); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseHandler) RevokeAllSessions(c *gin.Context) {
	count, err := h.sessions.RevokeAll(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"revoked": count})
}

func (h *EnterpriseHandler) RequestDeletion(c *gin.Context) {
	var req service.RequestDeletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.deletion.Request(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnterpriseHandler) CancelDeletion(c *gin.Context) {
	if err := h.deletion.Cancel(c.Request.Context(), actorFromContext(c)); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseHandler) RequestEmailChange(c *gin.Context) {
	var req service.EmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.emailChg.Request(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnterpriseHandler) ConfirmEmailChange(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.emailChg.Confirm(c.Request.Context(), req.Token); err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EnterpriseHandler) ListAuditExports(c *gin.Context) {
	resp, err := h.auditExp.List(c.Request.Context(), actorFromContext(c))
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EnterpriseHandler) CreateAuditExport(c *gin.Context) {
	var req service.CreateAuditExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.auditExp.Create(c.Request.Context(), actorFromContext(c), &req)
	if err != nil {
		writeHierarchyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}
