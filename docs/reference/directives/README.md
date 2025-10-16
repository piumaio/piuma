# Directive Syntax

Directive string format examples:
- `WIDTH_HEIGHT_QUALITY:FORMAT`
- `WIDTH_HEIGHT_QUALITYa:FORMAT` (adaptive quality search)
- `WIDTH_HEIGHT:FORMAT` (quality defaults to 100)
- `WIDTH:FORMAT` (height omitted => original height)
- `WIDTH_HEIGHT_QUALITY:auto:webp,avif` (auto negotiation restricted to listed formats)

Rules:
- Omitted sections collapse to zero or default quality.
- `a` suffix after quality triggers adaptive search.
- `auto` format chooses best negotiated type from Accept list.
