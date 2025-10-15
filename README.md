# goneypot

Low-interaction SSH honeypot written in Go.

Attackers will be able to log in, and send commands, but nothing is ever executed, just logged.

## Deploy

Generate SSH keys:

```bash
ssh-keygen -f id_rsa -N "" -t rsa
```

Create log directory and files:

```bash
mkdir goneypot_logs
touch goneypot_logs/{goneypot.log,credentials.log}
```

Run container:

```bash
docker run \
  -p 2222:2222 \
  --userns=keep-id \
  -v ./goneypot_logs:/var/log/goneypot \
  -v ./id_rsa:/id_rsa \
  ghcr.io/deoktr/goneypot:latest -logroot "/var/log/goneypot"
```

Connect to the honeypot:

```bash
ssh -p 2222 user@localhost
```

You can then audit the logs in `goneypot_logs/`:

```bash
cat goneypot_logs/goneypot.log
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
  -v ./goneypot_logs:/var/log/goneypot \
  -v ./id_rsa:/id_rsa \
  -v ./creds:/creds \
  ghcr.io/deoktr/goneypot:latest -creds-file creds
```

### Prometheus

goneypot supports [Prometheus](https://prometheus.io/), to enable it use flag `-enable-prometheus`:

```bash
docker run \
  -p 2222:2222 \
  -p 9001:9001 \
  -v ./goneypot_logs:/var/log/goneypot \
  -v ./id_rsa:/id_rsa \
  ghcr.io/deoktr/goneypot:latest -enable-prometheus -prom-port 9001 -prom-addr 0.0.0.0
```

You should create a Docker network and never expose the Prometheus interface, this is just an example.

### AppArmor

An AppArmor profile can be found in `./extras/apparmor/`.

### Systemd

Goneypot can be started in a systemd service, you can find examples in `./extras/systemd/`.

First create a user and a group `goneypot`, then run:

```bash
VERSION=$(git describe --tags)
REVISION=$(git rev-parse --short HEAD)
REVISION_TIME=$(git log -1 --format=%cd --date=iso-strict)
go build -o /usr/bin/goneypot \
  -buildvcs=false \
  -trimpath \
  -ldflags " \
  -X 'github.com/deoktr/goneypot/main.Version=${VERSION}' \
  -X 'github.com/deoktr/goneypot/main.Revision=${REVISION}' \
  -X 'github.com/deoktr/goneypot/main.RevisionTime=${REVISION_TIME}' \
  "
cp ./extras/systemd/goneypot{*.socket,.service} /etc/systemd/system/
cp ./extras/systemd/goneypotpre /usr/bin/goneypotpre
systemctl daemon-reload
systemctl status goneypot.service
```

> [!NOTE]
> By default goneypot (via systemd) will listen on port `22`, this can be changed in `/etc/systemd/system/goneypot.socket`.

> [!NOTE]
> Goneypot configuration can be changed in `/etc/systemd/system/goneypot.service`.

## Development

Build container image locally:

```bash
VERSION=$(git describe --tags)
REVISION=$(git rev-parse --short HEAD)
REVISION_TIME=$(git log -1 --format=%cd --date=iso-strict)
SOURCE_DATE_EPOCH=$(git log -1 --format=%ct)
docker build . -f Containerfile \
  -t "goneypot:${VERSION}" \
  --timestamp ${SOURCE_DATE_EPOCH} \
  --build-arg "VERSION=${VERSION}" \
  --build-arg "REVISION=${REVISION}" \
  --build-arg "REVISION_TIME=${REVISION_TIME}"
```

Run:

```bash
docker run \
  -p 2222:2222 \
  -v ./goneypot_logs:/var/log/goneypot \
  -v ./id_rsa:/id_rsa \
  "goneypot:${VERSION}" -logroot "/var/log/goneypot"
```

## Alternatives

- [cowrie](https://github.com/cowrie/cowrie)
- [sshesame](https://github.com/jaksi/sshesame)

## TODO

- add connections timeout
- add receive limits

## License

goneypot is licensed under [MIT](./LICENSE).
