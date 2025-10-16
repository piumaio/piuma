# Overview

Piuma is a Go-based image optimization service providing on-the-fly resizing, format conversion, adaptive perceptual compression, caching, and format negotiation. It is designed for high throughput and low latency using a worker pool and minimal allocations.

Key Features:
- Multi-format support: PNG, JPEG, WebP (lossy + lossless), AVIF.
- Adaptive compression with DSSIM-based binary search.
- Automatic format negotiation based on `Accept` header with penalization logic.
- External tool integration (optipng, jpegoptim, avifenc/avifdec, dssim) via overridable seams for testing.
- Disk caching of upstream downloads and optimized results.
- Concurrency via a worker manager and per-file locking to avoid double work.
- Deterministic, testable architecture with extensive unit and integration tests.
