package db

import (
	"database/sql"
)

// GetMeta возвращает salt, iterations, verifier (id=1)
func GetMeta(db *sql.DB) (salt []byte, iterations int, verifier []byte, err error) {
	row := db.QueryRow(`SELECT salt, iterations, verifier FROM meta WHERE id = 1`)
	err = row.Scan(&salt, &iterations, &verifier)
	return
}

// UpdateVerifier обновляет verifier (если нужно поменять мастер-пароль)
func UpdateVerifier(db *sql.DB, newVerifier []byte) error {
	_, err := db.Exec(`UPDATE meta SET verifier = ? WHERE id = 1`, newVerifier)
	return err
}
