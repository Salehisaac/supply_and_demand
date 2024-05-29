package repositories

import (
	"database/sql"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
    "fmt"
    "strings"
)



func GetCategoryByName(db *sql.DB, Name string) (models.Category, error) {
    var category models.Category

    
    query := "SELECT id, name, is_active, inactive_message, created_at, updated_at, deleted_at FROM users WHERE name = ?"

    
    err := db.QueryRow(query, Name).Scan(&category.ID, &category.Name,&category.Is_active, &category.Inactive_message, &category.Created_at,&category.Updated_at,&category.Deleted_at )
    if err != nil {
        return category, err
    }

    return category, nil
}

func GetCategoryByID(db *sql.DB, ID int) (models.Category, error) {
    var category models.Category

    
    query := "SELECT id, name, is_active, inactive_message, created_at, updated_at, deleted_at FROM categories WHERE id = ?"

    
    err := db.QueryRow(query, ID).Scan(&category.ID, &category.Name,&category.Is_active, &category.Inactive_message, &category.Created_at,&category.Updated_at,&category.Deleted_at )
    if err != nil {
        return category, err
    }

    return category, nil
}


func GetAllCategories(db *sql.DB) ([]models.Category, error) {
    var categories []models.Category

    query := "SELECT id, name, is_active, inactive_message, created_at, updated_at, deleted_at FROM categories"
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var category models.Category
        if err := rows.Scan(&category.ID, &category.Name, &category.Is_active, &category.Inactive_message, &category.Created_at, &category.Updated_at, &category.Deleted_at); err != nil {
            return nil, err
        }
        categories = append(categories, category)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return categories, nil
}

func InsertCategory(db *sql.DB, category models.Category) error {
    
    stmt, err := db.Prepare("INSERT INTO categories (name) VALUES (?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    
    _, err = stmt.Exec(category.Name)
    if err != nil {
        return err
    }

    return nil
}


func GetCategoryMembers(db *sql.DB, categoryID int) ([]models.User, error) {
	query := "SELECT user_id FROM categories_users WHERE category_id = ?"
	userIDs := make([]int, 0)
    var responsers []models.User

	rows, err := db.Query(query, categoryID)
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
        user, err := GetUserByID(db, int64(userID))
        if err != nil {
            return nil, err
        }
        responsers = append(responsers, user)
    }


	return responsers, nil
}

func UpdateCategory(db *sql.DB, category *models.Category, values []interface{}, columns ...string) error {
    
    var setValues []string
    for _, column := range columns {
        setValues = append(setValues, fmt.Sprintf("%s = ?", column))
    }
    setClause := strings.Join(setValues, ", ")
    
   
    query := fmt.Sprintf("UPDATE categories SET %s WHERE id = ?", setClause)
    
    
    values = append(values, category.ID)
    
    
    _, err := db.Exec(query, values...)
    if err != nil {
        return err
    }
    
    return nil
}

func DeleteCategory(db *sql.DB, categoryid int) error {

    _, err := db.Exec("DELETE FROM categories WHERE id = ?", categoryid)
    if err != nil {
        return err
    }
    return nil
}