package database

import(
	"log"
	"database/sql"	
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func ConnectDB(connectionString string) error {
    database, err := sql.Open("mysql", connectionString)
    if err != nil {
        return err
    }

    if err := database.Ping(); err != nil {
        return err
    }

    log.Println("Connected to the database")

    db = database

    return nil
}

func GetDB() *sql.DB {
    return db
}