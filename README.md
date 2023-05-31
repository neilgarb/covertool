# Covertool

Calculate total coverage from a Go coverage profile, excluding generated files

Example usage:

```sh
$ go test -coverprofile=cover.out ./...
$ covertool -profile=cover.out
```
