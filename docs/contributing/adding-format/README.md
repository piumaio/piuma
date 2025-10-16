# Adding a New Image Format

1. Create handler implementing `ImageHandler` interface.
2. Provide `Decode` and `Encode`; introduce seams for external tools.
3. Register handler in maps (`image_base.go`).
4. Add tests for success and failure encode/decode paths.
5. Update auto negotiation logic if format should participate.
6. Add documentation page and example directives.
