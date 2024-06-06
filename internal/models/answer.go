package models

import (
    "time"
)

type Answer struct {
    ID           int64      `db:"id"`
    ResponserID  int64      `db:"responser_id"`
    CustomerID   int64      `db:"customer_id"`
    Text         string     `db:"text"`
    CreatedAt    time.Time  `db:"created_at"`
    DeletedAt    *time.Time `db:"deleted_at"`
    TrackingCode string     `db:"tracking_code"`
    RequestID    int64      `db:"request_id"`
    UpdatedAt    time.Time  `db:"updated_at"`
}