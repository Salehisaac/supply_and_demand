package repositories

import(
	"database/sql"
    "github.com/Salehisaac/Supply-and-Demand.git/internal/models"
)



func GetUserByChatID(db *sql.DB, chatID int64) (models.User, error) {
    var user models.User
    
    query := "SELECT id, username, chatId, type, name, command, is_active, inactive_message, created_at FROM users WHERE chatId = ?"

    
    err := db.QueryRow(query, chatID).Scan(&user.ID, &user.Username, &user.ChatID, &user.Type, &user.Name, &user.Command, &user.Is_active, &user.Inactive_message, &user.Created_at)
    if err != nil {
        return user, err
    }

    return user, nil
}

func GetUserByID(db *sql.DB, id int64) (models.User, error) {
    var user models.User
    
    query := "SELECT id, username, chatId, type, name, command, is_active, inactive_message, created_at FROM users WHERE id = ?"

    
    err := db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.ChatID, &user.Type, &user.Name, &user.Command, &user.Is_active, &user.Inactive_message, &user.Created_at)
    if err != nil {
        return user, err
    }

    return user, nil
}