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
		group_id INTEGER
	);

    CREATE TABLE IF NOT EXISTS groups (
        id integer PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE
    );