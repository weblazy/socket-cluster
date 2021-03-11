package session_storage

const (
	ClientPrefix = "client#"
)

type SessionStorage interface {
	SetNodeId(nodeId string)
	IsOnline(clientId int64) bool
	BindClientId(clientId int64) error
	GetIps(clientId int64) ([]string, error)
	GetClientsIps(clientIds []string) ([]string, map[string][]string, error)
	ClientIdsOnline(clientIds []int64) []int64
	OnClientPing(clientId int64) error
}
