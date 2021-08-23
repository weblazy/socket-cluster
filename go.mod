module github.com/weblazy/socket-cluster

go 1.14

require (
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/go-redis/redis/v8 v8.5.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/labstack/echo/v4 v4.1.16
	github.com/lucas-clemente/quic-go v0.19.3
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cast v1.3.1
	github.com/weblazy/core v1.1.1
	github.com/weblazy/easy v1.1.2
	github.com/weblazy/goutil v1.1.2
	google.golang.org/genproto v0.0.0-20210126160654-44e461bb6506 // indirect
	google.golang.org/grpc v1.35.0 // indirect

)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
