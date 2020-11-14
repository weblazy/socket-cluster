package config

var localConfig = `
[dbDefault]
dbHost = "127.0.0.1"        #数据库连接地址
dbName = "websocket_cluster"                  #数据库名称
dbUser = "root"           #数据库用户名
dbPasswd = "123456"   #数据库密码

dbPort = "3306"                     #数据库端口号
dbOpenconns_max = 20                #最大连接数
dbIdleconns_max = 20                #最大空闲连接
dbType = "mysql"
`
