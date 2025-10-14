FROM docker.io/library/golang:1.25.2@sha256:6ce16ef3c8267b7aec46c0f41cb8f4a40295be847b237579324a2859c2f76b1e AS build

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

FROM gcr.io/distroless/base-debian12@sha256:9e9b50d2048db3741f86a48d939b4e4cc775f5889b3496439343301ff54cdba8

COPY --from=build /go/bin/goneypot /usr/bin/goneypot
ENTRYPOINT [ "/usr/bin/goneypot" ]
