FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
ENV GOWORK=off
COPY libs ./libs
COPY services/lexicon ./services/lexicon
WORKDIR /src/services/lexicon
RUN go mod download
RUN CGO_ENABLED=0 go build -o /out/lexicon ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/lexicon /bin/lexicon
ENTRYPOINT ["/bin/lexicon"]
