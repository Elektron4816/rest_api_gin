package repository

import (
	"github.com/jmoiron/sqlx"
	todo "github.com/rest_api_gin"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{
		db: db,
	}
}

func (r *AuthPostgres) CreateUser(user todo.User) (int, error) {
	var id int

	if err := r.db.QueryRow(
		"INSERT INTO users (name, username, password_hash) VALUES ($1, $2, $3) RETURNING id",
		user.Name, user.UserName, user.Password,
	).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *AuthPostgres) GetUser(username, password string) (todo.User, error) {
	var user todo.User
	if err := r.db.QueryRow(
		"SELECT id FROM users WHERE username = $1 AND password_hash = $2",
		username, password,
	).Scan(&user.Id); err != nil {
		return todo.User{}, err
	}

	return user, nil
}
