# goneypot

Low-interaction SSH honeypot written in Go.

Attackers will be able to log in, and send commands, but nothing is ever executed, just logged.

## Usage

Generate SSH server private keys without passphrase:

```bash
ssh-keygen -f id_rsa -N "" -t rsa
```

```bash
go run .
```

## License

goneypot is licensed under [MIT](./LICENSE).
