package core_test

import (
	"testing"

	"github.com/piumaio/piuma/core"
)

const succeeded = "\u2713"
const failed = "\u2717"

// TestParser tests that parser works correctly.
func TestParser(t *testing.T) {
	tt := []struct {
		input           string
		imageParameters core.ImageParameters
	}{
		{
			input: "100_100_50",
			imageParameters: core.ImageParameters{
				Width:   100,
				Height:  100,
				Quality: 50,
			},
		},
		{
			input: "100_100_50/http://someurl",
			imageParameters: core.ImageParameters{
				Width:   100,
				Height:  100,
				Quality: 50,
			},
		},
		{
			input: "100_100__50",
			imageParameters: core.ImageParameters{
				Width:   0,
				Height:  0,
				Quality: 0,
			},
		},
		{
			input: "100_100",
			imageParameters: core.ImageParameters{
				Width:   100,
				Height:  100,
				Quality: 100,
			},
		},
		{
			input: "100_100:png",
			imageParameters: core.ImageParameters{
				Width:   100,
				Height:  100,
				Quality: 100,
				Convert: "png",
			},
		},
		{
			input: "100:jpeg",
			imageParameters: core.ImageParameters{
				Width:   100,
				Quality: 100,
				Convert: "jpeg",
			},
		},
		{
			input: "200_10_10a:jpeg",
			imageParameters: core.ImageParameters{
				Width:           200,
				Height:          10,
				Quality:         10,
				AdaptiveQuality: true,
				Convert:         "jpeg",
			},
		},
	}

	for i, tst := range tt {
		t.Logf("\tTest %d: \t%s", i, tst.input)

		result, _ := core.Parser(tst.input)
		if result.Width != tst.imageParameters.Width {
			t.Fatalf("\t%s\t Should have correct width:  exp[%d] got[%d] ", failed, tst.imageParameters.Width, result.Width)
		}
		t.Logf("\t%s\tShould have correct width\n", succeeded)

		if result.Height != tst.imageParameters.Height {
			t.Fatalf("\t%s\t Should have correct height:  exp[%d] got[%d] ", failed, tst.imageParameters.Height, result.Height)
		}
		t.Logf("\t%s\tShould have correct height\n", succeeded)

		if result.Quality != tst.imageParameters.Quality {
			t.Fatalf("\t%s\t Should have correct quality:  exp[%d] got[%d] ", failed, tst.imageParameters.Quality, result.Quality)
		}
		t.Logf("\t%s\tShould have correct quality\n", succeeded)

		if result.Convert != tst.imageParameters.Convert {
			t.Fatalf("\t%s\t Should have correct convert format:  exp[%d] got[%d] ", failed, tst.imageParameters.Quality, result.Quality)
		}
		t.Logf("\t%s\tShould have correct convert format\n", succeeded)

		if result.AdaptiveQuality != tst.imageParameters.AdaptiveQuality {
			t.Fatalf("\t%s\t Should have correct adaptive quality value:  exp[%d] got[%d] ", failed, tst.imageParameters.Quality, result.Quality)
		}
		t.Logf("\t%s\tShould have correct adaptive quality value\n", succeeded)
	}
}
