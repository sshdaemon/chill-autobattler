module chill/server

go 1.27.4

require (
	chill/sim v0.0.0
	connectrpc.com/connect/v2 v2.0.2
	google.golang.org/protobuf v1.36.12
)

replace chill/sim => ../sim
