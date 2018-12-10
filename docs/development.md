## Environment

We use the [Go](https://golang.org) programming language and set up the
environment as described in its [documentation](https://golang.org/doc/code.html)

## Useful commands

- Go [documentation server](http://localhost:6061/pkg/github.com/bbva/qed/)

```
godoc -http=:6061 # http://localhost:6061/pkg/github.com/bbva/qed/
```

- Test everything

```
go test -v "${GOPATH}"/src/github.com/bbva/qed/...
```
