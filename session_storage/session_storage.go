package session_storage

const (
	ClientPrefix = "client#"
)

type SessionStorage interface {
	IsOnline(clientId int64) bool
	BindClientId(nodeId string, clientId int64) error
	GetIps(clientId int64) ([]string, error)
	GetClientsIps(clientIds []string) (map[string][]string, error)
	ClientIdsOnline(clientIds []int64) []int64
	OnClientPing(nodeId string, clientId int64) error
}
