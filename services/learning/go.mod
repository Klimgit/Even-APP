module github.com/even-app/even-app/services/learning

go 1.24.0

require (
	github.com/even-app/even-app/libs/config v0.0.0
	github.com/even-app/even-app/libs/core v0.0.0
	github.com/even-app/even-app/libs/http v0.0.0
	github.com/even-app/even-app/libs/postgres v0.0.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/even-app/even-app/libs/jwt v0.0.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.44.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

replace (
	github.com/even-app/even-app/libs/config => ../../libs/config
	github.com/even-app/even-app/libs/core => ../../libs/core
	github.com/even-app/even-app/libs/http => ../../libs/http
	github.com/even-app/even-app/libs/jwt => ../../libs/jwt
	github.com/even-app/even-app/libs/postgres => ../../libs/postgres
)
