package repositories

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
)


func GetCategoryByName(tx *sql.Tx, db *sql.DB, name string) (models.Category, error) {
	var category models.Category
	query := "SELECT id, name, is_active, inactive_message, created_at, updated_at, deleted_at, contain_active_responser FROM categories WHERE name = ?"

	var err error
	if tx != nil {
		err = tx.QueryRow(query, name).Scan(&category.ID, &category.Name, &category.Is_active, &category.Inactive_message, &category.Created_at, &category.Updated_at, &category.Deleted_at, &category.Contain_active_responser)
	} else {
		err = db.QueryRow(query, name).Scan(&category.ID, &category.Name, &category.Is_active, &category.Inactive_message, &category.Created_at, &category.Updated_at, &category.Deleted_at, &category.Contain_active_responser)
	}

	return category, err
}

func GetAllCategories(tx *sql.Tx, db *sql.DB) ([]models.Category, error) {
	var categories []models.Category
	query := "SELECT id, name, is_active, inactive_message, created_at, updated_at, deleted_at, contain_active_responser FROM categories"

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(query)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Is_active, &category.Inactive_message, &category.Created_at, &category.Updated_at, &category.Deleted_at, &category.Contain_active_responser); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func GetActiveCategories(tx *sql.Tx, db *sql.DB) ([]models.Category, error) {
	var categories []models.Category
	query := "SELECT id, name, is_active, inactive_message, created_at, updated_at, deleted_at, contain_active_responser FROM categories WHERE is_active = 1"

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(query)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Is_active, &category.Inactive_message, &category.Created_at, &category.Updated_at, &category.Deleted_at, &category.Contain_active_responser); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func GetInactiveCategories(tx *sql.Tx, db *sql.DB) ([]models.Category, error) {
	var categories []models.Category
	query := "SELECT id, name, is_active, inactive_message, created_at, updated_at, deleted_at, contain_active_responser FROM categories WHERE is_active = 0 AND contain_active_responser = 1"

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(query)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Is_active, &category.Inactive_message, &category.Created_at, &category.Updated_at, &category.Deleted_at, &category.Contain_active_responser); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func InsertCategory(tx *sql.Tx, db *sql.DB, category models.Category) error {
	query := "INSERT INTO categories (name, inactive_message) VALUES (?, ?)"
	var stmt *sql.Stmt
	var err error

	if tx != nil {
		stmt, err = tx.Prepare(query)
		if err != nil {
			log.Printf("Error preparing statement in transaction: %v", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(category.Name, category.Inactive_message)
		if err != nil {
			log.Printf("Error executing statement in transaction: %v", err)
			return err
		}

	} else {
		stmt, err = db.Prepare(query)
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(category.Name, category.Inactive_message)
		if err != nil {
			log.Printf("Error executing statement: %v", err)
			return err
		}
	}

	log.Println("Category inserted successfully")
	return nil
}

func GetCategoryMembers(tx *sql.Tx, db *sql.DB, categoryID int) ([]models.User, error) {
	query := "SELECT user_id FROM categories_users WHERE category_id = ?"
	userIDs := make([]int, 0)
	var responsers []models.User

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(query, categoryID)
	} else {
		rows, err = db.Query(query, categoryID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, userID := range userIDs {
		user, err := GetUserByID(tx, db, int64(userID))
		if err != nil {
			return nil, err
		}
		responsers = append(responsers, user)
	}

	return responsers, nil
}

func GetActiveCategoryMembers(tx *sql.Tx , db *sql.DB, categoryID int) ([]models.User, error) {
	query := "SELECT user_id FROM categories_users WHERE category_id = ?"
	userIDs := make([]int, 0)
	var responsers []models.User

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(query, categoryID)
	} else {
		rows, err = db.Query(query, categoryID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, userID := range userIDs {
		user, err := GetUserByID(tx, db, int64(userID))
		if err != nil {
			return nil, err
		}
		if user.Is_active {
			responsers = append(responsers, user)
		}
	}

	return responsers, nil
}

func GetInactiveCategoryMembers(tx *sql.Tx, db *sql.DB, categoryID int) ([]models.User, error) {
	query := "SELECT user_id FROM categories_users WHERE category_id = ?"
	userIDs := make([]int, 0)
	var responsers []models.User

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(query, categoryID)
	} else {
		rows, err = db.Query(query, categoryID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, userID := range userIDs {
		user, err := GetUserByID(tx, db, int64(userID))
		if err != nil {
			return nil, err
		}
		if !user.Is_active {
			responsers = append(responsers, user)
		}
	}

	return responsers, nil
}

func UpdateCategory(tx *sql.Tx, db *sql.DB, category models.Category, values []interface{}, columns ...string) error {
	var setValues []string
	for _, column := range columns {
		setValues = append(setValues, fmt.Sprintf("%s = ?", column))
	}
	setClause := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE categories SET %s WHERE id = ?", setClause)
	values = append(values, category.ID)

	var err error
	if tx != nil {
		_, err = tx.Exec(query, values...)
		if err != nil {
			log.Printf("Error executing update in transaction: %v", err)
			return err
		}

	} else {
		_, err = db.Exec(query, values...)
		if err != nil {
			log.Printf("Error executing update: %v", err)
			return err
		}
	}

	log.Println("Category updated successfully")
	return nil
}

func DeleteCategory(tx *sql.Tx, db *sql.DB, categoryID int) error {
	query := "DELETE FROM categories WHERE id = ?"
	var err error
	if tx != nil {
		_, err = tx.Exec(query, categoryID)
        if err != nil {
			log.Printf("Error executing update in transaction: %v", err)
			return err
		}
	} else {
		_, err = db.Exec(query, categoryID)
        if err != nil {
			log.Printf("Error executing update: %v", err)
			return err
		}
	}
    log.Println("Category deleted successfully")
	return nil
	
}

func GetCategoryRequests(tx *sql.Tx, db *sql.DB, categoryID int) ([]models.Request, error) {
	var requests []models.Request
	query := "SELECT * FROM requests WHERE category_id = ?"

	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(query, categoryID)
	} else {
		rows, err = db.Query(query, categoryID)
	}
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

func GetResponserWithLowestRequest(tx *sql.Tx, db *sql.DB, categoryID int) (models.User, error) {
	var user models.User

	members, err := GetActiveCategoryMembers(tx,db , categoryID)
	if err != nil {
		return user, err
	}

	if len(members) == 1 {
		user = members[0]
	} else {
		for i := 0; i < len(members)-1; i++ {
			if members[i].Request_count < members[i+1].Request_count {
				user = members[i]
			} else {
				user = members[i+1]
			}
		}
	}

	return user, nil
}


