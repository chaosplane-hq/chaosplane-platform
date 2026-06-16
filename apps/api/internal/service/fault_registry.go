package service

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ValidationError is mapped to HTTP 400 by handlers, distinguishing bad input
// from internal (500) failures.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func newValidationError(format string, args ...any) *ValidationError {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

func asValidationError(err error, target **ValidationError) bool {
	return errors.As(err, target)
}

// ParamSpec.Required follows the frontend convention: a param without a
// default value in the web create form is a required input.
type ParamSpec struct {
	Key      string `json:"key"`
	Required bool   `json:"required"`
}

type FaultType struct {
	Type   string      `json:"type"`
	Group  string      `json:"group"`
	Params []ParamSpec `json:"params"`
}

type FaultGroup struct {
	Group string      `json:"group"`
	Types []FaultType `json:"types"`
}

func req(key string) ParamSpec { return ParamSpec{Key: key, Required: true} }
func opt(key string) ParamSpec { return ParamSpec{Key: key, Required: false} }

// faultGroups is the canonical fault-type registry. It is the single source of
// truth for the API and mirrors apps/web/src/lib/types.ts ACTION_TYPE_GROUPS
// (group order + membership) and the per-type params in the web create form
// (apps/web .../experiments/create, where a param lacking a default is required).
var faultGroups = []FaultGroup{
	{Group: "Pod", Types: []FaultType{
		{Type: "pod-kill"},
		{Type: "container-kill", Params: []ParamSpec{req("containerName")}},
		{Type: "pod-cpu-stress", Params: []ParamSpec{opt("workers"), opt("load")}},
		{Type: "pod-memory-stress", Params: []ParamSpec{opt("workers"), opt("size")}},
		{Type: "pod-io-stress", Params: []ParamSpec{opt("workers"), req("volumePath")}},
		{Type: "pod-dns-error", Params: []ParamSpec{req("patterns")}},
		{Type: "pod-http-abort", Params: []ParamSpec{opt("port"), opt("path"), opt("method")}},
		{Type: "pod-http-delay", Params: []ParamSpec{opt("port"), opt("path"), opt("delay"), opt("method")}},
	}},
	{Group: "Network", Types: []FaultType{
		{Type: "network-delay", Params: []ParamSpec{opt("latency"), opt("jitter"), opt("correlation")}},
		{Type: "network-loss", Params: []ParamSpec{opt("loss"), opt("correlation")}},
		{Type: "network-corrupt", Params: []ParamSpec{opt("corrupt"), opt("correlation")}},
		{Type: "network-duplicate", Params: []ParamSpec{opt("duplicate"), opt("correlation")}},
		{Type: "network-partition", Params: []ParamSpec{opt("direction")}},
		{Type: "network-bandwidth", Params: []ParamSpec{opt("rate"), opt("limit"), opt("buffer")}},
	}},
	{Group: "Node", Types: []FaultType{
		{Type: "node-drain"},
		{Type: "node-taint", Params: []ParamSpec{req("key"), req("value"), opt("effect")}},
		{Type: "node-restart"},
		{Type: "node-cpu-stress", Params: []ParamSpec{opt("workers"), opt("load")}},
	}},
	{Group: "Stress", Types: []FaultType{
		{Type: "stress-cpu", Params: []ParamSpec{opt("workers"), opt("load")}},
		{Type: "stress-memory", Params: []ParamSpec{opt("workers"), opt("size")}},
	}},
	{Group: "eBPF", Types: []FaultType{
		{Type: "ebpf-network-delay", Params: []ParamSpec{opt("latency"), opt("interface")}},
		{Type: "ebpf-network-loss", Params: []ParamSpec{opt("loss"), opt("interface")}},
		{Type: "ebpf-dns-chaos", Params: []ParamSpec{req("patterns"), opt("action")}},
	}},
	{Group: "AWS", Types: []FaultType{
		{Type: "aws-ec2-stop", Params: []ParamSpec{req("instanceId"), opt("region")}},
		{Type: "aws-ec2-terminate", Params: []ParamSpec{req("instanceId"), opt("region")}},
		{Type: "aws-rds-failover", Params: []ParamSpec{req("dbClusterIdentifier"), opt("region")}},
		{Type: "aws-ecs-stop-task", Params: []ParamSpec{req("cluster"), req("taskId"), opt("region")}},
		{Type: "aws-az-failure", Params: []ParamSpec{req("az"), opt("region")}},
	}},
	{Group: "Azure", Types: []FaultType{
		{Type: "azure-vm-stop", Params: []ParamSpec{req("resourceGroup"), req("vmName")}},
		{Type: "azure-aks-scale", Params: []ParamSpec{req("resourceGroup"), req("clusterName"), opt("nodeCount")}},
		{Type: "azure-cosmosdb-failover", Params: []ParamSpec{req("resourceGroup"), req("accountName"), req("failoverRegion")}},
	}},
	{Group: "GCP", Types: []FaultType{
		{Type: "gcp-gke-scale", Params: []ParamSpec{req("project"), req("cluster"), req("nodePool"), opt("nodeCount")}},
		{Type: "gcp-cloudsql-failover", Params: []ParamSpec{req("project"), req("instance")}},
		{Type: "gcp-cloudrun-stop", Params: []ParamSpec{req("project"), req("service"), opt("region")}},
	}},
	{Group: "VM", Types: []FaultType{
		{Type: "vm-cpu-stress", Params: []ParamSpec{opt("workers"), opt("load")}},
		{Type: "vm-memory-stress", Params: []ParamSpec{opt("workers"), opt("size")}},
		{Type: "vm-disk-stress", Params: []ParamSpec{opt("workers"), opt("path"), opt("size")}},
		{Type: "vm-network-delay", Params: []ParamSpec{opt("latency"), opt("interface")}},
		{Type: "vm-process-kill", Params: []ParamSpec{req("processName"), opt("signal")}},
		{Type: "vm-process-suspend", Params: []ParamSpec{req("processName")}},
	}},
}

