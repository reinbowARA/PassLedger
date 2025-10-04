package db

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/reinbowARA/PassLedger/crypto"
)

func OpenOrCreateDatabase(dbPath, masterPassword string) (*sql.DB, []byte, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return CreateNewDatabase(dbPath, masterPassword)
	}
	return OpenAndAuthenticate(dbPath, masterPassword)
}

func CreateNewDatabase(dbPath, masterPassword string) (*sql.DB, []byte, error) {
	os.MkdirAll(filepath.Dir(dbPath), 0700)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS meta (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		salt BLOB NOT NULL,
		iterations INTEGER NOT NULL,
		verifier BLOB NOT NULL
	);
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title BLOB NOT NULL,
		username BLOB NOT NULL,
		password BLOB NOT NULL,
		url BLOB,
		notes BLOB,
		"group" TEXT
	);`
	if _, err := db.Exec(schema); err != nil {
		return nil, nil, err
	}

	salt, _ := crypto.GenerateSalt(16)
	iterations := 20000
	key, _ := crypto.DeriveKeyFromPassword([]byte(masterPassword), salt, iterations)
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
