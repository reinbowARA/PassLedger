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

type SearchFilters struct {
	Title    bool
	Username bool
	URL      bool
	Group    bool
	Notes    bool
}

type Groups struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}
