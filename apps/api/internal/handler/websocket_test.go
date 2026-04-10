package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

func TestWebSocketHandler_MissingToken(t *testing.T) {
	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			experimentGVR: "ChaosExperimentList",
			policyGVR:     "BlastRadiusPolicyList",
		},
	)
	k8s := service.NewK8sClientFromDynamic(client)
	svc := service.NewExperimentService(k8s)
	h := NewWebSocketHandler(svc)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/ws/experiments/:name", h.ExperimentStatus)
	req := httptest.NewRequest(http.MethodGet, "/ws/experiments/test-exp", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestWebSocketHandler_ConnectAndReceive(t *testing.T) {
	scheme := runtime.NewScheme()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			experimentGVR: "ChaosExperimentList",
			policyGVR:     "BlastRadiusPolicyList",
		},
	)

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos.chaosplane.io/v1alpha1",
			"kind":       "ChaosExperiment",
			"metadata": map[string]interface{}{
				"name":      "ws-test",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"action": map[string]interface{}{"type": "pod-kill"},
			},
			"status": map[string]interface{}{
				"phase": "Completed",
			},
		},
	}
	_, err := client.Resource(experimentGVR).Namespace("default").Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	k8s := service.NewK8sClientFromDynamic(client)
	svc := service.NewExperimentService(k8s)
	h := NewWebSocketHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws/experiments/:name", h.ExperimentStatus)

	srv := httptest.NewServer(r)
	defer srv.Close()

	wsURL := "ws" + srv.URL[4:] + "/ws/experiments/ws-test?token=test-key&namespace=default"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	_, msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("websocket read failed: %v", err)
	}
	if len(msg) == 0 {
		t.Error("expected non-empty message")
	}
}

func TestConnectionLimiter(t *testing.T) {
	cl := NewConnectionLimiter()

	for i := 0; i < maxSubscriptionsPerConn; i++ {
		if !cl.Acquire("key1") {
			t.Fatalf("expected acquire to succeed at %d", i)
		}
	}

	if cl.Acquire("key1") {
		t.Error("expected acquire to fail at max")
	}

	cl.Release("key1")
	if !cl.Acquire("key1") {
		t.Error("expected acquire to succeed after release")
	}

	if !cl.Acquire("key2") {
		t.Error("expected acquire for different key to succeed")
	}
}
