module github.com/even-app/even-app/libs/http

go 1.23

require github.com/even-app/even-app/libs/jwt v0.0.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace (
	github.com/even-app/even-app/libs/core => ../core
	github.com/even-app/even-app/libs/jwt => ../jwt
)
