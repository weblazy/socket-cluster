package session_storage

const (
	ClientPrefix = "client#"
)

type SessionStorage interface {
	IsOnline(clientId string) bool
	BindClientId(nodeAddr string, clientId string) error
	GetIps(clientId string) ([]string, error)
	GetClientsIps(clientIds []string) (map[string][]string, error)
	ClientIdsOnline(clientIds []string) []string
	OnClientPing(nodeAddr string, clientId string) error
}
