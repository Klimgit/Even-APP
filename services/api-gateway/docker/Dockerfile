FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
ENV GOWORK=off
COPY libs ./libs
COPY services/api-gateway ./services/api-gateway
WORKDIR /src/services/api-gateway
RUN go mod download
RUN CGO_ENABLED=0 go build -o /out/api-gateway ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/api-gateway /bin/api-gateway
ENTRYPOINT ["/bin/api-gateway"]
