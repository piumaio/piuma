---
home: true
heroImage: /images/logo.png
heroText: Piuma Image Optimization Service
tagline: A high-performance adaptive image optimization server in Go.
actions:
  - text: Quick Start →
    link: /guide/quick-start/
    type: primary
  - text: View on GitHub
    link: https://github.com/piumaio/piuma
    type: secondary
features:
  - title: 🚀 High-performance
    details: On-the-fly resize, convert & compress with minimal latency using a worker pool.
  - title: 🖼️ Multi-format support
    details: PNG, JPEG, WebP (lossy & lossless) and AVIF via pluggable handlers.
  - title: 🗜️ Adaptive compression
    details: Perceptual quality search balances fidelity and file size automatically.
  - title: 🤝 Smart format negotiation
    details: Accept header + auto-conf penalization chooses the best format per image request.
  - title: 💾 Disk caching
    details: Upstream & optimized image caching with scheduled purge maintains speed & freshness.
  - title: 🧪 Comprehensive test suite
    details: Unit, integration, seam, and optional real-tool tests (use build tag `tools`).
---