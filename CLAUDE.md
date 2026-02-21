# paylio-go

## Changed-Scope Checks (mandatory before every commit)
- `go build ./...` (mandatory)
- `go test -race -coverprofile=coverage.out ./...`
- `go vet ./...`
- `gofmt -l .` (must produce no output)
