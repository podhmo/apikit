package db

type DB struct{}

func NewDB() *DB { return &DB{} }

func (db *DB) Open() (*Connection, error) { return nil, nil }
func (db *DB) Name() string               { return "db" }

type Connection struct{}

func (c *Connection) Close() error { return nil }
