package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (s *dbstorage) PutURLsBatch(ctx context.Context, uid string, store map[string][]string) error {
	// шаг 1 — объявляем транзакцию
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	// шаг 1.1 — если возникает ошибка, откатываем изменения
	defer tx.Rollback()

	// шаг 2 — готовим инструкцию
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO users_info (uid, short_link, original_url) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	// шаг 2.1 — не забываем закрыть инструкцию, когда она больше не нужна
	defer stmt.Close()

	fmt.Println("bdStore - store", store)

	for _, value := range store {
		// шаг 3 — указываем, что каждая запись будет добавлена в транзакцию
		if _, err = stmt.ExecContext(ctx, uid, value[0], value[1]); err != nil {
			return err
		}
	}

	// шаг 4 — сохраняем изменения
	return tx.Commit()
}
