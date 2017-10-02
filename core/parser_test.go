package core_test

import (
	"testing"

	"github.com/lotrekagency/piuma/core"
)

const succeeded = "\u2713"
const failed = "\u2717"

// TestParser tests that parser works correctly.
func TestParser(t *testing.T) {
	tt := []struct {
		parameters string
		width      uint
		height     uint
		quality    uint
	}{
		{
			parameters: "100_100_50",
			width:      100,
			height:     100,
			quality:    50,
		},
		{
			parameters: "100_100_50/http://someurl",
			width:      100,
			height:     100,
			quality:    50,
		},
		{
			parameters: "100_100__50",
			width:      0,
			height:     0,
			quality:    0,
		},
		{
			parameters: "100_100",
			width:      100,
			height:     100,
			quality:    100,
		},
	}

	for i, tst := range tt {
		t.Logf("\tTest %d: \t%s", i, tst.parameters)

		width, height, quality, _ := core.Parser(tst.parameters)
		if width != tst.width {
			t.Fatalf("\t%s\t Should have correct width:  exp[%d] got[%d] ", failed, tst.width, width)
		}
		t.Logf("\t%s\tShould have correct width\n", succeeded)

		if height != tst.height {
			t.Fatalf("\t%s\t Should have correct height:  exp[%d] got[%d] ", failed, tst.height, height)
		}
		t.Logf("\t%s\tShould have correct height\n", succeeded)

		if quality != tst.quality {
			t.Fatalf("\t%s\t Should have correct quality:  exp[%d] got[%d] ", failed, tst.quality, quality)
		}
		t.Logf("\t%s\tShould have correct quality\n", succeeded)
	}
}
