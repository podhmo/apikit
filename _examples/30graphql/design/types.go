package design

type ID string
type Date string
type Role string // enum (USER, ADMIN)
const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

type Node struct {
	ID ID
}

type User struct {
	ID       ID
	Username string
	Email    string
	Role     Role
}

type Chat struct {
	ID       ID
	Users    []*User
	Messages []*ChatMessage
}

type ChatMessage struct {
	ID      ID
	Content string
	Time    Date
	User    User
}

type SearchResult interface {
	// User, Chat, ChatMessage
}
