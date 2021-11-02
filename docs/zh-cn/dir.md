目录结构
---
.
├── LICENSE
├── README.md
├── api
│   ├── api.go
│   ├── ecode
│   │   ├── ecode.go
│   │   └── ecode_test.go
│   └── response.go
├── conf
│   ├── nacos
│   │   ├── local.go
│   │   ├── nacos.go
│   │   └── viper.go
│   └── viper
│       ├── viper.go
│       └── viper_test.go
├── db
│   ├── orm
│   │   ├── gorm.go
│   │   ├── gorm_test.go
│   │   ├── model.go
│   │   └── sqlx.go
│   └── redis
│       └── redis.go
├── example
│   └── README.md
├── glog
│   ├── log.go
│   ├── log_test.go
│   ├── sls
│   │   └── sls.go
│   └── zap
│       ├── zap.go
│       └── zap_test.go
├── go.mod
├── go.sum
├── mq
│   ├── aliyunmq
│   │   ├── consumer.go
│   │   ├── log.go
│   │   ├── producer.go
│   │   └── rocketmq.go
│   ├── gokafka
│   │   ├── kafka.go
│   │   └── pc.go
│   ├── rabbitmq
│   │   └── rabbitmq.go
│   └── xmqtt
│       ├── model.go
│       └── mqtt.go
├── rpcx
│   ├── config.go
│   ├── grpc_client.go
│   ├── grpc_client_test.go
│   ├── grpc_server.go
│   ├── grpc_server_test.go
│   ├── interceptor
│   │   ├── chain_client.go
│   │   ├── chain_server.go
│   │   ├── crash.go
│   │   └── timeout.go
│   └── logx
│       ├── aliyunlog.go
│       └── logx.go
├── tools
│   └── gocore
│       ├── cmd
│       │   ├── add_mysql.go
│       │   ├── create_service.go
│       │   ├── create_yaml.go
│       │   └── ui.go
│       ├── conf
│       │   ├── const.go
│       │   ├── gocore.go
│       │   ├── gocore_test.go
│       │   └── mysql.go
│       ├── def
│       │   └── def.go
│       ├── file
│       │   ├── file.go
│       │   └── writer.go
│       ├── generate
│       │   └── generate.go
│       ├── gocore.exe
│       ├── main.go
│       ├── makefile
│       ├── template
│       │   ├── Dockerfile.docker
│       │   ├── Dockerfile.docker.go
│       │   ├── README.md
│       │   ├── README.md.go
│       │   ├── api.got
│       │   ├── api.got.go
│       │   ├── api_request.got
│       │   ├── api_request.got.go
│       │   ├── api_routes.got
│       │   ├── api_routes.got.go
│       │   ├── cmd_api.got
│       │   ├── cmd_api.got.go
│       │   ├── cmd_cronjob.got
│       │   ├── cmd_cronjob.got.go
│       │   ├── cmd_init.got
│       │   ├── cmd_init.got.go
│       │   ├── cmd_job.got
│       │   ├── cmd_job.got.go
│       │   ├── conf_base.got
│       │   ├── conf_base.got.go
│       │   ├── conf_const.got
│       │   ├── conf_const.got.go
│       │   ├── conf_local.got
│       │   ├── conf_local.got.go
│       │   ├── conf_mysql.got
│       │   ├── conf_mysql.got.go
│       │   ├── cronjob.got
│       │   ├── cronjob.got.go
│       │   ├── domain.got
│       │   ├── domain.got.go
│       │   ├── domain_handler.got
│       │   ├── domain_handler.got.go
│       │   ├── err_code.got
│       │   ├── err_code.got.go
│       │   ├── job.got
│       │   ├── job.got.go
│       │   ├── main.got
│       │   ├── main.got.go
│       │   ├── model.got
│       │   ├── model.got.go
│       │   ├── model_table.got
│       │   ├── model_table.got.go
│       │   ├── progress.go
│       │   └── template.go
│       └── ui
│           └── view
│               └── home.go
└── utils
    ├── banner.go
    ├── banner_test.go
    ├── closes
    │   ├── close.go
    │   └── close_test.go
    ├── codec
    │   ├── gzip.go
    │   ├── hex.go
    │   └── url.go
    ├── cryption
    │   ├── aes
    │   │   └── aes.go
    │   ├── des
    │   │   └── des.go
    │   └── gorsa
    │       └── README.md
    ├── file
    │   ├── dir.go
    │   ├── file.go
    │   └── zip.go
    ├── gomail
    │   ├── README.md
    │   └── gomail.go
    ├── hash
    │   ├── hmac.go
    │   ├── hmac_test.go
    │   ├── md5.go
    │   ├── md5_test.go
    │   ├── sha.go
    │   ├── sha_test.go
    │   └── test
    ├── http-request
    │   └── http.go
    ├── random.go
    ├── retry.go
    ├── retry_test.go
    ├── trace.go
    ├── type_transform.go
    └── utils.go

