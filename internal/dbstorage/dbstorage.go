package dbstorage

import (
	"database/sql"
	"errors"
	"log"
)

type dbstorage struct {
	db *sql.DB
}

func NewDBStorage(DBAddress string) *dbstorage {
	s := dbstorage{}
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

func (s *dbstorage) PutURL(uid, shortLink, urlForCuts string) error {
	_, err := s.db.Exec(`INSERT INTO users_info (uid, short_link, original_url) VALUES ($1, $2, $3)`,
		uid, shortLink, urlForCuts)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (s *dbstorage) GetURL(uid, shortLink string) (string, error) {
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

func (s *dbstorage) GetUserURLs(uid string) (map[string]string, error) {
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
		originalURL string
	)
	for rows.Next() {
		if err := rows.Scan(&shortLink, &originalURL); err != nil {
			log.Fatal(err)
		}
		m[shortLink] = originalURL
	}
	return m, nil
}