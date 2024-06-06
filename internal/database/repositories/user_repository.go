package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	 tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
)

func InsertUserIntoDatabase(chat *tgbotapi.Chat, db *sql.DB) error {
	query := `
		INSERT INTO users (chatId, username, name)
		VALUES (?, ?, ?)
	`
	_, err := db.Exec(query, chat.ID, chat.UserName, chat.LastName)
	return err
}

func GetUserByChatID(tx *sql.Tx, db *sql.DB, chatID int64) (models.User, error) {
	var user models.User
	query := "SELECT id, username, chatId, type, name, request_count, is_active, inactive_message, created_at FROM users WHERE chatId = ?"

	var err error
	if tx != nil {
		err = tx.QueryRow(query, chatID).Scan(&user.ID, &user.Username, &user.ChatID, &user.Type, &user.Name, &user.Request_count, &user.Is_active, &user.Inactive_message, &user.Created_at)
	} else {
		err = db.QueryRow(query, chatID).Scan(&user.ID, &user.Username, &user.ChatID, &user.Type, &user.Name, &user.Request_count, &user.Is_active, &user.Inactive_message, &user.Created_at)
	}
	if err != nil {
		return user, err
	}

	return user, nil
}

func GetUserByID(tx *sql.Tx, db *sql.DB, id int64) (models.User, error) {
	var user models.User
	query := "SELECT id, username, chatId, type, name, request_count, is_active, inactive_message, created_at FROM users WHERE id = ?"

	var err error
	if tx != nil {
		err = tx.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.ChatID, &user.Type, &user.Name, &user.Request_count, &user.Is_active, &user.Inactive_message, &user.Created_at)
	} else {
		err = db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.ChatID, &user.Type, &user.Name, &user.Request_count, &user.Is_active, &user.Inactive_message, &user.Created_at)
	}
	if err != nil {
		return user, err
	}

	return user, nil
}

func GetUserByUsername(tx *sql.Tx, db *sql.DB, username string) (models.User, error) {
	var user models.User
	query := "SELECT id, username, chatId, type, name, request_count, is_active, inactive_message, created_at FROM users WHERE username = ?"

	var err error
	if tx != nil {
		err = tx.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.ChatID, &user.Type, &user.Name, &user.Request_count, &user.Is_active, &user.Inactive_message, &user.Created_at)
	} else {
		err = db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.ChatID, &user.Type, &user.Name, &user.Request_count, &user.Is_active, &user.Inactive_message, &user.Created_at)
	}
	if err != nil {
		return user, err
	}

	return user, nil
}

func UserExistsInCategoryCheck(tx *sql.Tx, db *sql.DB, userID int, categoryID int) (bool, error) {
	query := "SELECT id FROM categories_users WHERE category_id = ? AND user_id = ?"
	var id int

	log.Printf("UserExistsInCategoryCheck: categoryID: %d, userID: %d", categoryID, userID)

	var err error
	if tx != nil {
		err = tx.QueryRow(query, categoryID, userID).Scan(&id)
	} else {
		err = db.QueryRow(query, categoryID, userID).Scan(&id)
	}
	log.Printf("UserExistsInCategoryCheck: query result: id: %d, err: %v", id, err)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("UserExistsInCategoryCheck: not found in database")
			return false, nil
		}
		log.Printf("UserExistsInCategoryCheck: error while checking if user exists in category: %v", err)
		return false, err
	}
	return true, nil
}

func AddingMemberToCategory(tx *sql.Tx, db *sql.DB, userID int, categoryID int) error {
	query := "INSERT INTO categories_users (category_id, user_id) VALUES (?, ?)"
	var err error
	if tx != nil {
		_, err = tx.Exec(query, categoryID, userID)
	} else {
		_, err = db.Exec(query, categoryID, userID)
	}
	return err
}

func RemovingMemberFromCategory(tx *sql.Tx, db *sql.DB, userID int, categoryID int) error {
	query := "DELETE FROM categories_users WHERE category_id = ? AND user_id = ?"
	var err error
	if tx != nil {
		_, err = tx.Exec(query, categoryID, userID)
	} else {
		_, err = db.Exec(query, categoryID, userID)
	}
	return err
}

func DeactivingMember(tx *sql.Tx, db *sql.DB, userID int) error {
	query := "UPDATE users SET is_active = 0 WHERE id = ?"
	var err error
	if tx != nil {
		_, err = tx.Exec(query, userID)
        if err != nil {
            log.Printf("Error committing transaction: %v", err)
            return err
        }
	} else {
		_, err = db.Exec(query, userID)
	}
    log.Println("member Updated Succefully")
	return err
}

func ReactivingMember(tx *sql.Tx, db *sql.DB, userID int) error {
	query := "UPDATE users SET is_active = 1 WHERE id = ?"
	var err error
	if tx != nil {
		_, err = tx.Exec(query, userID)
	} else {
		_, err = db.Exec(query, userID)
	}
	return err
}

func UpdateUser(tx *sql.Tx, user models.User, values []interface{}, columns ...string) error {
	var setValues []string
	for _, column := range columns {
		setValues = append(setValues, fmt.Sprintf("%s = ?", column))
	}
	setClause := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", setClause)
	values = append(values, user.ID)

	_, err := tx.Exec(query, values...)
    if err != nil {
        log.Printf("Error executing statement in transaction: %v", err)
        return err
    }

    log.Println("User Updated successfully")
	return nil
}

func GetUserCategories(tx *sql.Tx, db *sql.DB, userID int) ([]models.Category, error) {
	query := "SELECT category_id FROM categories_users WHERE user_id = ?"
	categoryIDs := make([]int, 0)
	var userCategories []models.Category

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(query, userID)
	} else {
		rows, err = db.Query(query, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var categoryID int
		if err := rows.Scan(&categoryID); err != nil {
			return nil, err
		}
		categoryIDs = append(categoryIDs, categoryID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, categoryID := range categoryIDs {
		category, err := GetCategoryByID(tx, db, categoryID)
		if err != nil {
			return nil, err
		}
		userCategories = append(userCategories, category)
	}

	return userCategories, nil
}

func GetAllUnseenTheResponderRequests(tx *sql.Tx, responderID int) ([]models.Request, error) {
	var requests []models.Request
	query := "SELECT * FROM requests WHERE responser_id = ? AND status = 'unseen' AND request_id IS NULL"

	rows, err := tx.Query(query, responderID)
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

func GetCategoryByID(tx *sql.Tx, db *sql.DB, categoryID int) (models.Category, error) {
	var category models.Category
	query := "SELECT id, name, is_active, inactive_message, created_at, updated_at, deleted_at, contain_active_responser FROM categories WHERE id = ?"

	var err error
	if tx != nil {
		err = tx.QueryRow(query, categoryID).Scan(&category.ID, &category.Name, &category.Is_active, &category.Inactive_message, &category.Created_at, &category.Updated_at, &category.Deleted_at, &category.Contain_active_responser)
	} else {
		err = db.QueryRow(query, categoryID).Scan(&category.ID, &category.Name, &category.Is_active, &category.Inactive_message, &category.Created_at, &category.Updated_at, &category.Deleted_at, &category.Contain_active_responser)
	}
	if err != nil {
		return category, err
	}

	return category, nil
}
