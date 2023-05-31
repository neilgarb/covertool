# covertool

Calculate total coverage from a Go coverage profile, excluding generated files.

Example usage:

```shell
go test -coverprofile=cover.out ./...
covertool -profile=cover.out

# Example output:
total: (statements) 32.66%
```

Installation:

```shell
go install github.com/neilgarb/covertool@latest
```