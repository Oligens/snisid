package compiler

import "fmt"

type PolicyAST struct {
	Name            string
	Condition       string
	ForbiddenAction string
	Invariant       string
}

type CompiledPolicy func(state State, action Action) bool

type State struct {
	FraudRate float64
	Trust     int
}

type Action struct {
	Name string
}

func Compile(ast PolicyAST) CompiledPolicy {
	fmt.Printf("GOVERNANCE-COMPILER: Compiling policy %s into formal predicate...\n", ast.Name)

	return func(s State, a Action) bool {
		// Executable Proof Logic
		if s.FraudRate > 0.7 && a.Name == ast.ForbiddenAction {
			return false
		}
		if s.Trust < 0 {
			return false
		}
		return true
	}
}
