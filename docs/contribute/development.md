## Environment

We use the [Go](https://golang.org) programming language and set up the
environment as described in its [documentation](https://golang.org/doc/code.html)

## Self-signed certificate

```
cd ~/.ssh/

# Generate private key (qed_ed25519|.pub)
go run main.go generate signerkeys


# Generation of self-signed(x509) public key (PEM-encodings qed_key.pem|qed_cert.pem)
go run main.go generate self-signed-cert --host qed.awesome.lan
```

## Useful commands

- Go [documentation server](http://localhost:6061/pkg/github.com/bbva/qed/)

```
godoc -http=:6061 # http://localhost:6061/pkg/github.com/bbva/qed/
```

- Test everything

```
go test -v "${GOPATH}"/src/github.com/bbva/qed/...
```
