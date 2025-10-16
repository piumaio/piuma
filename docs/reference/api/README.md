# API Reference

Endpoints:
- `GET /image/<directive>/<url>`: Optimize image per directive. Example: `/image/800_600_70:webp/https://example.com/pic.jpg`
- `OPTIONS /image`: Returns service capability info (supported formats, max dimensions). (Implementation partially covered by tests.)

Headers:
- `Accept`: Influences auto negotiation order.

Response:
- 200 with optimized bytes.
- 4xx/5xx with plain text error message for sentinel errors.
