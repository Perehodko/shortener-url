package DBStorage

import (
	"database/sql"
	"errors"
	"log"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(DBAddress string) *DBStorage {
	s := DBStorage{}
	var err error
	s.db, err = sql.Open("postgres", DBAddress)
	if err != nil {
		//fmt.Errorf("cant't open database: %w", err)
		panic(err)
	}

	_, err = s.db.Exec(`
        CREATE TABLE IF NOT EXISTS users_info (
            id SERIAL PRIMARY KEY,
            uid TEXT,
            short_link TEXT,
            original_url TEXT
        )
    `)
	if err != nil {
		//fmt.Errorf("cant't create table: %w", err)
		panic(err)
	}
	return &s
}

func (s *DBStorage) PutURL(uid, shortLink, urlForCuts string) error {
	_, err := s.db.Exec(`INSERT INTO users_info (uid, short_link, original_url) VALUES ($1, $2, $3)`,
		uid, shortLink, urlForCuts)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (s *DBStorage) GetURL(uid, shortLink string) (string, error) {
	rows, _ := s.db.Query(
		"SELECT uid, original_url FROM users_info WHERE uid=$1 and short_link=$2",
		uid, shortLink)
	err := rows.Err()

	if err != nil {
		return "", err
	}
	defer rows.Close()

	var (
		UID         string
		originalURL string
	)
	for rows.Next() {
		if err := rows.Scan(&UID, &originalURL); err != nil {
			log.Fatal(err)
		}
	}
	if len(originalURL) == 0 {
		return "", errors.New("in DB no shortURL from request")
	} else {
		return originalURL, nil
	}
}

func (s *DBStorage) GetUserURLs(uid string) (map[string]string, error) {
	rows, _ := s.db.Query(
		"SELECT short_link, original_url FROM users_info WHERE uid=$1",
		uid)
	err := rows.Err()

	if err != nil {
		return map[string]string{}, errors.New("in map no shortURL from request")
	}
	defer rows.Close()

	m := make(map[string]string)
	var (
		shortLink   string
		originalUrl string
	)
	for rows.Next() {
		if err := rows.Scan(&shortLink, &originalUrl); err != nil {
			log.Fatal(err)
		}
		m[shortLink] = originalUrl
	}
	return m, nil
}
