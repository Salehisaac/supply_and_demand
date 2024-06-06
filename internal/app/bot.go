package app

import (
	"context"
	"database/sql"
	_ "encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/Salehisaac/Supply-and-Demand.git/internal/database"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/database/repositories"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
    "github.com/Salehisaac/Supply-and-Demand.git/internal/models"
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

//states
const (
    //Bot states
    WaitingForTheToken = "waiiting for main admin to enter the given token"
    //Customer states
    StateWaitingForRequest = "waiiting for user to enter requests and submit them"
    StateRequestSubmitted  = "request submited"
    StateTrackingCode      = "waiiting for user to enter the tracking code of wanted request"
    AddingToTheRequest = "adding a request to another request"
    //Admin states
    WaitingForEnteringCategoryName = "waitiing for admin to enter the category name"
    UpdateCategory = "updating a category"
    WaitingForCategoryUpdate = "waiting for category update"
    CategoryList = "list of categories"
    DeleteCategory = "deleting a category"
    AddingMemberToCategory = "adding a member to a category"
    WaitingForUserNameToAdd = "waiting for admin to enter username to add"
    WaitingForUserNameToRemove = "waiting for admin to enter username to remove"
    RemovingMemberToCategory = "removing member from category"
    WaitingForBotInactiveMessage = "waiting for admin to enter inactive message for bot"
    WaitingForCategoryInactiveMessage = "waiting for admin to enter inactive message for category"
    DeactiveCategory = "deactive a category"
    ReactiveCategory = "reactive a category"
    DeactiveCategoryMember = "deactive a category member"
    ReactiveCategoryMember = "reactive a category member"
    WaitingForCategoryMemberUsernameToDeactive = "waiting for admin to enter member username to deactive"
    WaitingForCategoryMemberUsernameToReactive = "waiting for admin to enter member username to reactive"
    //Responder states
    WaittingForRequestReply = "waiiting for responder to replay"
)

const WelcomeMessage =`سلام! 👋

به ربات تلگرام ما خوش آمدید! 🤖

ما اینجا هستیم تا درخواست‌های  شما را دریافت و پیگیری کنیم. 🏗


اگر نیاز به کمک دارید، با پشتیبانی تماس بگیرید. 📞

با تشکر از شما، تیم ما آماده خدمت رسانی به شماست! 😊`



func NewBot() (*Bot , error) {

    //reading the botToken from .env
    err := godotenv.Load() 
    if err != nil {
        log.Println("Error loading .env file")
    }
    token := os.Getenv("BOT_TOKEN")

    //creating a new instance of tgbotapi
	bot, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        log.Println(err)
    }

    bot.Debug = true

    log.Printf("Authorized on account %s", bot.Self.UserName)

    //creating the bot struct and configuring the redis
    return &Bot{bot: bot,
        rdb: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
            DB : 0,
		}),
        } , nil
}

func (b *Bot) Start() error {

    //reading datas from .env
    err := godotenv.Load() 
    if err != nil {
        log.Println("Error loading .env file")
    }
    dbUser := os.Getenv("MYSQL_USER")
    dbPassword := os.Getenv("MYSQL_PASSWORD")
    dbHost := os.Getenv("MYSQL_HOST")
    dbPort := os.Getenv("MYSQL_PORT")
    dbName := os.Getenv("MYSQL_DB")
    ChatID := os.Getenv("CHAT_ID")

    chatid_int,_ := strconv.Atoi(ChatID)

    //conecting to database
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        dbUser, dbPassword, dbHost, dbPort, dbName)

    if err := database.ConnectDB(dsn); err != nil {
        log.Printf("Error connecting to database: %v", err)
    }
    
    //get an instance from database
    db := database.GetDB()



    //creating maps for reciving in app datas
    userRequests = make(map[int64][]string)
    userStates = make(map[int64]string)
   
	u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    //getting the updates every 60 seconds
    updates := b.bot.GetUpdatesChan(u)

    //handeling updates
    for update := range updates {

        //waitting for user to enter the token for initialization
        if userStates[int64(chatid_int)] == WaitingForTheToken{
            InitializingTheBot(b, db, int64(chatid_int) , update)

        }else{
            //check if the bot is initialazed or not 
            if CheckTheBotInstallation(b , db , int64(chatid_int)) {
                go handleUpdate(update, db, b)
            }
        }   
	} 
    return nil
}

