package session_storage

const (
	ClientPrefix = "client#"
)

type SessionStorage interface {
	IsOnline(clientId string) bool
	BindClientId(nodeId string, clientId string) error
	GetIps(clientId string) ([]string, error)
	GetClientsIps(clientIds []string) (map[string][]string, error)
	ClientIdsOnline(clientIds []string) []string
	OnClientPing(nodeId string, clientId string) error
}
