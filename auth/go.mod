module auth-service

go 1.24.6

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/lpernett/godotenv v0.0.0-20230527005122-0de1d4c5ef5e
	github.com/redis/go-redis/v9 v9.12.1
	google.golang.org/grpc v1.75.0
	gorm.io/driver/postgres v1.5.0
	gorm.io/gorm v1.25.1
	hkchat/proto/tokenverify v0.0.0
	hkchat/tables v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)

replace (
	hkchat/proto/tokenverify => ../shared/proto/tokenverify
	hkchat/tables => ../shared/tables
)
