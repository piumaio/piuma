# Adaptive Compression

Adaptive compression targets a perceptual quality threshold using DSSIM (structural dissimilarity). The algorithm:

1. Start with high quality guess.
2. Encode, create temporary PNG for DSSIM comparison if needed.
3. Invoke `dssim` command (overridable seam) capturing similarity score.
4. Adjust quality bounds via binary search until difference below tolerance or max iterations reached.
5. Fall back to last successful quality.

AVIF skips adaptive search due to its encoder semantics; other formats (JPEG, WebP, PNG) can use it when `a` flag present (`80a` means target around quality 80 adaptively).

Failure Modes & Handling:
- Missing `dssim` binary: returns error, caller may fall back.
- Temp file creation failure: surfaced via error path tests.
- Score parse errors: abort adaptive search.
