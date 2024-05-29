package models

import (
    "time"
	"database/sql"
    
)

type User struct {
    ID       int
    Username string
    ChatID   string
    Type     string
    Name     string
    Command  sql.NullString
    Is_active bool
    Inactive_message sql.NullString
    Created_at *time.Time
}