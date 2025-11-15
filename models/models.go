package models

type PasswordEntry struct {
	ID       int
	Title    string
	Username string
	Password string
	URL      string
	Notes    string
	Group    string
}

type FilterSettings struct {
	Field string
	Query string
}

type Groups struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}
