package service

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type PolicyService struct {
	k8s *K8sClient
}

func NewPolicyService(k8s *K8sClient) *PolicyService {
	return &PolicyService{k8s: k8s}
}

type PolicyResponse struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Enforcement string `json:"enforcement"`
	MaxTargets  *int64 `json:"maxTargets,omitempty"`
}

type PaginatedPolicyResponse struct {
	Items      []PolicyResponse `json:"items"`
	TotalCount int              `json:"totalCount"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
}

func (s *PolicyService) Get(ctx context.Context, namespace, name string) (*PolicyResponse, error) {
	obj, err := s.k8s.GetPolicy(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	return toPolicyResponse(obj), nil
}

func (s *PolicyService) List(ctx context.Context, namespace string, limit, offset int) (*PaginatedPolicyResponse, error) {
	list, err := s.k8s.ListPolicies(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	total := len(list.Items)
	end := offset + limit
	if end > total {
		end = total
	}

	var items []PolicyResponse
	if offset < total {
		for _, item := range list.Items[offset:end] {
			items = append(items, *toPolicyResponse(&item))
		}
	}
	if items == nil {
		items = []PolicyResponse{}
	}

	return &PaginatedPolicyResponse{
		Items:      items,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func toPolicyResponse(obj *unstructured.Unstructured) *PolicyResponse {
	resp := &PolicyResponse{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	if enforcement, ok, _ := unstructured.NestedString(obj.Object, "spec", "enforcement"); ok {
		resp.Enforcement = enforcement
	}
	if maxTargets, ok, _ := unstructured.NestedInt64(obj.Object, "spec", "targetLimits", "maxTargets"); ok {
		resp.MaxTargets = &maxTargets
	}
	return resp
}
