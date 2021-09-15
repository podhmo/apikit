package design

type DB struct{}
type User struct{}
type Messenger struct{}

func ListUser(db *DB) ([]*User, error) {
	return nil, nil
}

func SendMessage(m *Messenger, title string) error {
	return nil
}
