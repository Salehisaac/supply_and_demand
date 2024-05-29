package app

import (

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
    "github.com/Salehisaac/Supply-and-Demand.git/internal/database"
    "github.com/Salehisaac/Supply-and-Demand.git/internal/database/repositories"
    "github.com/joho/godotenv"
    "github.com/redis/go-redis/v9"
    "os"
    "fmt"
    "time"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    _"encoding/json"
    "math/rand"
    "context"
    
)

type Bot struct {
    bot *tgbotapi.BotAPI
    rdb  	*redis.Client
}



type Request struct {
    ResponserID  int64
    CustomerID   int64
    CategoryID   int
    Text         string
    CreatedAt    time.Time
    DeletedAt    sql.NullTime
    TrackingCode string
    Status       string
    RequestID    int
    UpdatedAt    time.Time
}

var userRequests map[int64][]string
var userStates map[int64]string


const (
    StateWaitingForRequest = "waiting_for_request"
    StateRequestSubmitted  = "request_submitted"
    StateTrackingCode      = "tracking_code"
    AddingToTheRequest = "adding_to_the_request"
    WaitingForEnteringCategoryName = "wait_for_entering_category_name"
    UpdateCategory = "updating_a_category"
    WaitingForCategoryUpdate = "waiting_for_category_update"
    CategoryList = "list_of_categories"
    DeleteCategory = "delete_category"
)


func NewBot(token string) (*Bot , error) {

	bot, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true

    log.Printf("Authorized on account %s", bot.Self.UserName)


    return &Bot{bot: bot,
        rdb: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
            DB : 1,
		}),
        } , nil
}

func (b *Bot) Start() error {

    //reading datas from .env
    err := godotenv.Load() 
    if err != nil {
        log.Fatalf("Error loading .env file")
    }
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")

    //conecting to database
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        dbUser, dbPassword, dbHost, dbPort, dbName)

    if err := database.ConnectDB(dsn); err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }
    
    //get an instance from database
    db := database.GetDB()
    userRequests = make(map[int64][]string)
    userStates = make(map[int64]string)
   
	u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := b.bot.GetUpdatesChan(u)

    for update := range updates {
		go handleUpdate(update, db, b)
	}

    
    return nil
}

func handleUpdate(update tgbotapi.Update, db *sql.DB, b *Bot) {
	var chatID int64

	if update.Message != nil {
		chatID = update.Message.Chat.ID

		redisKey := fmt.Sprintf("user:%d:signed_in", chatID)
		redisExists, err := b.rdb.Exists(context.Background(), redisKey).Result()
		if err != nil {
			log.Printf("Error checking user sign-in status in Redis: %v", err)
			return
		}

		if redisExists == 0 {
			if err := insertUserIntoDatabase(update.Message.Chat, db); err != nil {
				log.Printf("Error inserting user into database: %v", err)
				return
			}

			if err := b.rdb.Set(context.Background(), redisKey, true, 0).Err(); err != nil {
				log.Printf("Error setting user sign-in status in Redis: %v", err)
				return
			}
		}

		user, err := repositories.GetUserByChatID(db, chatID)
		if err != nil {
			log.Printf("Error getting user from database: %v", err)
			return
		}

		switch user.Type {
		case "customer":
			HandleCustomerUpdateMessage(update, b, db)
		case "admin":
			HandleAdminUpdateMessage(update, b, db)
		}

	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID

		user, err := repositories.GetUserByChatID(db, chatID)
		if err != nil {
			log.Printf("Error getting user from database: %v", err)
			return
		}

		switch user.Type {
		case "customer":
			HandleCustomerUpdateCallBack(update, b, db)
		case "admin":
            HandleAdminUpdateCallBack(update, b, db)
		}
	}
}

func insertUserIntoDatabase(chat *tgbotapi.Chat, db *sql.DB) error {
	query := `
		INSERT INTO users (chatId, username, name)
		VALUES (?, ?, ?)
	`
	_, err := db.Exec(query, chat.ID, chat.UserName, chat.LastName)
	return err
}



func generateTrackingCode() string {
    
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

    
    rng := rand.New(rand.NewSource(time.Now().UnixNano()))

    
    const length = 12

    
    trackingCode := make([]byte, length)

   
    for i := range trackingCode {
        trackingCode[i] = charset[rng.Intn(len(charset))]
    }

    
    return string(trackingCode)
}

func DeleteMessage(b *Bot, chatID int64, messageID int) error {
    msg := tgbotapi.NewDeleteMessage(chatID, messageID)
    _, err := b.bot.Send(msg)
    return err
}

