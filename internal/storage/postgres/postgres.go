package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/foreground-eclipse/url-shortener/internal/storage"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

// New initializing new database connection
func New() (*Storage, error) {
	const op = "storage.postgres.postgres"
	// docker run --name shortenerdb -e POSTGRES_PASSWORD=shortener -p 5432:5432 -d postgres
	// docker run -d --name shortenerdb -e POSTGRES_USER=myuser -e POSTGRES_PASSWORD=shortener -e POSTGRES_DB=mydatabase -p 5432:5432 postgres
	connStr := "user=myuser dbname=mydatabase password=shortener port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Init() error {
	if err := s.createUrlTable(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) createUrlTable() error {
	const op = "storage.postgres.createUrlTable"

	query := `create table  if not exists url(
		id serial primary key,
		url varchar(100) NOT NULL UNIQUE,
		alias varchar(100) NOT NULL UNIQUE
		);
		create index if not exists idx_alias on url(alias);`

	_, err := s.db.Exec(query)
	return err
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	lastInsertId := 0

	query := `insert into url 
	(url, alias) values ($1, $2) RETURNING id`

	// res, err := s.db.Query(query, urlToSave, alias)

	err := s.db.QueryRow(query, urlToSave, alias).Scan(&lastInsertId)

	if err != nil {

		if postgresErr := err.(*pq.Error); postgresErr.Code == "23505" {
			return 0, fmt.Errorf("%s, %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s, %w", op, err)
	}

	return int64(lastInsertId), nil

}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	var selectedAlias string
	err := s.db.QueryRow("select url from url where alias = $1", alias).Scan(&selectedAlias)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return selectedAlias, nil

}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	_, err := s.db.Query("delete from url where alias = $1", alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) CheckIfAliasExists(alias string) (bool, error) {
	const op = "storage.postgres.CheckAlias"

	var exists bool
	err := s.db.QueryRow("select exists (select 1 from url where alias = $1)", alias).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return exists, nil
}
