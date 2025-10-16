# Error Handling

Common sentinel errors mapped to HTTP responses:
- Invalid status code from upstream.
- Invalid content type (non-image).
- Invalid domain (not in allow-list).
- Malformed URL.

Errors propagate from optimization (decode failures, external tool failures, adaptive compression issues). Response builder sets appropriate status and body when sentinel encountered.
