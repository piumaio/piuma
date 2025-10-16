# Automatic Format Negotiation & Penalization

Negotiation chooses best format based on the client `Accept` header and historical success.

Process:
1. Parse `Accept` header -> ordered media types.
2. Intersect with internal supported handlers.
3. Create or read auto-conf gob file listing still-eligible types.
4. Select first available preferred type; if directive `auto:webp,avif` restricts to subset.
5. After optimization if new bytes are larger than original and different format was used, remove the format from auto-conf (penalization) and reset directive to `auto` for future.

Penalization prevents repeatedly selecting inefficient formats for specific images.
