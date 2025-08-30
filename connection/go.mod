module connection-service

go 1.24.6

require (
	github.com/gorilla/websocket v1.5.3
	// github.com/hikkmind/hkchat v0.0.0-20250828083838-bd444f2f8e06
	github.com/lpernett/godotenv v0.0.0-20230527005122-0de1d4c5ef5e
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.30.2
	hkchat/structs v0.0.0
	hkchat/tables v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	// golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	// golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
// google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
// google.golang.org/grpc v1.75.0 // indirect
// google.golang.org/protobuf v1.36.6 // indirect
)

require (
	github.com/hikkmind/hkchat v0.0.0-20250828083838-bd444f2f8e06
	google.golang.org/grpc v1.75.0
)

require (
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)

replace (
	hkchat/proto/tokenverify => ../shared/proto/tokenverify
	hkchat/structs => ../shared/structs
	hkchat/tables => ../shared/tables
)
