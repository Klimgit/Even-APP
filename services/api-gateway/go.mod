module github.com/even-app/even-app/services/api-gateway

go 1.23

require (
	github.com/even-app/even-app/libs/config v0.0.0
	github.com/even-app/even-app/libs/core v0.0.0
	github.com/even-app/even-app/libs/http v0.0.0
	github.com/even-app/even-app/libs/jwt v0.0.0
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace (
	github.com/even-app/even-app/libs/config => ../../libs/config
	github.com/even-app/even-app/libs/core => ../../libs/core
	github.com/even-app/even-app/libs/http => ../../libs/http
	github.com/even-app/even-app/libs/jwt => ../../libs/jwt
)