var faultIndex = buildFaultIndex()

func buildFaultIndex() map[string]FaultType {
	idx := make(map[string]FaultType)
	for _, g := range faultGroups {
		for _, ft := range g.Types {
			ft.Group = g.Group
			idx[ft.Type] = ft
		}
	}
	return idx
}

func FaultCatalog() []FaultGroup {
	out := make([]FaultGroup, len(faultGroups))
	for i, g := range faultGroups {
		types := make([]FaultType, len(g.Types))
		for j, ft := range g.Types {
			ft.Group = g.Group
			types[j] = ft
		}
		out[i] = FaultGroup{Group: g.Group, Types: types}
	}
	return out
}

func LookupFaultType(faultType string) (FaultType, bool) {
	ft, ok := faultIndex[faultType]
	return ft, ok
}

// ValidateAction accepts param values as strings, numbers, or booleans because
// the frontend sends mixed types for a single parameter map.
func ValidateAction(faultType string, params json.RawMessage) error {
	ft, ok := faultIndex[faultType]
	if !ok {
		return newValidationError("unknown fault type %q", faultType)
	}

	parsed := map[string]json.RawMessage{}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &parsed); err != nil {
			return newValidationError("invalid parameters for fault %q: %v", faultType, err)
		}
	}

	for _, p := range ft.Params {
		if !p.Required {
			continue
		}
		raw, present := parsed[p.Key]
		if !present || isEmptyParam(raw) {
			return newValidationError("fault %q requires parameter %q", faultType, p.Key)
		}
	}
	return nil
}

// isEmptyParam treats JSON null and an empty/whitespace string literal as
// missing so a blank form field cannot satisfy a required parameter.
func isEmptyParam(raw json.RawMessage) bool {
	s := string(raw)
	if s == "" || s == "null" || s == `""` {
		return true
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		for _, r := range str {
			if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
				return false
			}
		}
		return true
	}
	return false
}
