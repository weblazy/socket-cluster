package config

import (
	"os"

	"time"

	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/sunmi-OS/gocore/nacos"
)

func InitNacos(runtime string) {
	time.Sleep(time.Second * 10)
	nacos.SetRunTime(runtime)
	nacos.ViperTomlHarder.SetviperBase(baseConfig)
	switch runtime {
	case "local":
		nacos.AddLocalConfig(runtime, localConfig)
	default:
		Endpoint := os.Getenv("ENDPOINT")
		NamespaceId := os.Getenv("NAMESPACE_ID")
		AccessKey := os.Getenv("ACCESS_KEY")
		SecretKey := os.Getenv("SECRET_KEY")
		if Endpoint == "" || NamespaceId == "" || AccessKey == "" || SecretKey == "" {
			panic("The configuration file cannot be empty.")
		}
		err := nacos.AddAcmConfig(runtime, constant.ClientConfig{
			Endpoint:    Endpoint,
			NamespaceId: NamespaceId,
			AccessKey:   AccessKey,
			SecretKey:   SecretKey,
		})
		if err != nil {
			panic(err)
		}
	}
	//dataIds:为config包含其他业务如[appstore][sunmipay]等
	nacos.ViperTomlHarder.SetDataIds("file-core", "system", "mysql", "oss", "contentmoderation","config")

	//配置变动回调
	nacos.ViperTomlHarder.SetCallBackFunc("file-core", "mysql", func(namespace, group, dataId, data string) {

	})
	nacos.ViperTomlHarder.NacosToViper()
}
