package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
    "log"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
)

func InsertRequest(tx *sql.Tx, db *sql.DB, request models.Request) error {
	stmt, err := getPreparedStmt(tx, db, "INSERT INTO requests (responser_id, customer_id, category_id, text,tracking_code, status, request_id) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(request.ResponserID, request.CustomerID, request.CategoryID, request.Text, request.TrackingCode, request.Status, request.RequestID)
	if err != nil {
		return err
	}

	responder, err := GetUserByID(tx, db, request.ResponserID)
	if err != nil {
		return err
	}

	newResponderRequestCount := responder.Request_count + 1
	values := []interface{}{newResponderRequestCount}
	columns := []string{"request_count"}
	err = UpdateUser(tx, responder, values, columns...)
	if err != nil {
		return err
	}

	return nil
}

func GetLastFiveRequests(db *sql.DB, id int) ([]models.Request, error) {
	var requests []models.Request

	rows, err := db.Query("SELECT * FROM requests WHERE customer_id = ? ORDER BY created_at DESC LIMIT 5", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var request models.Request
		if err := rows.Scan(&request.ID, &request.ResponserID, &request.CustomerID, &request.CategoryID, &request.Text, &request.CreatedAt, &request.DeletedAt, &request.TrackingCode, &request.Status, &request.RequestID, &request.UpdatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func GetRequestByTrackingCode(tx *sql.Tx, db *sql.DB, trackingCode string) (*models.Request, error) {
	var request models.Request

	query := "SELECT * FROM requests WHERE tracking_code = ?"
	err := tx.QueryRow(query, trackingCode).Scan(&request.ID, &request.ResponserID, &request.CustomerID, &request.CategoryID, &request.Text, &request.CreatedAt, &request.DeletedAt, &request.TrackingCode, &request.Status, &request.RequestID, &request.UpdatedAt)

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
	err := db.QueryRow(query, id).Scan(&request.ID, &request.ResponserID, &request.CustomerID, &request.CategoryID, &request.Text, &request.CreatedAt, &request.DeletedAt, &request.TrackingCode, &request.Status, &request.RequestID, &request.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("no request found for the given  ID")
	} else if err != nil {
		return nil, err
	}

	return &request, nil
}

func UpdateRequest(tx *sql.Tx, db *sql.DB, request *models.Request, values []interface{}, columns ...string) error {
	var setValues []string
	for _, column := range columns {
		setValues = append(setValues, fmt.Sprintf("%s = ?", column))
	}
	setClause := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE requests SET %s WHERE id = ?", setClause)

	values = append(values, request.ID)

	_, err := tx.Exec(query, values...)
    if err != nil {
        log.Printf("Error executing statement in transaction: %v", err)
        return err
    }


    log.Println("Request Updated successfully")
	return nil
}

func GetRequestAnswers(tx *sql.Tx, requestID int64) ([]models.Answer, error) {
	var answers []models.Answer

	query := "SELECT id, responser_id, customer_id, text, created_at, deleted_at, tracking_code, request_id, updated_at FROM answers WHERE request_id = ?"
	rows, err := tx.Query(query, requestID)
	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var answer models.Answer
		if err := rows.Scan(&answer.ID, &answer.ResponserID, &answer.CustomerID, &answer.Text, &answer.CreatedAt, &answer.DeletedAt, &answer.TrackingCode, &answer.RequestID, &answer.UpdatedAt); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		answers = append(answers, answer)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return nil, err
	}

	log.Printf("Retrieved %d answers for request ID: %d", len(answers), requestID)
	return answers, nil
}


func getPreparedStmt(tx *sql.Tx, db *sql.DB, query string) (*sql.Stmt, error) {
	if tx != nil {
		return tx.Prepare(query)
	}
	return db.Prepare(query)
}
