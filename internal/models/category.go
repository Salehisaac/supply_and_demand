package models

import (
    "time"
	"database/sql"
    
)

type Category struct {
    ID       int
    Name     string
    Is_active bool
    Inactive_message sql.NullString
    Created_at *time.Time
    Updated_at *time.Time
    Deleted_at *time.Time
}