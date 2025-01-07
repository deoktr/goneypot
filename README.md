# goneypot

SSH honeypot written in Go.

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
