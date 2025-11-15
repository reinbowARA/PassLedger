package models

const (
	DefaultDBPath        string = "data/passwords.db"
	DefaultDBCreateTable string = "db/table.sql"
)

// form name
const (
	TITLE  string = "Название"
	LOGIN  string = "Логин"
	PASSWD string = "Пароль"
	URL    string = "URL"
	NOTES  string = "Заметки"
	GROUP  string = "Группа"
)

const(
	TIME_CLEAR_PASSWD int = 10 //second
)