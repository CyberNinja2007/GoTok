package controllers

import (
	"context"
	"fmt"

	"simple-gotok/db"
	"simple-gotok/db/models"
)

func GetUser(pg *db.Postgres, ctx context.Context, guid string) (*models.User, error) {
	query := fmt.Sprintf("SELECT * FROM users WHERE id=%s;", guid)

	rows, err := pg.Db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить пользователя с id %s: %w", guid, err)
	}
	defer rows.Close()

	if rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.Email, &user.Refresh); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных пользователя с id %s: %w", guid, err)
		}
		return &user, nil
	}

	return nil, fmt.Errorf("пользователь с ID '%s' не найден", guid)
}

func UpdateRefresh(pg *db.Postgres, ctx context.Context, guid string, refresh string) (*models.User, error) {
	query := fmt.Sprintf("UPDATE users SET refresh=%s WHERE id=%s RETURNING *;", refresh, guid)

	rows, err := pg.Db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("не удалось обновить refresh пользователя с id %s: %w", guid, err)
	}
	defer rows.Close()

	if rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.Email, &user.Refresh); err != nil {
			return nil, fmt.Errorf("ошибка при получении обновленных данных пользователя с id %s: %w", guid, err)
		}
		return &user, nil
	}

	return nil, fmt.Errorf("пользователь с ID '%s' не найден", guid)
}
