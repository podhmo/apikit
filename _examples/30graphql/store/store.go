package store

type Store struct {
	Users        []*User
	Chats        []*Chat
	ChatMessages []*ChatMessage

	UserToChats map[UserID][]ChatID
	ChatToUsers map[ChatID][]UserID
}

func (s *Store) GetUser(id UserID) *User {
	for _, u := range s.Users {
		if u.ID == id {
			return u
		}
	}
	return nil
}
func (s *Store) GetChat(id ChatID) *Chat {
	for _, u := range s.Chats {
		if u.ID == id {
			return u
		}
	}
	return nil
}
func (s *Store) GetChatMessage(id ChatMessageID) *ChatMessage {
	for _, u := range s.ChatMessages {
		if u.ID == id {
			return u
		}
	}
	return nil
}

type UserID string
type ChatID string
type ChatMessageID string
type Date string

type User struct {
	ID       UserID `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Chat struct {
	ID   ChatID `json:"id"`
	Name string `json:"name"`
}

type ChatToUser struct {
	ChatID ChatID `json:"chatId"`
	UserID UserID `json:"userId"`
}

type ChatMessage struct {
	ID      ChatMessageID `json:"id"`
	ChatID  ChatID        `json:"chatID"`
	UserID  UserID        `json:"userID"`
	Content string        `json:"content"`
	Time    Date          `json:"time"`
}
