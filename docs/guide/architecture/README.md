# Architecture

The system processes an image request through these stages:

1. Parse directive string (e.g. `800_600_70:webp`) into `ImageParameters`.
2. Download original from allowed domain with caching (`DownloadImage`).
3. Dispatch optimization (`Dispatch` / async path) guarding against duplicate in-flight work with a file mutex map.
4. Optimize: decode, resize, possibly convert format, apply fixed or adaptive quality, penalize if larger, persist result.
5. Build HTTP response streaming optimized bytes.

Components:
- **Parser**: Converts directive strings to structured parameters; handles adaptive flag (`a`) and format lists (`auto:webp,avif`).
- **Dispatcher**: Manages locking and initiating optimization, can reuse existing results.
- **WorkerManager**: Goroutine pool receiving async optimization jobs.
- **Optimizer**: Central transformation logic plus penalization and auto-conf file maintenance.
- **Auto Compressor**: `CompressByDSSIM` drives adaptive quality search except for AVIF.
- **Image Handlers**: Encode/decode per format via unified interface.
- **Auto Negotiation**: Maintains gob file of candidate formats; removes penalized ones when larger than original.
