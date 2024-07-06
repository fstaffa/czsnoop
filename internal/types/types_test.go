package types

import (
	"testing"
)

func Test_CreateIci_FailsWithInvalidIco(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		name string
		ico  string
	}{
		"too short":    {ico: "1234567"},
		"too long":     {ico: "123456789"},
		"not a number": {ico: "1234567a"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := CreateIco(test.ico)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}
		})
	}
}
