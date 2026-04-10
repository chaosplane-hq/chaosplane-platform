package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
)

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

// K8sClient wraps the dynamic Kubernetes client for CRD operations.
type K8sClient struct {
	client dynamic.Interface
}

// NewK8sClient creates a K8s dynamic client. Uses in-cluster config when available,
// falls back to kubeconfig for local development.
func NewK8sClient(cfg *config.Config) (*K8sClient, error) {
	var restCfg *rest.Config
	var err error

	if cfg.Kubeconfig != "" {
		restCfg, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	} else {
		restCfg, err = rest.InClusterConfig()
		if err != nil {
			// Fallback: try default kubeconfig location
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			restCfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				loadingRules, configOverrides).ClientConfig()
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to build k8s config: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	slog.Info("k8s client initialized")
	return &K8sClient{client: dynClient}, nil
}

// NewK8sClientFromDynamic creates a K8sClient from an existing dynamic.Interface (for testing).
func NewK8sClientFromDynamic(client dynamic.Interface) *K8sClient {
	return &K8sClient{client: client}
}

// CreateExperiment creates a ChaosExperiment CRD in the given namespace.
func (k *K8sClient) CreateExperiment(ctx context.Context, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return k.client.Resource(experimentGVR).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
}

// GetExperiment retrieves a ChaosExperiment by name and namespace.
func (k *K8sClient) GetExperiment(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	return k.client.Resource(experimentGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListExperiments lists ChaosExperiments in the given namespace. Empty namespace lists across all namespaces.
func (k *K8sClient) ListExperiments(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	if namespace == "" {
		return k.client.Resource(experimentGVR).List(ctx, opts)
	}
	return k.client.Resource(experimentGVR).Namespace(namespace).List(ctx, opts)
}

// DeleteExperiment deletes a ChaosExperiment by name and namespace.
func (k *K8sClient) DeleteExperiment(ctx context.Context, namespace, name string) error {
	return k.client.Resource(experimentGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// PatchExperiment applies a JSON merge patch to a ChaosExperiment.
func (k *K8sClient) PatchExperiment(ctx context.Context, namespace, name string, patch []byte) (*unstructured.Unstructured, error) {
	return k.client.Resource(experimentGVR).Namespace(namespace).Patch(
		ctx, name, types.MergePatchType, patch, metav1.PatchOptions{},
	)
}

// AbortExperiment adds the abort annotation to a ChaosExperiment.
func (k *K8sClient) AbortExperiment(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	patch, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				"chaosplane.io/abort": "true",
			},
		},
	})
	return k.PatchExperiment(ctx, namespace, name, patch)
}

// GetPolicy retrieves a BlastRadiusPolicy by name and namespace.
func (k *K8sClient) GetPolicy(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	if namespace == "" {
		return k.client.Resource(policyGVR).Get(ctx, name, metav1.GetOptions{})
	}
	return k.client.Resource(policyGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListPolicies lists BlastRadiusPolicies in the given namespace. Empty namespace lists across all namespaces.
func (k *K8sClient) ListPolicies(ctx context.Context, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	if namespace == "" {
		return k.client.Resource(policyGVR).List(ctx, opts)
	}
	return k.client.Resource(policyGVR).Namespace(namespace).List(ctx, opts)
}
