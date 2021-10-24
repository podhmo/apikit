package database

type User struct {
	Name string `json:"name"`
	Password string `json:"-"`
}

type DB struct {
	Users map[string]*User
}

