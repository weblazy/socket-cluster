module github.com/weblazy/socket-cluster

go 1.14

require (
	github.com/cockroachdb/pebble v0.0.0-20210125151244-6c981576bc1e // indirect
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/go-redis/redis/v8 v8.5.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/jinzhu/gorm v1.9.16
	github.com/labstack/echo/v4 v4.1.16
	github.com/nacos-group/nacos-sdk-go v1.0.6
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/spf13/cast v1.3.1
	github.com/sunmi-OS/gocore v1.5.10
	github.com/tidwall/gjson v1.6.0
	github.com/weblazy/core v1.1.1
	github.com/weblazy/crypto v1.0.1
	github.com/weblazy/easy v1.1.2
	github.com/weblazy/goutil v1.1.2
	go.uber.org/zap v1.16.0 // indirect
	google.golang.org/genproto v0.0.0-20210126160654-44e461bb6506 // indirect
	google.golang.org/grpc v1.35.0 // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
