FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
ENV GOWORK=off
COPY libs ./libs
COPY services/auth ./services/auth
WORKDIR /src/services/auth
RUN go mod download
RUN CGO_ENABLED=0 go build -o /out/auth ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/auth /bin/auth
ENTRYPOINT ["/bin/auth"]
