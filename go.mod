module github.com/weblazy/socket-cluster

go 1.14

require (
	git.apache.org/thrift.git v0.0.0-20190629060710-d9019fc5a4a2
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-redis/redis/v8 v8.5.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/henrylee2cn/goutil v0.0.0-20191020121818-c6a890a2c537
	github.com/jinzhu/gorm v1.9.16
	github.com/labstack/echo/v4 v4.1.16
	github.com/lucas-clemente/quic-go v0.19.3
	github.com/nacos-group/nacos-sdk-go v1.0.6
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cast v1.3.1
	github.com/sunmi-OS/gocore v1.5.10
	github.com/tidwall/gjson v1.6.0
	github.com/weblazy/core v1.1.1
	github.com/weblazy/crypto v1.0.1
	github.com/weblazy/easy v1.1.2
	github.com/weblazy/goutil v1.1.2
	google.golang.org/genproto v0.0.0-20210126160654-44e461bb6506 // indirect
	google.golang.org/grpc v1.35.0 // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df

)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
