FROM docker.io/library/golang:1.24 AS build

WORKDIR /src

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src .
RUN go build -o /go/bin/goneypot

FROM gcr.io/distroless/base-debian12:nonroot

USER nonroot
COPY --from=build /go/bin/goneypot /usr/bin/goneypot
ENTRYPOINT [ "/usr/bin/goneypot" ]
