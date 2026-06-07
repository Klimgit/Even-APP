FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
ENV GOWORK=off
COPY libs ./libs
COPY services/media ./services/media
WORKDIR /src/services/media
RUN go mod download
RUN CGO_ENABLED=0 go build -o /out/media ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/media /bin/media
ENTRYPOINT ["/bin/media"]
