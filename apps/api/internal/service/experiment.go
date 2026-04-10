package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ExperimentService struct {
	k8s *K8sClient
}

func NewExperimentService(k8s *K8sClient) *ExperimentService {
	return &ExperimentService{k8s: k8s}
}

type CreateExperimentRequest struct {
	Name      string        `json:"name" binding:"required"`
	Namespace string        `json:"namespace" binding:"required"`
	Action    ActionRequest `json:"action" binding:"required"`
	Target    TargetRequest `json:"target" binding:"required"`
	Duration  string        `json:"duration" binding:"required"`
}

type ActionRequest struct {
	Type       string          `json:"type" binding:"required"`
	Parameters json.RawMessage `json:"parameters,omitempty"`
}

type TargetRequest struct {
	Kind          string            `json:"kind" binding:"required"`
	Namespace     string            `json:"namespace,omitempty"`
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
	Names         []string          `json:"names,omitempty"`
}

type ExperimentResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Action    string `json:"action"`
	Phase     string `json:"phase"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

type PaginatedResponse struct {
	Items      []ExperimentResponse `json:"items"`
	TotalCount int                  `json:"totalCount"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
}

func (s *ExperimentService) Create(ctx context.Context, req *CreateExperimentRequest) (*ExperimentResponse, error) {
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		return nil, fmt.Errorf("invalid duration %q: %w", req.Duration, err)
	}

	spec := map[string]interface{}{
		"target":   buildTargetSpec(req.Target),
		"action":   buildActionSpec(req.Action),
		"duration": fmt.Sprintf("%.0fs", duration.Seconds()),
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "chaos.chaosplane.io/v1alpha1",
			"kind":       "ChaosExperiment",
			"metadata": map[string]interface{}{
				"name":      req.Name,
				"namespace": req.Namespace,
			},
			"spec": spec,
		},
	}

	created, err := s.k8s.CreateExperiment(ctx, req.Namespace, obj)
	if err != nil {
		return nil, fmt.Errorf("k8s create failed: %w", err)
	}

	return toExperimentResponse(created), nil
}

func (s *ExperimentService) Get(ctx context.Context, namespace, name string) (*ExperimentResponse, error) {
	obj, err := s.k8s.GetExperiment(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	return toExperimentResponse(obj), nil
}

func (s *ExperimentService) List(ctx context.Context, namespace string, limit, offset int) (*PaginatedResponse, error) {
	list, err := s.k8s.ListExperiments(ctx, namespace, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	total := len(list.Items)
	end := offset + limit
	if end > total {
		end = total
	}

	var items []ExperimentResponse
	if offset < total {
		for _, item := range list.Items[offset:end] {
			items = append(items, *toExperimentResponse(&item))
		}
	}
	if items == nil {
		items = []ExperimentResponse{}
	}

	return &PaginatedResponse{
		Items:      items,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *ExperimentService) Delete(ctx context.Context, namespace, name string) error {
	return s.k8s.DeleteExperiment(ctx, namespace, name)
}

func (s *ExperimentService) Abort(ctx context.Context, namespace, name string) (*ExperimentResponse, error) {
	obj, err := s.k8s.AbortExperiment(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	return toExperimentResponse(obj), nil
}

func buildTargetSpec(t TargetRequest) map[string]interface{} {
	target := map[string]interface{}{
		"kind": t.Kind,
	}
	if t.Namespace != "" {
		target["namespace"] = t.Namespace
	}
	if len(t.LabelSelector) > 0 {
		matchLabels := make(map[string]interface{}, len(t.LabelSelector))
		for k, v := range t.LabelSelector {
			matchLabels[k] = v
		}
		target["labelSelector"] = map[string]interface{}{
			"matchLabels": matchLabels,
		}
	}
	if len(t.Names) > 0 {
		names := make([]interface{}, len(t.Names))
		for i, n := range t.Names {
			names[i] = n
		}
		target["names"] = names
	}
	return target
}

func buildActionSpec(a ActionRequest) map[string]interface{} {
	action := map[string]interface{}{
		"type": a.Type,
	}
	if len(a.Parameters) > 0 {
		var params interface{}
		if json.Unmarshal(a.Parameters, &params) == nil {
			action["parameters"] = params
		}
	}
	return action
}

func toExperimentResponse(obj *unstructured.Unstructured) *ExperimentResponse {
	resp := &ExperimentResponse{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	if action, ok, _ := unstructured.NestedString(obj.Object, "spec", "action", "type"); ok {
		resp.Action = action
	}
	if phase, ok, _ := unstructured.NestedString(obj.Object, "status", "phase"); ok {
		resp.Phase = phase
	}
	if resp.Phase == "" {
		resp.Phase = "Pending"
	}
	if startTime, ok, _ := unstructured.NestedString(obj.Object, "status", "startTime"); ok {
		resp.StartTime = startTime
	}
	if endTime, ok, _ := unstructured.NestedString(obj.Object, "status", "endTime"); ok {
		resp.EndTime = endTime
	}

	return resp
}
