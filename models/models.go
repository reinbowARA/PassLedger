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

type Settings struct {
	DBPath       string `json:"db_path"`
	ThemeVariant int    `json:"theme_variant"`
	TimerSeconds int    `json:"timer_seconds"`
}

type PasswordGeneratorOptions struct {
	Length       int
	UseUppercase bool
	UseLowercase bool
	UseDigits    bool
	UseSpecial   bool
	UseSpace     bool
	UseBrackets  bool
}
