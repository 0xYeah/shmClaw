package rbac

import (
	"strings"
	"sync"
)

// Enforcer is a simple Role-Based Access Control engine.
type Enforcer struct {
	// policies maps a subject -> object -> action -> true
	policies map[string]map[string]map[string]bool
	mu       sync.RWMutex
}

// NewEnforcer creates a new Enforcer.
func NewEnforcer() *Enforcer {
	return &Enforcer{
		policies: make(map[string]map[string]map[string]bool),
	}
}

// AddPolicy grants a subject permission to perform an action on an object.
// The object can end with a wildcard '*' to denote prefix matching.
func (e *Enforcer) AddPolicy(subject, object, action string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.policies[subject] == nil {
		e.policies[subject] = make(map[string]map[string]bool)
	}
	if e.policies[subject][object] == nil {
		e.policies[subject][object] = make(map[string]bool)
	}
	e.policies[subject][object][action] = true
}

// Check evaluates if a subject is allowed to perform an action on a specific object.
func (e *Enforcer) Check(subject, object, action string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	subjectPolicies, ok := e.policies[subject]
	if !ok {
		return false
	}

	// 1. Exact match check
	if actions, found := subjectPolicies[object]; found {
		if actions[action] {
			return true
		}
	}

	// 2. Wildcard prefix check
	// Iterating over registered objects to find matching wildcards
	for pattern, actions := range subjectPolicies {
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(object, prefix) {
				if actions[action] {
					return true
				}
			}
		}
	}

	return false
}
