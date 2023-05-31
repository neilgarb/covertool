# covertool

Calculate total coverage from a Go coverage profile, excluding generated files.

Example usage:

```shell
go test -coverprofile=cover.out ./...
covertool -profile=cover.out
```

Installation:

```shell
go install github.com/neilgarb/covertool@latest
```