# Caching & Concurrency

Caching:
- Upstream originals cached on disk by hashed path; later requests reuse local copy.
- Periodic purge goroutine removes old entries based on TTL.

Concurrency:
- File-level mutex map prevents simultaneous optimization of same target.
- WorkerManager channel + goroutine pool handles async optimization tasks; idempotent `Close` avoids panics.
- Timeout test ensures long-running optimization does not block indefinitely.
