package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

var (
	experimentGVR = schema.GroupVersionResource{
		Group:    "chaos.chaosplane.io",
		Version:  "v1alpha1",
		Resource: "chaosexperiments",
	}
	policyGVR = schema.GroupVersionResource{
		Group:    "chaos.chaosplane.io",
		Version:  "v1alpha1",
		Resource: "blastradiuspolicies",
	}
)

func newTestExperimentHandler() (*ExperimentHandler, *service.ExperimentService) {
	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			experimentGVR: "ChaosExperimentList",
			policyGVR:     "BlastRadiusPolicyList",
		},
	)
	k8s := service.NewK8sClientFromDynamic(client)
	svc := service.NewExperimentService(k8s)
	return NewExperimentHandler(svc, nil, nil), svc
}

func TestExperimentHandler_Create(t *testing.T) {
	h, _ := newTestExperimentHandler()

	body := `{"name":"test-exp","namespace":"default","duration":"30s","action":{"type":"pod-kill"},"target":{"kind":"Pod","names":["nginx"]}}`
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/api/v1/experiments", h.Create)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/experiments", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp service.ExperimentResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Name != "test-exp" {
		t.Errorf("expected name test-exp, got %s", resp.Name)
	}
}

func TestExperimentHandler_Create_BadRequest(t *testing.T) {
	h, _ := newTestExperimentHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/v1/experiments", h.Create)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/experiments", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestExperimentHandler_List(t *testing.T) {
	h, svc := newTestExperimentHandler()

	_, _ = svc.Create(nil, &service.CreateExperimentRequest{
		Name: "exp-1", Namespace: "default", Duration: "10s",
		Action: service.ActionRequest{Type: "pod-kill"},
		Target: service.TargetRequest{Kind: "Pod"},
	})

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/v1/experiments", h.List)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/experiments?namespace=default", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp service.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.TotalCount != 1 {
		t.Errorf("expected totalCount 1, got %d", resp.TotalCount)
	}
}

func TestExperimentHandler_Get(t *testing.T) {
	h, svc := newTestExperimentHandler()

	ctx := context.Background()
	_, _ = svc.Create(ctx, &service.CreateExperimentRequest{
		Name: "exp-get", Namespace: "default", Duration: "10s",
		Action: service.ActionRequest{Type: "pod-kill"},
		Target: service.TargetRequest{Kind: "Pod"},
	})

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/v1/experiments/:name", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/experiments/exp-get?namespace=default", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestExperimentHandler_Delete(t *testing.T) {
	h, svc := newTestExperimentHandler()

	ctx := context.Background()
	_, _ = svc.Create(ctx, &service.CreateExperimentRequest{
		Name: "exp-del", Namespace: "default", Duration: "10s",
		Action: service.ActionRequest{Type: "pod-kill"},
		Target: service.TargetRequest{Kind: "Pod"},
	})

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.DELETE("/api/v1/experiments/:name", h.Delete)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/experiments/exp-del?namespace=default", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestExperimentHandler_Abort(t *testing.T) {
	h, svc := newTestExperimentHandler()

	ctx := context.Background()
	_, _ = svc.Create(ctx, &service.CreateExperimentRequest{
		Name: "exp-abort", Namespace: "default", Duration: "10s",
		Action: service.ActionRequest{Type: "pod-kill"},
		Target: service.TargetRequest{Kind: "Pod"},
	})

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/v1/experiments/:name/abort", h.Abort)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/experiments/exp-abort/abort?namespace=default", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestExperimentHandler_GetNotFound(t *testing.T) {
	h, _ := newTestExperimentHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/v1/experiments/:name", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/experiments/nonexistent?namespace=default", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		query      string
		wantLimit  int
		wantOffset int
	}{
		{"", 20, 0},
		{"limit=50&offset=10", 50, 10},
		{"limit=200", 100, 0},
		{"limit=-1", 20, 0},
		{"offset=-5", 20, 0},
		{"limit=abc", 20, 0},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)

		limit, offset := parsePagination(c)
		if limit != tt.wantLimit {
			t.Errorf("query=%q: expected limit %d, got %d", tt.query, tt.wantLimit, limit)
		}
		if offset != tt.wantOffset {
			t.Errorf("query=%q: expected offset %d, got %d", tt.query, tt.wantOffset, offset)
		}
	}
}

func newTestPolicyHandler() *PolicyHandler {
	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			experimentGVR: "ChaosExperimentList",
			policyGVR:     "BlastRadiusPolicyList",
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "chaos.chaosplane.io/v1alpha1",
				"kind":       "BlastRadiusPolicy",
				"metadata": map[string]interface{}{
					"name":      "test-policy",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"enforcement": "Enforce",
				},
			},
		},
	)
	k8s := service.NewK8sClientFromDynamic(client)
	svc := service.NewPolicyService(k8s)
	return NewPolicyHandler(svc)
}

func TestPolicyHandler_List(t *testing.T) {
	h := newTestPolicyHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/v1/policies", h.List)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/policies?namespace=default", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPolicyHandler_Get(t *testing.T) {
	h := newTestPolicyHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/v1/policies/:name", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/test-policy?namespace=default", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPolicyHandler_GetNotFound(t *testing.T) {
	h := newTestPolicyHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/v1/policies/:name", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/policies/nonexistent?namespace=default", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
