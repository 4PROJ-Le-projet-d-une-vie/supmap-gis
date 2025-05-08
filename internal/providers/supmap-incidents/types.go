package supmap_incidents

import "time"

type Incident struct {
	ID        int64      `json:"id"`
	User      *User      `json:"user"`
	Type      *Type      `json:"type"`
	Latitude  float64    `json:"lat"`
	Longitude float64    `json:"lon"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Distance  float64    `json:"distance"`
}

type User struct {
	ID     int64  `json:"id"`
	Handle string `json:"handle"`
	Role   *Role  `json:"role"`
}

type Role struct {
	Name string `json:"name"`
}

type Type struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
