# Configuration Reference

Server initialization parses configuration (environment flags or defaults). Key options you should document here (adapt once code exposes them explicitly):
- Allowed domains list for upstream downloads.
- Cache directory paths (temporary and media storage).
- Worker pool size.
- Purge interval and TTL for cache entries.
- Max adaptive iterations or quality tolerance.

(If not yet exposed via flags, add them; this doc is a placeholder for forthcoming explicit config parsing details.)
