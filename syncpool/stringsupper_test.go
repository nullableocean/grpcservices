package syncpool

import (
	"fmt"
	"testing"
)

// go test -v stringsupper.go stringsupper_test.go
func TestProcessString_Work(t *testing.T) {
	examples := []string{
		"hello, world!",
		"gopher",
		"lorem ipsum dolor sit amet",
	}

	for _, s := range examples {
		processed := ProcessString(s)
		fmt.Printf("Original: %q\nProcessed: %q\n\n", s, processed)
	}
}
