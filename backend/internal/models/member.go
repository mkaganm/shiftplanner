package models

import (
	"time"
)

// Member takım üyesi modeli
type Member struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

