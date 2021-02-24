package session_storage

const (
	ClientPrefix = "client#"
)

type SessionStorage interface {
	IsOnline(clientId string) bool
	BindClientId(clientId string) error
	GetIps(clientId string) ([]string, error)
	ClientIdsOnline(clientIds []int64) []int64
}
