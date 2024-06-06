package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
)

func InsertAnswer(tx *sql.Tx,db *sql.DB , answer models.Answer) error {
	stmt, err := tx.Prepare("INSERT INTO answers (responser_id, customer_id, text, tracking_code, request_id, created_at , updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(answer.ResponserID, answer.CustomerID, answer.Text, answer.TrackingCode, answer.RequestID, time.Now() , time.Now())
	if err != nil {
		return err
	}

	responder, err := GetUserByID(tx, db, answer.ResponserID)
	if err != nil {
		return err
	}



	newResponderRequestCount := responder.Request_count - 1
	values := []interface{}{newResponderRequestCount}
	columns := []string{"request_count"}
	err = UpdateUser(tx, responder, values, columns...)
	if err != nil {
		return err
	}

	log.Println("answer inserted succusfully")

	return nil
}

func GetResponseStatistics(tx *sql.Tx, responderID int64) (string, error) {
	answers, err := GetAnswersByResponderID(tx, responderID)
	if err != nil {
		return "", err
	}

	var totalCount, todayCount, lastMonthCount int



	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfLastMonth := now.AddDate(0, -1, 0)

	for _, answer := range answers {
		totalCount++
		if answer.CreatedAt.After(startOfToday) {
			todayCount++
		}
		if answer.CreatedAt.After(startOfLastMonth) {
			lastMonthCount++
		}
	}
	

	result := fmt.Sprintf(
		"تعداد پاسخگویی امروز: %d\nتعداد پاسخگویی یک ماه گذشته: %d\nتعداد پاسخگویی از ابتدا تا کنون: %d",
		todayCount, lastMonthCount, totalCount,
	)

	return result, nil
}


func GetAnswersByResponderID(tx *sql.Tx, responderID int64) ([]models.Answer, error) {
	var answers []models.Answer

	query := "SELECT id, responser_id, customer_id, text, created_at, deleted_at, tracking_code, request_id, updated_at FROM answers WHERE responser_id = ?"
	log.Printf("Executing query: %s with responderID: %d", query, responderID)
	rows, err := tx.Query(query, responderID)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var answer models.Answer
        if err := rows.Scan(&answer.ID, &answer.ResponserID, &answer.CustomerID, &answer.Text, &answer.CreatedAt, &answer.DeletedAt, &answer.TrackingCode, &answer.RequestID, &answer.UpdatedAt); err != nil {
            log.Printf("Error scanning row: %v", err)
            return nil, err
        }
		log.Printf("Scanned answer: %+v", answer)
		answers = append(answers, answer)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error with rows: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d answers", len(answers))
	return answers, nil
}

