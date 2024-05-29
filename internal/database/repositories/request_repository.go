package repositories

import(
	"database/sql"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
    "errors"
    "fmt"
    "strings"
)



func InsertRequest(db *sql.DB, request models.Request) error {
    
    stmt, err := db.Prepare("INSERT INTO requests (responser_id, customer_id, category_id, text,tracking_code, status, request_id) VALUES (?, ?, ?, ?, ?, ?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    
    _, err = stmt.Exec(request.ResponserID, request.CustomerID, request.CategoryID, request.Text, request.TrackingCode, request.Status, request.RequestID )
    if err != nil {
        return err
    }

    return nil
}

func GetLastFiveRequests(db *sql.DB , id int) ([]models.Request, error) {
    
    var requests []models.Request

    rows, err := db.Query("SELECT * FROM requests WHERE customer_id = ? AND status = ? ORDER BY created_at DESC LIMIT 5", id, "پاسخ داده نشده")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var request models.Request
        if err := rows.Scan(&request.ID,&request.ResponserID, &request.CustomerID, &request.CategoryID, &request.Text, &request.CreatedAt, &request.DeletedAt, &request.TrackingCode, &request.Status, &request.RequestID, &request.UpdatedAt ); err != nil {
            return nil, err
        }
        requests = append(requests, request)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return requests, nil
}
func GetRequestByTrackingCode(db *sql.DB, trackingCode string) (*models.Request, error) {
    var request models.Request
    
    query := "SELECT * FROM requests WHERE tracking_code = ?"
    err := db.QueryRow(query, trackingCode).Scan(&request.ID,&request.ResponserID, &request.CustomerID, &request.CategoryID, &request.Text, &request.CreatedAt, &request.DeletedAt, &request.TrackingCode, &request.Status, &request.RequestID, &request.UpdatedAt)
    
    if err == sql.ErrNoRows {
       
        return nil, errors.New("no request found for the given tracking code")
    } else if err != nil {
        
        return nil, err
    }

    return &request, nil
}

func GetRequestByID(db *sql.DB, id int) (*models.Request, error) {
    var request models.Request
    
    query := "SELECT * FROM requests WHERE id = ?"
    err := db.QueryRow(query, id).Scan(&request.ID,&request.ResponserID, &request.CustomerID, &request.CategoryID, &request.Text, &request.CreatedAt, &request.DeletedAt, &request.TrackingCode, &request.Status, &request.RequestID, &request.UpdatedAt)
    
    if err == sql.ErrNoRows {
       
        return nil, errors.New("no request found for the given  ID")
    } else if err != nil {
        
        return nil, err
    }

    return &request, nil
}

func UpdateRequest(db *sql.DB, request *models.Request, values []interface{}, columns ...string) error {
    
    var setValues []string
    for _, column := range columns {
        setValues = append(setValues, fmt.Sprintf("%s = ?", column))
    }
    setClause := strings.Join(setValues, ", ")
    
   
    query := fmt.Sprintf("UPDATE requests SET %s WHERE id = ?", setClause)
    
    
    values = append(values, request.ID)
    
    
    _, err := db.Exec(query, values...)
    if err != nil {
        return err
    }
    
    return nil
}