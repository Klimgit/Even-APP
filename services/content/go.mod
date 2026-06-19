module github.com/even-app/even-app/services/content

go 1.25.0

require (
	github.com/even-app/even-app/libs/clients v0.0.0
	github.com/even-app/even-app/libs/config v0.0.0
	github.com/even-app/even-app/libs/core v0.0.0
	github.com/even-app/even-app/libs/http v0.0.0
	github.com/even-app/even-app/libs/jwt v0.0.0
	github.com/even-app/even-app/libs/postgres v0.0.0
	github.com/go-faster/errors v0.7.1
	github.com/go-faster/jx v1.2.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/joho/godotenv v1.5.1
	github.com/ogen-go/ogen v1.18.0
)

require (
	github.com/dlclark/regexp2 v1.12.0 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-faster/yaml v0.4.6 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/rogpeppe/go-internal v1.15.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/exp v0.0.0-20230725093048-515e97ebf090 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/even-app/even-app/libs/clients => ../../libs/clients
	github.com/even-app/even-app/libs/config => ../../libs/config
	github.com/even-app/even-app/libs/core => ../../libs/core
	github.com/even-app/even-app/libs/http => ../../libs/http
	github.com/even-app/even-app/libs/jwt => ../../libs/jwt
	github.com/even-app/even-app/libs/postgres => ../../libs/postgres
)
