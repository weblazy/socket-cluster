package websocket_cluster

type (
	MasterConf struct {
		Port     int64  //Socket config
		Password string //Password for auth when node connect on
	}
)

// NewPeer creates a new peer.
func NewMasterConf() *MasterConf {
	return &MasterConf{
		Port:     defaultMasterPort,
		Password: defaultPassword,
	}
}

func (conf *MasterConf) WithPassword(password string) *MasterConf {
	conf.Password = password
	return conf
}

func (conf *MasterConf) WithPort(port int64) *MasterConf {
	conf.Port = port
	return conf
}
