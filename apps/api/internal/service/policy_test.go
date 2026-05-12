package service

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestPolicyService_Get(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewPolicyService(k8s)

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
				"targetLimits": map[string]interface{}{
					"maxTargets": int64(5),
				},
			},
		},
	}
	_, err := k8s.client.Resource(policyGVR).Namespace("default").Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	resp, err := svc.Get(context.Background(), "default", "test-policy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "test-policy" {
		t.Errorf("expected name test-policy, got %s", resp.Name)
	}
	if resp.Enforcement != "Enforce" {
		t.Errorf("expected enforcement Enforce, got %s", resp.Enforcement)
	}
	if resp.MaxTargets == nil || *resp.MaxTargets != 5 {
		t.Errorf("expected maxTargets 5, got %v", resp.MaxTargets)
	}
}

func TestPolicyService_List(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewPolicyService(k8s)

	for _, name := range []string{"pol-a", "pol-b"} {
		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "chaos.chaosplane.dev/v1alpha1",
				"kind":       "BlastRadiusPolicy",
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"enforcement": "Audit",
				},
			},
		}
		_, _ = k8s.client.Resource(policyGVR).Namespace("default").Create(context.Background(), obj, metav1.CreateOptions{})
	}

	resp, err := svc.List(context.Background(), "default", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected totalCount 2, got %d", resp.TotalCount)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

func TestPolicyService_GetNotFound(t *testing.T) {
	k8s := newFakeK8sClient()
	svc := NewPolicyService(k8s)

	_, err := svc.Get(context.Background(), "default", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent policy, got nil")
	}
}
