module github.com/even-app/even-app/libs/s3

go 1.24.0

require (
	github.com/even-app/even-app/libs/config v0.0.0
	github.com/minio/minio-go/v7 v7.0.82
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

replace github.com/even-app/even-app/libs/config => ../config
