package service

// ScenarioStep is one node in a multi-step experiment DAG. It mirrors the
// engine's chaosworkflow WorkflowTemplate (name + dependencies) wrapping a
// single fault (action/target/duration), so the agent can translate a
// persisted scenario directly into a ChaosWorkflow CRD.
type ScenarioStep struct {
	Name      string        `json:"name"`
	DependsOn []string      `json:"dependsOn,omitempty"`
	Action    ActionRequest `json:"action"`
	Target    TargetRequest `json:"target"`
	Duration  string        `json:"duration,omitempty"`
}

func ValidateSteps(steps []ScenarioStep) error {
	if len(steps) == 0 {
		return newValidationError("a multi-step scenario requires at least one step")
	}

	byName := make(map[string]ScenarioStep, len(steps))
	for _, s := range steps {
		if s.Name == "" {
			return newValidationError("every scenario step requires a name")
		}
		if _, dup := byName[s.Name]; dup {
			return newValidationError("duplicate scenario step name %q", s.Name)
		}
		byName[s.Name] = s
	}

	for _, s := range steps {
		if err := ValidateAction(s.Action.Type, s.Action.Parameters); err != nil {
			return newValidationError("step %q: %s", s.Name, err.Error())
		}
		for _, dep := range s.DependsOn {
			if dep == s.Name {
				return newValidationError("step %q cannot depend on itself", s.Name)
			}
			if _, ok := byName[dep]; !ok {
				return newValidationError("step %q depends on unknown step %q", s.Name, dep)
			}
		}
	}

	return validateAcyclic(steps, byName)
}

// validateAcyclic runs Kahn's algorithm; if any node never reaches in-degree
// zero, the dependency graph contains a cycle and cannot be scheduled.
func validateAcyclic(steps []ScenarioStep, byName map[string]ScenarioStep) error {
	indegree := make(map[string]int, len(steps))
	dependents := make(map[string][]string, len(steps))
	for _, s := range steps {
		indegree[s.Name] += len(s.DependsOn)
		for _, dep := range s.DependsOn {
			dependents[dep] = append(dependents[dep], s.Name)
		}
	}

	queue := make([]string, 0, len(steps))
	for name, deg := range indegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	visited := 0
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		visited++
		for _, child := range dependents[name] {
			indegree[child]--
			if indegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if visited != len(steps) {
		return newValidationError("scenario steps contain a dependency cycle")
	}
	return nil
}
