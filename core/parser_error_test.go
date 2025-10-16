package core_test

import (
	"testing"

	"github.com/piumaio/piuma/core"
)

func TestParserError(t *testing.T) {
	_, err := core.Parser("abc_100_50")
	if err == nil {
		t.Fatalf("expected error for non-numeric width")
	}
}
