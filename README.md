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

Resources:

- https://go.dev/testing/coverage
- [ParseProfiles](https://github.com/golang/go/blob/0104a31b8fbcbe52728a08867b26415d282c35d2/src/cmd/cover/profile.go#L44) documentation
