# Directive Examples

```
800_600_70:webp                # resize + convert to webp quality 70
800_0_85:jpeg                  # width 800 preserve height
400:png                        # width 400 preserve height convert to png
800_600_80a:auto:webp,avif     # adaptive search, auto negotiation restricted
100_100__50                    # malformed example (will error)
```
