module github.com/even-app/even-app/services/lexicon

go 1.23

require (
	github.com/even-app/even-app/libs/config v0.0.0
	github.com/even-app/even-app/libs/core v0.0.0
	github.com/even-app/even-app/libs/http v0.0.0
	github.com/even-app/even-app/libs/jwt v0.0.0
	github.com/even-app/even-app/libs/media v0.0.0
	github.com/even-app/even-app/libs/postgres v0.0.0
	github.com/even-app/even-app/libs/s3 v0.0.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.82 // indirect
	github.com/rs/xid v1.6.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace (
	github.com/even-app/even-app/libs/config => ../../libs/config
	github.com/even-app/even-app/libs/core => ../../libs/core
	github.com/even-app/even-app/libs/http => ../../libs/http
	github.com/even-app/even-app/libs/jwt => ../../libs/jwt
	github.com/even-app/even-app/libs/media => ../../libs/media
	github.com/even-app/even-app/libs/postgres => ../../libs/postgres
	github.com/even-app/even-app/libs/s3 => ../../libs/s3
)
