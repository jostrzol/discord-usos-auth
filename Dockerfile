FROM golang:1.16.2 AS builder

WORKDIR /go/src/discord-usos-auth
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

FROM alpine:3.13.3
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/discord-usos-auth/discord-usos-auth /root/
ENTRYPOINT ./discord-usos-auth -t ${TOKEN} -s ${SETTINGS_FILE} -f