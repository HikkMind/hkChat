module connection-service

go 1.24.6

require (
	github.com/gorilla/websocket v1.5.3
	hkchat/proto/datastream v0.0.0
	hkchat/structs v0.0.0
	hkchat/tables v0.0.0
)

require (
	// golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
// google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
// google.golang.org/grpc v1.75.0 // indirect
// google.golang.org/protobuf v1.36.6 // indirect
)

require (
	github.com/hikkmind/hkchat v0.0.0-20250828083838-bd444f2f8e06
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.8
)

require (
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
)

replace (
	hkchat/proto/datastream => ../shared/proto/datastream
	hkchat/proto/tokenverify => ../shared/proto/tokenverify
	hkchat/structs => ../shared/structs
	hkchat/tables => ../shared/tables
)
