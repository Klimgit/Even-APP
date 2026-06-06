FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
ENV GOWORK=off
COPY libs ./libs
COPY services/content ./services/content
WORKDIR /src/services/content
RUN go mod download
RUN CGO_ENABLED=0 go build -o /out/content ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/content /bin/content
ENTRYPOINT ["/bin/content"]
