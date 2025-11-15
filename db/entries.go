package db

import (
	"database/sql"
	"fmt"

	"github.com/reinbowARA/PassLedger/crypto"
	"github.com/reinbowARA/PassLedger/models"
)

func getOrCreateGroup(dbConn *sql.DB, name string) (sql.NullInt64, error) {
	if name == "" {
		return sql.NullInt64{Valid: false}, nil
	}
	var id sql.NullInt64
	err := dbConn.QueryRow(`SELECT id FROM groups WHERE name = ?`, name).Scan(&id)
	if err == sql.ErrNoRows {
		result, err := dbConn.Exec(`INSERT INTO groups (name) VALUES (?)`, name)
		if err != nil {
			return sql.NullInt64{}, err
		}
		insertID, _ := result.LastInsertId()
		id = sql.NullInt64{Int64: insertID, Valid: true}
	} else if err != nil {
		return sql.NullInt64{}, err
	}
	return id, nil
}

// SaveEntry сохраняет новую запись (шифрует поля)
func SaveEntry(dbConn *sql.DB, key []byte, e models.PasswordEntry) error {
	groupId, err := getOrCreateGroup(dbConn, e.Group)
	if err != nil {
		return err
	}
	encTitle, _ := crypto.EncryptData(key, []byte(e.Title))
	encUser, _ := crypto.EncryptData(key, []byte(e.Username))
	encPass, _ := crypto.EncryptData(key, []byte(e.Password))
	var encURL []byte
	if e.URL != "" {
		encURL, _ = crypto.EncryptData(key, []byte(e.URL))
	}
	var encNotes []byte
	if e.Notes != "" {
		encNotes, _ = crypto.EncryptData(key, []byte(e.Notes))
	}

	_, err = dbConn.Exec(`INSERT INTO entries (title, username, password, url, notes, group_id)
		VALUES (?, ?, ?, ?, ?, ?)`, encTitle, encUser, encPass, encURL, encNotes, groupId)
	return err
}

// LoadAllEntries загружает все записи и дешифрует их
func LoadAllEntries(dbConn *sql.DB, key []byte) ([]models.PasswordEntry, error) {
	rows, err := dbConn.Query(`SELECT e.id, e.title, e.username, e.password, e.url, e.notes, g.name as group_name FROM entries e LEFT JOIN groups g ON e.group_id = g.id ORDER BY e.id`)
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

	_, err = dbConn.Exec(`UPDATE entries SET title=?, username=?, password=?, url=?, notes=?, group_id=(select id from groups where name = ?) WHERE id=?`,
		encTitle, encUser, encPass, encURL, encNotes, e.Group, e.ID)
	return err
}

// DeleteGroup удаляет все записи в группе (если нужно)
func DeleteGroup(dbConn *sql.DB, id int) error {
	_, err := dbConn.Exec(`DELETE FROM groups WHERE id = ?`, id)
	return err
}

func AddGroup(dbConn *sql.DB, name string) error {
	_, err := dbConn.Exec(`INSERT INTO groups (name) VALUES (?)`, name)
	return err
}

func GetGroup(dbConn *sql.DB) (listGroup []models.Groups, err error) {
	rows, err := dbConn.Query(`SELECT * FROM groups`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var group models.Groups
		err = rows.Scan(&group.Id, &group.Name)
		if err != nil {
			err = fmt.Errorf("ошибка сканирования строки: %w", err)
			return
		}
		listGroup = append(listGroup, group)
	}

	if err = rows.Err(); err != nil {
		err = fmt.Errorf("ошибка при обходе строк: %w", err)
		return
	}
	return
}
