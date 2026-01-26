package domain

import "time"

type Professor struct {
	ID        int       `db:"id" json:"id"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	Email     *string   `db:"email" json:"email"`
	Rating    *float64  `db:"rating" json:"rating"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
