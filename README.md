# goneypot

Low-interaction SSH honeypot written in Go.

Attackers will be able to log in, and send commands, but nothing is ever executed, just logged.

## Usage

Generate SSH server private keys without passphrase:

```bash
ssh-keygen -f id_rsa -N "" -t rsa
```

Build:

```bash
go build .
```

Run:

```bash
./goneypot -key id_rsa -addr 0.0.0.0 -port 2222
```

Test:

```bash
ssh -p 2222 user@localhost
```

### Prometheus

goneypot supports [Prometheus](https://prometheus.io/), to enable it use flag `-enable-prometheus`:

```bash
goneypot -enable-prometheus -prom-port 9001 -prom-addr localhost
```

## Alternatives

- [cowrie](https://github.com/cowrie/cowrie)
- [sshesame](https://github.com/jaksi/sshesame)

## License

goneypot is licensed under [MIT](./LICENSE).
