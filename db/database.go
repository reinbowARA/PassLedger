package db

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"embed"

	_ "github.com/mattn/go-sqlite3"
	"github.com/reinbowARA/PassLedger/crypto"
)

//go:embed table.sql
var DefaultDBCreateTable embed.FS

func OpenOrCreateDatabase(dbPath, masterPassword string) (*sql.DB, []byte, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return CreateNewDatabase(dbPath, masterPassword)
	}
	return OpenAndAuthenticate(dbPath, masterPassword)
}

func CreateNewDatabase(dbPath, masterPassword string) (db *sql.DB, key []byte, err error) {
	err = os.MkdirAll(filepath.Dir(dbPath), 0700)
	if err != nil {
		return
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, nil, err
	}

	schema, err := DefaultDBCreateTable.ReadFile("table.sql")
	if err != nil {
		return nil, nil, err
	}
	if _, err := db.Exec(string(schema)); err != nil {
		return nil, nil, err
	}

	salt, _ := crypto.GenerateSalt(16)
	iterations := 20000
	key, _ = crypto.DeriveKeyFromPassword([]byte(masterPassword), salt, iterations)
	verifier := crypto.HMACStreebog256(key, []byte("verifier"))
	_, err = db.Exec(`INSERT INTO meta (id, salt, iterations, verifier) VALUES (1, ?, ?, ?)`, salt, iterations, verifier)
	if err != nil {
		return nil, nil, err
	}
	return db, key, nil
}

func OpenAndAuthenticate(dbPath, masterPassword string) (*sql.DB, []byte, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, nil, err
	}
	row := db.QueryRow(`SELECT salt, iterations, verifier FROM meta WHERE id = 1`)
	var salt []byte
	var iterations int
	var verifier []byte
	if err := row.Scan(&salt, &iterations, &verifier); err != nil {
		return nil, nil, errors.New("ошибка чтения метаданных БД")
	}
	key, _ := crypto.DeriveKeyFromPassword([]byte(masterPassword), salt, iterations)
	expected := crypto.HMACStreebog256(key, []byte("verifier"))
	if !crypto.HmacEqual(expected, verifier) {
		return nil, nil, errors.New("неверный мастер-пароль")
	}
	return db, key, nil
}
