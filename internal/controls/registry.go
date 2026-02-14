package controls

import (
	"fmt"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = make(map[string]SavingsCalculator)
)

// Register adds a SavingsCalculator to the global registry.
// It panics if a calculator for the same control name is already registered.
func Register(c SavingsCalculator) {
	mu.Lock()
	defer mu.Unlock()

	name := c.ControlName()
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("controls: duplicate registration for %q", name))
	}
	registry[name] = c
}

// Get returns the SavingsCalculator for the given control name, or nil if none is registered.
func Get(controlName string) SavingsCalculator {
	mu.RLock()
	defer mu.RUnlock()
	return registry[controlName]
}

// All returns a copy of all registered calculators.
func All() map[string]SavingsCalculator {
	mu.RLock()
	defer mu.RUnlock()

	out := make(map[string]SavingsCalculator, len(registry))
	for k, v := range registry {
		out[k] = v
	}
	return out
}
