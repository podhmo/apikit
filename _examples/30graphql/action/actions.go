package action

import (
	"context"
	"fmt"
	"m/30graphql/design"
	s "m/30graphql/store"
)

func toUser(store *s.Store, u *s.User) *design.User {
	return &design.User{
		ID:       design.ID(u.ID),
		Username: u.Username,
		Email:    u.Email,
		Role:     design.RoleAdmin, // xxx
	}
}
func toChat(store *s.Store, c *s.Chat) *design.Chat {
	return &design.Chat{
		ID: design.ID(c.ID),
		// without Users and Messages
	}
}
func toChatMessage(store *s.Store, m *s.ChatMessage) *design.ChatMessage {
	return &design.ChatMessage{
		ID:      design.ID(m.ID),
		UserID:  design.ID(m.UserID),
		ChatID:  design.ID(m.ChatID),
		Content: m.Content,
		Time:    design.Date(m.Time),
	}
}

func ChatActiveUsers(
	ctx context.Context,
	store *s.Store,
	chat *design.Chat) ([]*design.User, error) {
	return nil, nil
}

func Me(
	ctx context.Context,
	store *s.Store,
) (*design.User, error) {
	u := store.Users[0]
	return toUser(store, u), nil
}

func User(
	ctx context.Context,
	store *s.Store,
	id design.ID,
) (*design.User, error) {
	u := store.GetUser(s.UserID(id))
	if u == nil {
		return nil, fmt.Errorf("not found") // todo: custom merror
	}
	return toUser(store, u), nil
}
func AllUsers(
	ctx context.Context,
	store *s.Store,
) ([]*design.User, error) {
	r := make([]*design.User, len(store.Users))
	for i, u := range store.Users {
		r[i] = toUser(store, u)
	}
	return r, nil
}

func MyChats(
	ctx context.Context,
	store *s.Store,
) ([]*design.Chat, error) {
	me, err := Me(ctx, store)
	if err != nil {
		return nil, err
	}

	var myChatIDs []s.ChatID
	for _, chatID := range store.UserToChats[s.UserID(me.ID)] {
		myChatIDs = append(myChatIDs, chatID)
	}

	r := make([]*design.Chat, 0, len(myChatIDs))
	for _, chat := range store.Chats {
		for _, myChatID := range myChatIDs {
			if myChatID == chat.ID {
				r = append(r, toChat(store, chat))
				break
			}
		}
	}
	return r, nil
}

func ChatToUsers(
	ctx context.Context,
	store *s.Store,
	chat *design.Chat,
) ([]*design.User, error) {
	var r []*design.User
	for _, userId := range store.ChatToUsers[s.ChatID(chat.ID)] {
		r = append(r, toUser(store, store.GetUser(userId)))
	}
	return r, nil
}
func ChatToMessages(
	ctx context.Context,
	store *s.Store,
	chat *design.Chat,
) ([]*design.ChatMessage, error) {
	var r []*design.ChatMessage
	chatID := s.ChatID(chat.ID)
	for _, m := range store.ChatMessages {
		if m.ChatID == chatID {
			r = append(r, toChatMessage(store, m))
		}
	}
	return r, nil
}
func ChatMessageToUser(
	ctx context.Context,
	store *s.Store,
	m *design.ChatMessage,
) (*design.User, error) {
	// todo: handling easily
	return toUser(store, store.GetUser(s.UserID(m.UserID))), nil
}

// r.Field("search", func(term string) ([]SearchResult, error) { return nil, nil }),
