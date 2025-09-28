# goneypot

Low-interaction SSH honeypot written in Go.

Attackers will be able to log in, and send commands, but nothing is ever executed, just logged.

## Deploy

Generate SSH keys:

```bash
ssh-keygen -f id_rsa -N "" -t rsa
chmod 666 id_rsa
```

Run container:

```bash
docker run \
  -p 2222:2222 \
  -v $(pwd)/id_rsa:/home/nonroot/id_rsa \
  -v $(pwd)/id_rsa:/home/nonroot/id_rsa \
  ghcr.io/deoktr/goneypot:latest
```

Connect to the honeypot:

```bash
ssh -p 2222 user@localhost
```

### Credentials

By default, goneypot accept any combinaison of username/password.

Login credentials can be added to restrict the username/password that can log in:

1. create a file with `username:password` in it:

```bash
echo "foo:foo" > creds
```

2. start goneypot with the `-creds-file` flag:

```bash
docker run \
  -p 2222:2222 \
  -v $(pwd)/id_rsa:/home/nonroot/id_rsa \
  -v $(pwd)/creds:/home/nonroot/creds \
  ghcr.io/deoktr/goneypot:latest -creds-file creds
```

### Prometheus

goneypot supports [Prometheus](https://prometheus.io/), to enable it use flag `-enable-prometheus`:

```bash
docker run \
  -p 2222:2222 \
  -p 9001:9001 \
  -v $(pwd)/id_rsa:/home/nonroot/id_rsa \
  ghcr.io/deoktr/goneypot:latest -enable-prometheus -prom-port 9001 -prom-addr 0.0.0.0
```

You should create a Docker network and never expose the Prometheus interface, this is just an example.

### AppArmor

An AppArmor profile can be found in `./extras/apparmor/`.

### Systemd

Goneypot can be started in a systemd service, you can find examples in `./extras/systemd/`.

First create a user and a group `goneypot`, then run:

```bash
go build -o /usr/bin/goneypot .
cp ./extras/systemd/goneypot{*.socket,.service} /etc/systemd/system/
cp ./extras/systemd/goneypotpre.sh /usr/bin/goneypotpre
systemctl daemon-reload
systemctl status goneypot.service
```

> [!NOTE]
> By default goneypot (via systemd) will listen on port `22`, this can be changed in `/etc/systemd/system/goneypot.socket`.

> [!NOTE]
> Goneypot configuration can be changed in `/etc/systemd/system/goneypot.service`.

## Alternatives

- [cowrie](https://github.com/cowrie/cowrie)
- [sshesame](https://github.com/jaksi/sshesame)

## TODO

- add connections timeout
- add receive limits

## License

goneypot is licensed under [MIT](./LICENSE).
