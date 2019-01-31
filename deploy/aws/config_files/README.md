# Info
This folder is a placeholder to deploy in the `qed` servers.

# Buildind stage
Here you will need the following files. If aren't present, `config_build.sh`
will generate development ones for testing.

- `id_ed25519` the private key to sing the snapshots
- `server.crt` the server certificate to use TLS connections
- `server.key` the server key to use TLS connections

Each execution will generate the following.

- `qed` the binary to use in the servers.