func handleUpdate(update tgbotapi.Update, db *sql.DB, b *Bot) {

    //begin the transaction
    tx, err := db.Begin()
    if err != nil {
        log.Println("Error beginning transaction:", err)
        return
    }
    //ensure thet the transaction gonna close
    defer func() {
        if err := recover(); err != nil {
            log.Println("Panic occurred, rolling back transaction:", err)
            if err := tx.Rollback(); err != nil {
                log.Println("Error rolling back transaction:", err)
            }
        }
    }()

	var chatID int64
    //get the bot configs
    bot_settings , err := repositories.GetTheBotConfigs(db)
        

    if err != nil {
        log.Println(err)
    }
    
	if update.Message != nil {
		chatID = update.Message.Chat.ID
        //check if the user already exists or not
		redisKey := fmt.Sprintf("user:%d:signed_in", chatID)
		redisExists, err := b.rdb.Exists(context.Background(), redisKey).Result()
		if err != nil {
			log.Printf("Error checking user sign-in status in Redis: %v", err)
			return
		}

        //inserting user into database
		if redisExists == 0 {
			if err := repositories.InsertUserIntoDatabase(update.Message.Chat, db); err != nil {
				log.Printf("Error inserting user into database: %v", err)
				return
			}

			if err := b.rdb.Set(context.Background(), redisKey, true, 0).Err(); err != nil {
				log.Printf("Error setting user sign-in status in Redis: %v", err)
				return
			}
		}

		user, err := repositories.GetUserByChatID(tx ,db, chatID)
		if err != nil {
			log.Printf("Error getting user from database: %v", err)
			return
		}

      

     
		switch user.Type {
		case "customer":
            //show inActive message if bot was in active
            if !bot_settings.Is_active {
                msg := tgbotapi.NewMessage(update.Message.Chat.ID, bot_settings.Inactive_message.String)
                if _, err := b.bot.Send(msg); err != nil{
                    log.Println(err)
                }
                
            }else{
                HandleCustomerUpdateMessage(update, b, db)
            }
			
		case "admin":
			HandleAdminUpdateMessage(update, b, db)
        case "responder":
			HandleResponderUpdateMessage(update, b, db)
		}
        

	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID

		user, err := repositories.GetUserByChatID(tx ,db, chatID)
		if err != nil {
			log.Printf("Error getting user from database: %v", err)
			return
		}

		switch user.Type {
		case "customer":
            //show inActive message if bot was in active
            if !bot_settings.Is_active {
                msg := tgbotapi.NewMessage(update.Message.Chat.ID, bot_settings.Inactive_message.String)
                if _, err := b.bot.Send(msg); err != nil{
                    log.Println(err)
                }
                
            }else{
                HandleCustomerUpdateCallBack(update, b, db)
            }
			
		case "admin":
            HandleAdminUpdateCallBack(update, b, db)
        case "responder":
            HandleResponderUpdateCallBack(update, b, db)
		}
	}
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

func CheckTheBotInstallation(b *Bot , db *sql.DB , chatID int64 ) bool {
    _ , err := repositories.GetTheBotConfigs(db)
    msg := tgbotapi.NewMessage(chatID, "")
    if err != nil {
        text := "توکن داده شده را وارد کنید"
		msg.Text = text
        _ , err := b.bot.Send(msg)
        if err != nil {
            log.Println(err)
        }
        userStates[chatID] = WaitingForTheToken
        return false
    }
    return true
}

func InitializingTheBot(b *Bot , db *sql.DB , chatID int64 , update tgbotapi.Update){

    //begin the transaction
    tx, err := db.Begin()
    if err != nil {
        log.Println("Error beginning transaction:", err)
        return
    }
    //ensure that the transaction gonna close
    defer func() {
        if err := recover(); err != nil {
            log.Println("Panic occurred, rolling back transaction:", err)
            if err := tx.Rollback(); err != nil {
                log.Println("Error rolling back transaction:", err)
            }
        }
    }()

    msg := tgbotapi.NewMessage(chatID, "")

    //input token
    user_token := update.Message.Text
    //actual token
    GetTokenRedisKey := "confirmation_token"
    actual_token,err := b.rdb.Get(context.Background(),GetTokenRedisKey).Result()

    if err != nil {
        log.Printf("Error: %v", err)
    }
    
    log.Println("actual : " , actual_token)
    log.Println("user_token : " , user_token)
    //checking the tokens
    if user_token == actual_token {

        redisKey := fmt.Sprintf("user:%d:signed_in", chatID)
        redisExists, err := b.rdb.Exists(context.Background(), redisKey).Result()
        if err != nil {
            log.Printf("Error checking user sign-in status in Redis: %v", err)
        }

        if redisExists == 0 {
            if err := repositories.InsertUserIntoDatabase(update.Message.Chat, db); err != nil {
                log.Printf("Error inserting user into database: %v", err)
            }

            if err := b.rdb.Set(context.Background(), redisKey, true, 0).Err(); err != nil {
                log.Printf("Error setting user sign-in status in Redis: %v", err)
            }
        }

        user , err := repositories.GetUserByChatID(tx ,db , chatID)
        if err != nil{
            log.Println(err)
        }

        values := []interface{}{"admin"}
        columns := []string{"type"}

        err = repositories.UpdateUser(tx , user ,values, columns...)
        if err != nil{
            log.Println(err)
        }
        // Commit the transaction if using one
        err = tx.Commit()
        if err != nil {
            log.Printf("Error committing transaction: %v", err)
        }

        bot_settings := models.Bot{
            Bot_token : user_token,
            Main_admin_id : int64(user.ID),
        }

        err = repositories.InsertBotSettings(db , bot_settings)

        if err != nil{
            log.Println(err)
        }
            text := "ربات شما با موفقیت فعال شد"
            msg.Text = text
            userStates[chatID] = ""
    }else{
            text := "توکن مطابقت ندارد دوباره تلاش کنید"
            msg.Text = text
    }

    _ , err = b.bot.Send(msg)
    if err != nil {
        log.Println(err)
    }
}

