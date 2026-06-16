package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

func newExperimentTestContext(t *testing.T, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/experiments", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "00000000-0000-0000-0000-000000000001")
	c.Set("tenant_id", "00000000-0000-0000-0000-000000000002")
	return c, w
}

// Create validates the request before any DB access, so a nil-pool service is
// sufficient to assert the input-rejection (400) contract.
func newValidationOnlyHandler() *ExperimentHandler {
	return NewExperimentHandler(service.NewExperimentService(nil), nil, nil)
}

func TestCreateRejectsUnknownFaultType(t *testing.T) {
	h := newValidationOnlyHandler()
	c, w := newExperimentTestContext(t, `{
		"name":"bad","namespace":"default",
		"action":{"type":"totally-not-real"},
		"target":{"namespace":"default"}
	}`)

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown fault type, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestCreateRejectsMissingRequiredParam(t *testing.T) {
	h := newValidationOnlyHandler()
	c, w := newExperimentTestContext(t, `{
		"name":"bad","namespace":"default",
		"action":{"type":"container-kill","parameters":{}},
		"target":{"namespace":"default"}
	}`)

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing required param, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestCreateRejectsCyclicWorkflow(t *testing.T) {
	h := newValidationOnlyHandler()
	c, w := newExperimentTestContext(t, `{
		"name":"cyclic","namespace":"default",
		"steps":[
			{"name":"a","dependsOn":["b"],"action":{"type":"pod-kill"},"target":{"namespace":"default"}},
			{"name":"b","dependsOn":["a"],"action":{"type":"node-drain"},"target":{"namespace":"default"}}
		]
	}`)

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for cyclic workflow, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestCreateRejectsBothActionAndSteps(t *testing.T) {
	h := newValidationOnlyHandler()
	c, w := newExperimentTestContext(t, `{
		"name":"both","namespace":"default",
		"action":{"type":"pod-kill"},
		"steps":[{"name":"a","action":{"type":"pod-kill"},"target":{"namespace":"default"}}]
	}`)

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when both action and steps provided, got %d", w.Code)
	}
}

func TestFaultCatalogReturnsGroups(t *testing.T) {
	h := newValidationOnlyHandler()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/fault-catalog", nil)

	h.FaultCatalog(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Groups []service.FaultGroup `json:"groups"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal catalog: %v", err)
	}
	if len(body.Groups) != 9 {
		t.Fatalf("expected 9 fault groups mirroring the frontend, got %d", len(body.Groups))
	}
}
