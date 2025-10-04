package db

import (
	"database/sql"

	"github.com/reinbowARA/PassLedger/crypto"
	"github.com/reinbowARA/PassLedger/models"
)

// SaveEntry сохраняет новую запись (шифрует поля)
func SaveEntry(dbConn *sql.DB, key []byte, e models.PasswordEntry) error {
	encTitle, err := crypto.EncryptData(key, []byte(e.Title))
	if err != nil {
		return err
	}
	encUser, err := crypto.EncryptData(key, []byte(e.Username))
	if err != nil {
		return err
	}
	encPass, err := crypto.EncryptData(key, []byte(e.Password))
	if err != nil {
		return err
	}
	var encURL, encNotes []byte
	if e.URL != "" {
		encURL, err = crypto.EncryptData(key, []byte(e.URL))
		if err != nil {
			return err
		}
	}
	if e.Notes != "" {
		encNotes, err = crypto.EncryptData(key, []byte(e.Notes))
		if err != nil {
			return err
		}
	}

	_, err = dbConn.Exec(`INSERT INTO entries (title, username, password, url, notes, "group")
		VALUES (?, ?, ?, ?, ?, ?)`, encTitle, encUser, encPass, encURL, encNotes, e.Group)
	return err
}

// LoadAllEntries загружает все записи и дешифрует их
func LoadAllEntries(dbConn *sql.DB, key []byte) ([]models.PasswordEntry, error) {
	rows, err := dbConn.Query(`SELECT id, title, username, password, url, notes, "group" FROM entries ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.PasswordEntry, 0)
	for rows.Next() {
		var id int
		var ctTitle, ctUser, ctPass, ctURL, ctNotes []byte
		var group sql.NullString

		if err := rows.Scan(&id, &ctTitle, &ctUser, &ctPass, &ctURL, &ctNotes, &group); err != nil {
			return nil, err
		}

		title := ""
		if len(ctTitle) > 0 {
			pt, err := crypto.DecryptData(key, ctTitle)
			if err != nil {
				return nil, err
			}
			title = string(pt)
		}
		username := ""
		if len(ctUser) > 0 {
			pt, err := crypto.DecryptData(key, ctUser)
			if err != nil {
				return nil, err
			}
			username = string(pt)
		}
		password := ""
		if len(ctPass) > 0 {
			pt, err := crypto.DecryptData(key, ctPass)
			if err != nil {
				return nil, err
			}
			password = string(pt)
		}
		url := ""
		if len(ctURL) > 0 {
			pt, err := crypto.DecryptData(key, ctURL)
			if err != nil {
				return nil, err
			}
			url = string(pt)
		}
		notes := ""
		if len(ctNotes) > 0 {
			pt, err := crypto.DecryptData(key, ctNotes)
			if err != nil {
				return nil, err
			}
			notes = string(pt)
		}

		out = append(out, models.PasswordEntry{
			ID:       id,
			Title:    title,
			Username: username,
			Password: password,
			URL:      url,
			Notes:    notes,
			Group:    group.String,
		})
	}
	return out, nil
}

// DeleteEntry удаляет запись по id
func DeleteEntry(dbConn *sql.DB, id int) error {
	_, err := dbConn.Exec(`DELETE FROM entries WHERE id = ?`, id)
	return err
}

// UpdateEntry обновляет запись (шифрует поля)
func UpdateEntry(dbConn *sql.DB, key []byte, e models.PasswordEntry) error {
	encTitle, err := crypto.EncryptData(key, []byte(e.Title))
	if err != nil {
		return err
	}
	encUser, err := crypto.EncryptData(key, []byte(e.Username))
	if err != nil {
		return err
	}
	encPass, err := crypto.EncryptData(key, []byte(e.Password))
	if err != nil {
		return err
	}
	encURL, err := crypto.EncryptData(key, []byte(e.URL))
	if err != nil {
		return err
	}
	encNotes, err := crypto.EncryptData(key, []byte(e.Notes))
	if err != nil {
		return err
	}

	_, err = dbConn.Exec(`UPDATE entries SET title=?, username=?, password=?, url=?, notes=?, "group"=? WHERE id=?`,
		encTitle, encUser, encPass, encURL, encNotes, e.Group, e.ID)
	return err
}

// DeleteGroup удаляет все записи в группе (если нужно)
func DeleteGroup(dbConn *sql.DB, group string) error {
	_, err := dbConn.Exec(`DELETE FROM entries WHERE "group" = ?`, group)
	return err
}
