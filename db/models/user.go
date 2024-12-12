package models

type User struct {
	Id      int    `db:"id"`
	Email   string `db:"email"`
	Refresh string `db:"refresh"`
}
