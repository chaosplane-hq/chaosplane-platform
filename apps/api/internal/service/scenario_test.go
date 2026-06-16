package service

import (
	"encoding/json"
	"testing"
)

func step(name, faultType string, deps ...string) ScenarioStep {
	return ScenarioStep{
		Name:      name,
		DependsOn: deps,
		Action:    ActionRequest{Type: faultType, Parameters: json.RawMessage(`{}`)},
		Target:    TargetRequest{Namespace: "default"},
		Duration:  "30s",
	}
}

func TestValidateStepsEmpty(t *testing.T) {
	if err := ValidateSteps(nil); err == nil {
		t.Fatal("expected error for a workflow with no steps")
	}
}

func TestValidateStepsValidLinear(t *testing.T) {
	steps := []ScenarioStep{
		step("a", "pod-kill"),
		step("b", "network-delay", "a"),
		step("c", "node-drain", "b"),
	}
	if err := ValidateSteps(steps); err != nil {
		t.Fatalf("expected valid linear DAG, got %v", err)
	}
}

func TestValidateStepsValidBranch(t *testing.T) {
	steps := []ScenarioStep{
		step("root", "pod-kill"),
		step("left", "network-loss", "root"),
		step("right", "stress-cpu", "root"),
		step("join", "node-drain", "left", "right"),
	}
	if err := ValidateSteps(steps); err != nil {
		t.Fatalf("expected valid branching DAG, got %v", err)
	}
}

func TestValidateStepsDuplicateName(t *testing.T) {
	steps := []ScenarioStep{
		step("a", "pod-kill"),
		step("a", "node-drain"),
	}
	if err := ValidateSteps(steps); err == nil {
		t.Fatal("expected error for duplicate step names")
	}
}

func TestValidateStepsMissingName(t *testing.T) {
	steps := []ScenarioStep{step("", "pod-kill")}
	if err := ValidateSteps(steps); err == nil {
		t.Fatal("expected error for empty step name")
	}
}

func TestValidateStepsUnknownDependency(t *testing.T) {
	steps := []ScenarioStep{
		step("a", "pod-kill"),
		step("b", "node-drain", "ghost"),
	}
	if err := ValidateSteps(steps); err == nil {
		t.Fatal("expected error for dependsOn referencing unknown step")
	}
}

func TestValidateStepsSelfDependency(t *testing.T) {
	steps := []ScenarioStep{step("a", "pod-kill", "a")}
	if err := ValidateSteps(steps); err == nil {
		t.Fatal("expected error for step depending on itself")
	}
}

func TestValidateStepsCycle(t *testing.T) {
	steps := []ScenarioStep{
		step("a", "pod-kill", "c"),
		step("b", "node-drain", "a"),
		step("c", "network-delay", "b"),
	}
	if err := ValidateSteps(steps); err == nil {
		t.Fatal("expected error for cyclic DAG")
	}
	var verr *ValidationError
	if !asValidationError(ValidateSteps(steps), &verr) {
		t.Fatal("expected *ValidationError for cycle")
	}
}

func TestValidateStepsInvalidFaultType(t *testing.T) {
	steps := []ScenarioStep{step("a", "made-up-fault")}
	if err := ValidateSteps(steps); err == nil {
		t.Fatal("expected error for unknown fault type in a step")
	}
}

func TestValidateStepsMissingRequiredParam(t *testing.T) {
	s := ScenarioStep{
		Name:     "a",
		Action:   ActionRequest{Type: "container-kill", Parameters: json.RawMessage(`{}`)},
		Target:   TargetRequest{Namespace: "default"},
		Duration: "30s",
	}
	if err := ValidateSteps([]ScenarioStep{s}); err == nil {
		t.Fatal("expected error for step missing a required fault param")
	}
}
