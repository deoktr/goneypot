FROM docker.io/library/golang:1.24 AS build

ARG VERSION
ARG REVISION
ARG REVISION_TIME

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go .
RUN go build -o /go/bin/goneypot \
	-buildvcs=false \
	-trimpath \
	-ldflags " \
	-X 'main.Version=${VERSION}' \
	-X 'main.Revision=${REVISION}' \
	-X 'main.RevisionTime=${REVISION_TIME}' \
	"

FROM gcr.io/distroless/base-debian12

COPY --from=build /go/bin/goneypot /usr/bin/goneypot
ENTRYPOINT [ "/usr/bin/goneypot" ]
