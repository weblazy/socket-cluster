package discover

//ServiceDiscovery 服务发现
type ServiceDiscovery interface {
	//设置租约
	// func (s *ServiceDiscovery) GetServices() ([]string,error)
	WatchService()
	Register() error
	// func (s *ServiceDiscovery) Ping(value []byte)
	// func (s *ServiceDiscovery) Close() error
}
