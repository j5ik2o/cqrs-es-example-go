// Package ciproof breaks go vet on purpose for the #701 validation draft.
package ciproof

import "fmt"

// Broken passes a string to %d so that go vet fails.
func Broken() string { return fmt.Sprintf("%d", "not a number") }
