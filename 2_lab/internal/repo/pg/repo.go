package pg

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Repo struct {
	db *sql.DB
}

func New(dsn string) (*Repo, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Repo{db: db}, nil
}

func (r *Repo) Migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS notes (
			id    SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			body  TEXT NOT NULL
		)
	`)
	return err
}

func (r *Repo) GetAll(ctx context.Context) ([]Note, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, body FROM notes ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *Repo) Create(ctx context.Context, title, body string) (Note, error) {
	var n Note
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO notes (title, body) VALUES ($1, $2) RETURNING id, title, body",
		title, body,
	).Scan(&n.ID, &n.Title, &n.Body)
	return n, err
}

func (r *Repo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
