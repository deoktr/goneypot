FROM docker.io/library/golang:1.25.2-alpine AS build

ARG VERSION
ARG REVISION
ARG REVISION_TIME

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go .
RUN go build -o /goneypot \
	-buildvcs=false \
	-trimpath \
	-ldflags " \
		-X 'main.Version=${VERSION}' \
		-X 'main.Revision=${REVISION}' \
		-X 'main.RevisionTime=${REVISION_TIME}' \
	"

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build /goneypot /goneypot
ENTRYPOINT ["/goneypot"]
