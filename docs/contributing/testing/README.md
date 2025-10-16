# Testing Strategy

Testing layers:
- Unit: parser, handlers, optimizer branches.
- Integration: full request lifecycle.
- External tool seams: success/failure without real binaries.
- Real tool tests behind build tag `tools`.

Run tests:
```
go test ./...
```
With real tools:
```
go test -tags tools ./...
```
Coverage:
```
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```
