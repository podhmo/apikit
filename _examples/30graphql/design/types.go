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
	ID       ID     `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     Role   `json:"role"`
}

type Chat struct {
	ID       ID             `json:"id"`
	Users    []*User        `json:"users"`
	Messages []*ChatMessage `json:"messages"`
}

type ChatMessage struct {
	ID      ID     `json:"id"`
	Content string `json:"content"`
	Time    Date   `json:"time"`
	User    *User  `json:"user"`
}

type SearchResult interface {
	// User, Chat, ChatMessage
}
