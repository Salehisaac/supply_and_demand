package models

import (
    "time"
    "database/sql"
)

type Request struct {
    ID           int64  
    ResponserID  int64
    CustomerID   int64
    CategoryID   int
    Text         string
    CreatedAt    time.Time
    DeletedAt    sql.NullTime
    TrackingCode string
    Status       string
    RequestID    *int64
    UpdatedAt    *time.Time
}