package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	todo "github.com/rest_api_gin"
)

type TodoListPostgres struct {
	db *sqlx.DB
}

func NewTodoListPostgres(db *sqlx.DB) *TodoListPostgres {
	return &TodoListPostgres{
		db: db,
	}
}

func (r *TodoListPostgres) Create(userId int, list todo.TodoList) (int, error) {
	tx, err := r.db.Begin()

	if err != nil {
		return 0, err
	}

	var id int

	if err := tx.QueryRow(
		"INSERT INTO todo_lists (title, description) VALUES ($1, $2) RETURNING id",
		list.Title, list.Description,
	).Scan(&id); err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec(
		"INSERT INTO users_lists (user_id, list_id) VALUES ($1, $2)",
		userId, id,
	)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return id, tx.Commit()
}

func (r *TodoListPostgres) GetAll(userId int) ([]todo.TodoList, error) {
	var lists []todo.TodoList

	if err := r.db.Select(
		&lists,
		"SELECT tl.id, tl.title, tl.description FROM todo_lists tl JOIN users_lists ON tl.id = users_lists.list_id WHERE user_id = $1",
		userId,
	); err != nil {
		return nil, err
	}

	return lists, nil
}

func (r *TodoListPostgres) GetById(userId, listId int) (todo.TodoList, error) {

	var list todo.TodoList
	if err := r.db.Get(
		&list,
		"SELECT tl.id, tl.title, tl.description FROM todo_lists tl JOIN users_lists ul ON tl.id = ul.list_id WHERE user_id = $1 AND tl.id = $2",
		userId, listId,
	); err != nil {
		return todo.TodoList{}, err
	}

	return list, nil
}

func (r *TodoListPostgres) Delete(userId, listId int) error {

	if _, err := r.db.Exec(
		"DELETE FROM todo_lists tl USING users_lists ul WHERE tl.id = ul.list_id AND ul.user_id = $1 AND ul.list_id = $2",
		userId, listId,
	); err != nil {
		return err
	}

	return nil
}

func (r *TodoListPostgres) Update(userId, listId int, input todo.UpdateListInput) error {

	setValue := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Title != nil {
		setValue = append(setValue, fmt.Sprintf("title = $%d", argId))
		args = append(args, *input.Title)
		argId++
	}

	if input.Description != nil {
		setValue = append(setValue, fmt.Sprintf("description = $%d", argId))
		args = append(args, *input.Description)
		argId++
	}

	setQueryString := strings.Join(setValue, ", ")

	query := fmt.Sprintf("UPDATE todo_lists tl SET %s FROM users_lists ul WHERE tl.id = ul.list_id AND ul.user_id=$%d AND ul.list_id = $%d",
		setQueryString,
		argId,
		argId+1,
	)
	args = append(args, userId, listId)

	if _, err := r.db.Exec(query, args...); err != nil {
		return err
	}

	return nil
}
