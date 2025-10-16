# External Tools & Seams

External binaries used:
- `optipng` for additional PNG optimization.
- `jpegoptim` for JPEG size reduction.
- `avifenc` / `avifdec` for AVIF encode/decode cycles.
- `dssim` for perceptual similarity measurement.

Each tool invocation is abstracted by a package-level function variable (seam) so tests can override or simulate failures (e.g. return `exec.Command("false")`). This enables deterministic unit tests without requiring system-level dependencies.

Add a new tool:
1. Wrap command invocation in a function variable.
2. Provide encode/decode logic using that seam.
3. Write tests overriding seam for success/failure.
