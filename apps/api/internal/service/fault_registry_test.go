package service

import (
	"encoding/json"
	"testing"
)

// expectedGroups mirrors apps/web/src/lib/types.ts ACTION_TYPE_GROUPS so a
// drift between the frontend catalog and backend registry fails the build.
var expectedGroups = map[string][]string{
	"Pod":     {"pod-kill", "container-kill", "pod-cpu-stress", "pod-memory-stress", "pod-io-stress", "pod-dns-error", "pod-http-abort", "pod-http-delay"},
	"Network": {"network-delay", "network-loss", "network-corrupt", "network-duplicate", "network-partition", "network-bandwidth"},
	"Node":    {"node-drain", "node-taint", "node-restart", "node-cpu-stress"},
	"Stress":  {"stress-cpu", "stress-memory"},
	"eBPF":    {"ebpf-network-delay", "ebpf-network-loss", "ebpf-dns-chaos"},
	"AWS":     {"aws-ec2-stop", "aws-ec2-terminate", "aws-rds-failover", "aws-ecs-stop-task", "aws-az-failure"},
	"Azure":   {"azure-vm-stop", "azure-aks-scale", "azure-cosmosdb-failover"},
	"GCP":     {"gcp-gke-scale", "gcp-cloudsql-failover", "gcp-cloudrun-stop"},
	"VM":      {"vm-cpu-stress", "vm-memory-stress", "vm-disk-stress", "vm-network-delay", "vm-process-kill", "vm-process-suspend"},
}

func TestFaultCatalogMirrorsFrontend(t *testing.T) {
	catalog := FaultCatalog()

	gotGroups := map[string][]string{}
	for _, g := range catalog {
		for _, ft := range g.Types {
			gotGroups[g.Group] = append(gotGroups[g.Group], ft.Type)
			if ft.Group != g.Group {
				t.Errorf("fault %q has group %q but is listed under %q", ft.Type, ft.Group, g.Group)
			}
		}
	}

	for group, types := range expectedGroups {
		got := gotGroups[group]
		if len(got) != len(types) {
			t.Errorf("group %q: expected %d types, got %d (%v)", group, len(types), len(got), got)
			continue
		}
		for i, want := range types {
			if got[i] != want {
				t.Errorf("group %q index %d: expected %q, got %q", group, i, want, got[i])
			}
		}
	}

	for group := range gotGroups {
		if _, ok := expectedGroups[group]; !ok {
			t.Errorf("registry has unexpected group %q not present in frontend catalog", group)
		}
	}
}

func TestLookupFaultType(t *testing.T) {
	if _, ok := LookupFaultType("pod-kill"); !ok {
		t.Fatal("expected pod-kill to be a known fault type")
	}
	if _, ok := LookupFaultType("totally-made-up"); ok {
		t.Fatal("expected unknown fault type to be rejected")
	}
}

func TestValidateActionUnknownType(t *testing.T) {
	err := ValidateAction("not-a-real-fault", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for unknown fault type")
	}
	var verr *ValidationError
	if !asValidationError(err, &verr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestValidateActionMissingRequiredParam(t *testing.T) {
	// container-kill requires containerName.
	err := ValidateAction("container-kill", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error when required param containerName is missing")
	}
	var verr *ValidationError
	if !asValidationError(err, &verr) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
}

func TestValidateActionEmptyRequiredParamRejected(t *testing.T) {
	// An explicit empty string must count as missing.
	err := ValidateAction("container-kill", json.RawMessage(`{"containerName":""}`))
	if err == nil {
		t.Fatal("expected error when required param is empty string")
	}
}

func TestValidateActionValidWithRequiredParam(t *testing.T) {
	if err := ValidateAction("container-kill", json.RawMessage(`{"containerName":"app"}`)); err != nil {
		t.Fatalf("expected valid container-kill, got %v", err)
	}
}

func TestValidateActionNoParamsRequired(t *testing.T) {
	// pod-kill takes no required params; empty/nil params must pass.
	if err := ValidateAction("pod-kill", json.RawMessage(`{}`)); err != nil {
		t.Fatalf("expected pod-kill with empty params to be valid, got %v", err)
	}
	if err := ValidateAction("pod-kill", nil); err != nil {
		t.Fatalf("expected pod-kill with nil params to be valid, got %v", err)
	}
}

func TestValidateActionOptionalParamsOmittable(t *testing.T) {
	// network-delay params all have frontend defaults => optional.
	if err := ValidateAction("network-delay", json.RawMessage(`{}`)); err != nil {
		t.Fatalf("expected network-delay with no params to be valid, got %v", err)
	}
}

func TestValidateActionMultiRequiredParams(t *testing.T) {
	// aws-ecs-stop-task requires cluster and taskId; region is optional.
	if err := ValidateAction("aws-ecs-stop-task", json.RawMessage(`{"cluster":"c"}`)); err == nil {
		t.Fatal("expected error when taskId missing")
	}
	if err := ValidateAction("aws-ecs-stop-task", json.RawMessage(`{"cluster":"c","taskId":"t"}`)); err != nil {
		t.Fatalf("expected valid aws-ecs-stop-task, got %v", err)
	}
}

func TestValidateActionAcceptsNonStringValues(t *testing.T) {
	// Params may arrive as numbers/bools, not just strings.
	if err := ValidateAction("azure-aks-scale", json.RawMessage(`{"resourceGroup":"rg","clusterName":"c","nodeCount":3}`)); err != nil {
		t.Fatalf("expected numeric param to be accepted, got %v", err)
	}
}
