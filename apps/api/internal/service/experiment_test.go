package service

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func newFakeK8sClient() *K8sClient {
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{
		Group:    "chaos.chaosplane.dev",
		Version:  "v1alpha1",
		Resource: "chaosexperiments",
	}
	_ = gvr
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			experimentGVR: "ChaosExperimentList",
			policyGVR:     "BlastRadiusPolicyList",
		},
	)
	return NewK8sClientFromDynamic(client)
}

func TestExperimentService_Create(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewExperimentService(k8s)

	req := &CreateExperimentRequest{
		Name:      "test-exp",
		Namespace: "default",
		Duration:  "30s",
		Action: ActionRequest{
			Type: "pod-kill",
		},
		Target: TargetRequest{
			Kind:  "Pod",
			Names: []string{"nginx-1"},
		},
	}

	resp, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "test-exp" {
		t.Errorf("expected name test-exp, got %s", resp.Name)
	}
	if resp.Namespace != "default" {
		t.Errorf("expected namespace default, got %s", resp.Namespace)
	}
	if resp.Action != "pod-kill" {
		t.Errorf("expected action pod-kill, got %s", resp.Action)
	}
	if resp.Phase != "Pending" {
		t.Errorf("expected phase Pending, got %s", resp.Phase)
	}
}

func TestExperimentService_Get(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewExperimentService(k8s)

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos.chaosplane.dev/v1alpha1",
			"kind":       "ChaosExperiment",
			"metadata": map[string]interface{}{
				"name":      "test-exp",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"action": map[string]interface{}{"type": "pod-kill"},
			},
			"status": map[string]interface{}{
				"phase": "Running",
			},
		},
	}
	_, err := k8s.CreateExperiment(context.Background(), "default", obj)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := svc.Get(context.Background(), "default", "test-exp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Phase != "Running" {
		t.Errorf("expected phase Running, got %s", resp.Phase)
	}
}

func TestExperimentService_List(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewExperimentService(k8s)

	for i := 0; i < 3; i++ {
		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "chaos.chaosplane.dev/v1alpha1",
				"kind":       "ChaosExperiment",
				"metadata": map[string]interface{}{
					"name":      "exp-" + string(rune('a'+i)),
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"action": map[string]interface{}{"type": "pod-kill"},
				},
			},
		}
		_, _ = k8s.CreateExperiment(context.Background(), "default", obj)
	}

	resp, err := svc.List(context.Background(), "default", 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalCount != 3 {
		t.Errorf("expected totalCount 3, got %d", resp.TotalCount)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Limit != 2 {
		t.Errorf("expected limit 2, got %d", resp.Limit)
	}
}

func TestExperimentService_ListPagination_OffsetBeyondTotal(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewExperimentService(k8s)

	resp, err := svc.List(context.Background(), "default", 20, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Items))
	}
}

func TestExperimentService_Delete(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewExperimentService(k8s)

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos.chaosplane.dev/v1alpha1",
			"kind":       "ChaosExperiment",
			"metadata": map[string]interface{}{
				"name":      "to-delete",
				"namespace": "default",
			},
			"spec": map[string]interface{}{},
		},
	}
	_, _ = k8s.CreateExperiment(context.Background(), "default", obj)

	err := svc.Delete(context.Background(), "default", "to-delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Get(context.Background(), "default", "to-delete")
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestExperimentService_Abort(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewExperimentService(k8s)

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos.chaosplane.dev/v1alpha1",
			"kind":       "ChaosExperiment",
			"metadata": map[string]interface{}{
				"name":      "to-abort",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"action": map[string]interface{}{"type": "pod-kill"},
			},
		},
	}
	_, _ = k8s.CreateExperiment(context.Background(), "default", obj)

	resp, err := svc.Abort(context.Background(), "default", "to-abort")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "to-abort" {
		t.Errorf("expected name to-abort, got %s", resp.Name)
	}

	updated, _ := k8s.GetExperiment(context.Background(), "default", "to-abort")
	annotations := updated.GetAnnotations()
	if annotations["chaosplane.dev/abort"] != "true" {
		t.Errorf("expected abort annotation, got %v", annotations)
	}
}

func TestExperimentService_CreateInvalidDuration(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewExperimentService(k8s)

	req := &CreateExperimentRequest{
		Name:      "bad-dur",
		Namespace: "default",
		Duration:  "not-a-duration",
		Action:    ActionRequest{Type: "pod-kill"},
		Target:    TargetRequest{Kind: "Pod"},
	}

	_, err := svc.Create(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid duration, got nil")
	}
}

func TestK8sClient_ListPolicies(t *testing.T) {
	k8s := newFakeK8sClient()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos.chaosplane.dev/v1alpha1",
			"kind":       "BlastRadiusPolicy",
			"metadata": map[string]interface{}{
				"name":      "test-policy",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"enforcement": "Enforce",
			},
		},
	}
	_, err := k8s.client.Resource(policyGVR).Namespace("default").Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	list, err := k8s.ListPolicies(context.Background(), "default", metav1.ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 1 {
		t.Errorf("expected 1 policy, got %d", len(list.Items))
	}
}
