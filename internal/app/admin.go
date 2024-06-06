package app

import(
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/Salehisaac/Supply-and-Demand.git/static/keyboards"
	"database/sql"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/database/repositories"
	"log"
	"fmt"
	"context"
	"strings"
	"strconv"
	"time"
)



func HandleAdminUpdateMessage(update tgbotapi.Update, b *Bot , db *sql.DB){

	tx, err := db.Begin()
    if err != nil {
        log.Println("Error beginning transaction:", err)
        return
    }
    defer func() {
        if err := recover(); err != nil {
            log.Println("Panic occurred, rolling back transaction:", err)
            if err := tx.Rollback(); err != nil {
                log.Println("Error rolling back transaction:", err)
            }
        }
    }()


	chatId := update.Message.Chat.ID
	msg := tgbotapi.NewMessage(chatId, "")
	msg.ReplyMarkup , _ = keyboards.GetAdminMainMessageKeyboard(db)
	var addToredis bool 

	switch update.Message.Text{
		case "/start":
			msg.Text = WelcomeMessage
			userStates[chatId] = ""
		case "◀️ بازگشت به منوی اصلی" :
			msg.Text = WelcomeMessage
			userStates[chatId] = ""
	}


	switch userStates[chatId] {

		
		case WaitingForEnteringCategoryName:
			category_name := update.Message.Text
			categoryKey := fmt.Sprintf("category_name:%s", category_name)
			categoryExists, err := b.rdb.Exists(context.Background(), categoryKey).Result()
			if err != nil {
				log.Printf("Error checking category status in Redis: %v", err)
				return
			}

			if categoryExists == 1 {
				msg.Text = "  این دسته بندی قبلا ایجاد شده است یک نام دیگر وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				if _,err := b.bot.Send(msg);err != nil{
					log.Panic(err)
				}
				return
			}
			
			res := b.rdb.Set(context.Background(), categoryKey, category_name,0)
			if err := res.Err();err != nil {
			    log.Println(err)
			}
			category :=  models.Category{
				Name    : category_name,
				Inactive_message: sql.NullString{String: "عدم وجود پاسخگوی فعال", Valid: true},
			}

			err = repositories.InsertCategory(tx,db , category)
			if err != nil {
				log.Println(err)  
			}

			msg.Text = fmt.Sprintf("دسته بندی با نام %s یا موفقیت ایجاد شد" , category_name) 
			userStates[chatId] = ""

		case WaitingForCategoryUpdate:
			category_name := update.Message.Text
			categoryKey := fmt.Sprintf("category_name:%s", category_name)
			categoryExists, err := b.rdb.Exists(context.Background(), categoryKey).Result()
			if err != nil {
				log.Printf("Error checking category status in Redis: %v", err)
				return
			}

			if categoryExists == 1 {
				msg.Text = "  این دسته بندی قبلا ایجاد شده است یک نام دیگر وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}

			redisGetKey := fmt.Sprintf("user:%d:categoryID", chatId)

			categoryId ,err := b.rdb.Get(context.Background(), redisGetKey).Int()

			if err != nil {
				log.Println(err)
			}

			category, err := repositories.GetCategoryByID(tx , db,categoryId)
			oldCategoryName := category.Name

			if err != nil {
				log.Println(err)
			}

			categoryKey = fmt.Sprintf("category_name:%s", oldCategoryName)

			if err := b.rdb.Del(context.Background(), categoryKey).Err(); err != nil {
				log.Println("Error deleting old key:", err)
				return
			}


			categoryUpdateKey := fmt.Sprintf("category_name:%s", category_name)

			res := b.rdb.Set(context.Background(), categoryUpdateKey, category_name,0)
			if err := res.Err();err != nil {
			    log.Println(err)
			}

			values := []interface{}{category_name, time.Now()}
			columns := []string{"name", "updated_at"}

			err = repositories.UpdateCategory(tx , db, category, values, columns...) 

			if err != nil {
				log.Println(err)
			}

			msg.Text = fmt.Sprintf("عنوان دسته بندی %s به %s تغییر یافت" , oldCategoryName , category_name )

			userStates[chatId] = ""		
	
		case WaitingForUserNameToAdd :
			username := update.Message.Text
			user, err := repositories.GetUserByUsername(tx , db, username)

			if err != nil {
				msg.Text = "کاربری با این نام کاربری وجود ندارد نام دیگری وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}

			categoryGetKey := fmt.Sprintf("user:%d:categoryID", chatId)

			categoryId ,err := b.rdb.Get(context.Background(), categoryGetKey).Int()

			if err != nil {
				log.Println(err)
			}

			category, err := repositories.GetCategoryByID(tx ,db ,categoryId)
			if err != nil {
				log.Println(err)
			}

			if exists, err := repositories.UserExistsInCategoryCheck(tx ,db, user.ID, category.ID); err != nil {
				log.Printf("Error checking if user exists in category: %v", err)
			} else if exists {
				msg.Text = " کاربر از قبل در این دسته بندی وجود دارد نام دیگری وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}
			err = repositories.AddingMemberToCategory(tx ,db,user.ID, category.ID)

			if err != nil {
				log.Println(err)
			}

			if user.Type != "responder" && user.Type != "admin" {
				values := []interface{}{"responder"}
				columns := []string{"type"}

				err := repositories.UpdateUser(tx, user, values, columns...)
				if err != nil{
					log.Println(err)
				}
			}

			if !category.Contain_active_responser{
				values := []interface{}{1 , 1}
				columns := []string{"contain_active_responser", "is_active"}

				err := repositories.UpdateCategory(tx ,db, category, values, columns...)
				if err != nil{
					log.Println(err)
				}
			}

			msg.Text = fmt.Sprintf("کاربر %s به دسته بندی %s اضافه شد" , user.Username , category.Name )
			userStates[chatId] = ""	

		case WaitingForUserNameToRemove:
			username := update.Message.Text
			user, err := repositories.GetUserByUsername(tx ,db, username)

			if err != nil {
				msg.Text = "کاربری با این نام کاربری وجود ندارد نام دیگری وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}

			categoryGetKey := fmt.Sprintf("user:%d:categoryID", chatId)

			categoryId ,err := b.rdb.Get(context.Background(), categoryGetKey).Int()

			if err != nil {
				log.Println(err)
			}

			category, err := repositories.GetCategoryByID(tx ,db,categoryId)
			if err != nil {
				log.Println(err)
			}

			if exists, err := repositories.UserExistsInCategoryCheck(tx ,db, user.ID, category.ID); err != nil {
				log.Printf("Error checking if user exists in category: %v", err)
			} else if !exists {
				msg.Text = " کاربر در این دسته بندی نیست"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}
			err = repositories.RemovingMemberFromCategory(tx ,db,user.ID, category.ID)

			if err != nil {
				log.Println(err)
			}

			msg.Text = fmt.Sprintf("کاربر %s از دسته بندی %s حذف شد" , user.Username , category.Name )

			categoryMembers , err := repositories.GetCategoryMembers(tx ,db, category.ID)
			if err != nil {
				log.Println(err)
			}
			userCategories , err := repositories.GetUserCategories(tx ,db, user.ID)

			if err != nil {
				log.Println(err)
			}

			if len(userCategories) == 0 && user.Type != "admin"{

				values := []interface{}{"customer"}
				columns := []string{"type"}

				err := repositories.UpdateUser(tx, user, values, columns...)
				if err != nil{
					log.Println(err)
				}
			}
			if len(categoryMembers) == 0{
				values := []interface{}{0 , 0, "عدم وجود پاسخگوی فعال"}
				columns := []string{"contain_active_responser", "is_active" , "inactive_message"}

				err := repositories.UpdateCategory(tx ,db, category, values, columns...)
				if err != nil{
					log.Println(err)
				}
			}

			userStates[chatId] = ""	
		
		case WaitingForBotInactiveMessage :
			Inactive_message := update.Message.Text
			InactiveMessageSetKey := fmt.Sprintf("user:%d:InactiveMessage", chatId)

			res := b.rdb.Set(context.Background(), InactiveMessageSetKey, Inactive_message,0)
			if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
			}

			confermationButtons := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("بله", "botInactiveMessageConfirmation:"),
					tgbotapi.NewInlineKeyboardButtonData("انصراف ","deleteCancelation:"),
				),)
			msg.Text = fmt.Sprintf("آیا این پیام را تایید میکنید ؟ \n %s", Inactive_message)
			msg.ReplyMarkup = confermationButtons
			addToredis = true

			userStates[chatId] = ""	

		case WaitingForCategoryInactiveMessage :

			Inactive_message := update.Message.Text
			InactiveMessageSetKey := fmt.Sprintf("user:%d:InactiveMessage", chatId)

			res := b.rdb.Set(context.Background(), InactiveMessageSetKey, Inactive_message,0)
			if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
			}

			confermationButtons := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("بله", "categoryInactiveMessageConfirmation:"),
					tgbotapi.NewInlineKeyboardButtonData("انصراف ","deleteCancelation:"),
				),)
			msg.Text = fmt.Sprintf("آیا این پیام را تایید میکنید ؟ \n %s", Inactive_message)
			msg.ReplyMarkup = confermationButtons
			addToredis = true
			userStates[chatId] = ""	
	
		case WaitingForCategoryMemberUsernameToDeactive:
			username := update.Message.Text
			user, err := repositories.GetUserByUsername(tx ,db, username)

			if err != nil {
				msg.Text = "کاربری با این نام کاربری وجود ندارد نام دیگری وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}

			redisGetKey := fmt.Sprintf("user:%d:categoryID", chatId)

			categoryId ,err := b.rdb.Get(context.Background(), redisGetKey).Int()

			if err != nil {
				log.Println(err)
			}

			category, err := repositories.GetCategoryByID(tx ,db,categoryId)
			if err != nil {
				log.Println(err)
			}

			if exists, err := repositories.UserExistsInCategoryCheck(tx ,db, user.ID, category.ID); err != nil {
				log.Printf("Error checking if user exists in category: %v", err)
			} else if !exists {
				msg.Text = " کاربر در این دسته بندی نیست"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}
			err = repositories.DeactivingMember(tx ,db,user.ID)
			if err != nil {
				log.Panic(err)
			}



			msg.Text = fmt.Sprintf("کاربر %s غیرفعال شد" , user.Username)

			ActivecategoryMembers , err := repositories.GetActiveCategoryMembers(tx,db, category.ID)
			if err != nil {
				log.Println(err)
			}
			if len(ActivecategoryMembers) == 0{
				values := []interface{}{0 , 0, "عدم وجود پاسخگوی فعال"}
				columns := []string{"contain_active_responser", "is_active" , "inactive_message"}

				err := repositories.UpdateCategory(tx ,db, category, values, columns...)
				if err != nil{
					log.Println(err)
				}
			}

			userStates[chatId] = ""	
		
		case WaitingForCategoryMemberUsernameToReactive:
			username := update.Message.Text
			user, err := repositories.GetUserByUsername(tx ,db, username)

			if err != nil {
				msg.Text = "کاربری با این نام کاربری وجود ندارد نام دیگری وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}

			redisGetKey := fmt.Sprintf("user:%d:categoryID", chatId)

			categoryId ,err := b.rdb.Get(context.Background(), redisGetKey).Int()

			if err != nil {
				log.Println(err)
			}

			category, err := repositories.GetCategoryByID(tx ,db,categoryId)
			if err != nil {
				log.Println(err)
			}

			if exists, err := repositories.UserExistsInCategoryCheck(tx ,db, user.ID, category.ID); err != nil {
				log.Printf("Error checking if user exists in category: %v", err)
			} else if !exists {
				msg.Text = " کاربر در این دسته بندی نیست"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}
			err = repositories.ReactivingMember(tx ,db,user.ID)

			if err != nil {
				log.Println(err)
			}

			msg.Text = fmt.Sprintf("کاربر %s فعال شد" , user.Username)

			if !category.Contain_active_responser{
				values := []interface{}{1 , 1}
				columns := []string{"contain_active_responser", "is_active"}

				err := repositories.UpdateCategory(tx ,db, category, values, columns...)
				if err != nil{
					log.Println(err)
				}
			}

			userStates[chatId] = ""	

	}

	messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
	messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
	DeleteMessage(b, chatId, messageID)

	switch update.Message.Text{

		case "ایجاد دسته بندی":

			msg.Text = "لطفا نام دسته بندی را وارد کنید"
			msg.ReplyMarkup = keyboards.InProgressKeyboard
			userStates[chatId] = WaitingForEnteringCategoryName

		case "لیست دسته بندی ها":
			
			CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(tx ,db)
			if err != nil {
				log.Panic(err)
			}
			if len(CategoriesButtons.InlineKeyboard) ==0 {
				msg.Text = "در حال حاضر دسته بندی فعالی وجود ندارد"
				break
			}
			msg.ReplyMarkup = CategoriesButtons
			msg.Text = "لیست دسته بندی ها"
			addToredis = true
			userStates[chatId] = CategoryList
		
		case "ویرایش دسته بندی":

			CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(tx ,db)
			if err != nil {
				log.Panic(err)
			}
			if len(CategoriesButtons.InlineKeyboard) ==0 {
				msg.Text = "در حال حاضر دسته بندی فعالی وجود ندارد"
				break
			}
			msg.ReplyMarkup = CategoriesButtons
			msg.Text = "لیست دسته بندی ها"
			addToredis = true
			userStates[chatId] = UpdateCategory
		
		case "حذف دسته بندی":

			CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(tx ,db)
			if err != nil {
				log.Panic(err)
			}
			if len(CategoriesButtons.InlineKeyboard) ==0 {
				msg.Text = "در حال حاضر دسته بندی فعالی وجود ندارد"
				break
			}
			msg.ReplyMarkup = CategoriesButtons
			msg.Text = "لیست دسته بندی ها"
			addToredis = true
			userStates[chatId] = DeleteCategory
		
		case "افزودن عضو به دسته بندی":

			CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(tx ,db)
			if err != nil {
				log.Panic(err)
			}
			if len(CategoriesButtons.InlineKeyboard) ==0 {
				msg.Text = "در حال حاضر دسته بندی فعالی وجود ندارد"
				break
			}
			msg.ReplyMarkup = CategoriesButtons
			msg.Text = "لیست دسته بندی ها"
			addToredis = true
			userStates[chatId] = AddingMemberToCategory
		
		case "حذف عضو از دسته بندی":
			
			CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(tx ,db)
			if err != nil {
				log.Panic(err)
			}
			if len(CategoriesButtons.InlineKeyboard) ==0 {
				msg.Text = "در حال حاضر دسته بندی فعالی وجود ندارد"
				break
			}
			msg.ReplyMarkup = CategoriesButtons
			msg.Text = "لیست دسته بندی ها"
			addToredis = true
			userStates[chatId] = RemovingMemberToCategory
		
		case "دریافت آمار":

			var text string

			categories , err := repositories.GetAllCategories(tx ,db)
			if err != nil {
				log.Panic(err)
			}

			var allCategoriesCount int
			var allMembersCount int
			var allRequestsCount int
			var allAnswersCount int


			allCategoriesCount = len(categories)

			for _ , category := range categories {
				
				members , err := repositories.GetCategoryMembers(tx ,db, category.ID)
				if err != nil {
					log.Panic(err)
				}
				membersCount := len(members)
				allMembersCount += membersCount

				requests , err := repositories.GetCategoryRequests(tx ,db , category.ID)

				if err != nil {
					println(err)
					return
				}
				requestsCount := len(requests)
				allRequestsCount += requestsCount

				var answersCount int

				for _ , request := range requests{
					answers , err := repositories.GetRequestAnswers(tx , request.ID)
					if err != nil {
						println(err)
						return
					}
					answersCount += len(answers)
				}
				allAnswersCount += answersCount

				text += fmt.Sprintf("%s\nتعداد کاربران : %d\nتعداد درخواست ها: %d\nتعداد پاسخ ها : %d\n\n ----------------------- \n", category.Name, membersCount, requestsCount, answersCount)

			}

			if text == "" {
				msg.Text = "فعالیتی انجام نشده است"
				break
			}

			msg.Text = text

			if _,err := b.bot.Send(msg); err != nil {
				println(err)
			}

			msg.Text = fmt.Sprintf("آمار کل: \nتعداد دسته‌ ها: %d\nتعداد اعضا: %d\nتعداد درخواست‌ ها:  %d\nتعداد پاسخ‌ها: %d",
			allCategoriesCount, allMembersCount, allRequestsCount, allAnswersCount)
		
		case "غیر فعال کردن موقت بات":

			bot_settings , err := repositories.GetTheBotConfigs(db)
			

			if err != nil {
				log.Println(err)
			}

			if !bot_settings.Is_active {
				msg.Text = "ربات غیر فعال است!"
				return
			}else{
				msg.Text = "پیام خود را وارد کنید (این پیام در هنگامی که ربات غیر فعال است به کاربران نمایش داده خواهد شد)"
				msg.ReplyMarkup = keyboards.InProgressKeyboard		
		
				userStates[chatId] = WaitingForBotInactiveMessage
			}

		case "فعال کردن بات":

			bot_settings , err := repositories.GetTheBotConfigs(db)
			

			if err != nil {
				log.Println(err)
			}
			values := []interface{}{1}
			columns := []string{"is_active"}
			err = repositories.UpdateBotSettings(db, bot_settings , values , columns...)
			if err != nil{
				println(err)
			}
			msg.Text = "ربات با موفقیت فعال شد"
			msg.ReplyMarkup, _ = keyboards.GetAdminMainMessageKeyboard(db)
		
		case "غیرفعال کردن موقت یک دسته بندی":

			CategoriesButtons , err:= keyboards.GetActiveCategoriesInButton(tx ,db)
			if err != nil {
				log.Panic(err)
			}
			if len(CategoriesButtons.InlineKeyboard) > 0 {
				msg.ReplyMarkup = CategoriesButtons
				msg.Text = "لیست دسته بندی های فعال"
				addToredis = true
				userStates[chatId] = DeactiveCategory
			} else {
				msg.Text = "هیچ دسته بندی فعالی وجود ندارد"	
			}	
		
		case "فعال کردن دسته بندی":

			CategoriesButtons , err:= keyboards.GetInactiveCategoriesInButton(tx ,db)
			if err != nil {
				log.Println(err)
			}
			
			if len(CategoriesButtons.InlineKeyboard) == 0 {
				msg.Text = "هیچ دسته بندی غیر فعالی وجود ندارد"
				log.Println("CategoriesButtons is nil")
			} else {
				msg.Text = "لیست دسته بندی های غیرفعال"
				msg.ReplyMarkup = CategoriesButtons
				addToredis = true
			}
			
			userStates[chatId] = ReactiveCategory	

		case "غیرفعال کردن موقت پاسخگو":

			CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(tx ,db)
			if err != nil {
				log.Panic(err)
			}
			if len(CategoriesButtons.InlineKeyboard) ==0 {
				msg.Text = "در حال حاضر دسته بندی فعالی وجود ندارد"
				break
			}
			msg.ReplyMarkup = CategoriesButtons
			msg.Text = "لیست دسته بندی ها"
			addToredis = true
			userStates[chatId] = DeactiveCategoryMember	

		case "فعال کردن پاسخگو":

		CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(tx ,db)
		if err != nil {
			log.Panic(err)
		}
		if len(CategoriesButtons.InlineKeyboard) ==0 {
			msg.Text = "در حال حاضر دسته بندی فعالی وجود ندارد"
			break
		}
		msg.ReplyMarkup = CategoriesButtons
		msg.Text = "لیست دسته بندی ها"
		addToredis = true
		userStates[chatId] = ReactiveCategoryMember	
		
	}

	if addToredis{
		SentMessage, err := b.bot.Send(msg)
		if err != nil{
		    log.Println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		
		res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to orders set: %w", err)
		}
	}else{
		_, err := b.bot.Send(msg)
		if err != nil{
		    log.Println(err)
		}
	}
	// Commit the transaction if using one
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
	}



	
}

func HandleAdminUpdateCallBack(update tgbotapi.Update, b *Bot , db *sql.DB){

	tx, err := db.Begin()
    if err != nil {
        log.Println("Error beginning transaction:", err)
        return
    }
    defer func() {
        if err := recover(); err != nil {
            log.Println("Panic occurred, rolling back transaction:", err)
            if err := tx.Rollback(); err != nil {
                log.Println("Error rolling back transaction:", err)
            }
        }
    }()

	chatId := update.CallbackQuery.Message.Chat.ID
	msg := tgbotapi.NewMessage(chatId, "")
	msg.ReplyMarkup , _ = keyboards.GetAdminMainMessageKeyboard(db)
	var addToredis bool 

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == CategoryList {

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))
		
		category, err := repositories.GetCategoryByID(tx ,db, categoryID)

		if err != nil {
			log.Println(err)
		} 

		if !category.Is_active && category.Inactive_message.String != "" {
			msg.Text = category.Inactive_message.String
			if _, err := b.bot.Send(msg); err != nil {
				panic(err)
			}
		}else{
			
			messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
			DeleteMessage(b, chatId, messageID)

			users , err := repositories.GetCategoryMembers(tx ,db , categoryID)
			if err !=nil {
				log.Println(err)
			}

			if len(users) <= 0 {
				msg.Text = "این دسته بندی فاقد عضو میباشد"
				if err != nil {
					log.Println(err)
				}

			}else{
				var text string

				for _, user := range users {
					text += fmt.Sprintf("🧑‍💼 chat_id : %s  name : %s", user.ChatID, user.Username) + "\n"
				}
	
				msg.Text = text
				msg.ReplyMarkup = keyboards.InProgressKeyboard
			}	
		}
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == UpdateCategory{

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, chatId, messageID)


		msg.Text = "نام جدید این دسته بندی را وارد کنید"

		
		categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatId)
		res := b.rdb.Set(context.Background(), categoryIdKey, categoryID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
		
		if err := b.rdb.SAdd(context.Background(), "categoryIDs", categoryIdKey).Err();err !=nil{	
			log.Fatal("failed to add to categories set: %w", err)
		}

		userStates[chatId] = WaitingForCategoryUpdate
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == DeleteCategory{



		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))
		category, err := repositories.GetCategoryByID(tx ,db, categoryID)
		if err != nil {
			log.Println(err)
		}
		confermationButtons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("بله", fmt.Sprintf("CategorydeleteConfirmation:%d", categoryID)),
			tgbotapi.NewInlineKeyboardButtonData("انصراف ", fmt.Sprintf("deleteCancelation:%d", categoryID)),
		),)
		
			
		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageIDstr, _ := b.rdb.Get(context.Background(),messageIdKey).Result()
		messageID, _ := strconv.Atoi(messageIDstr)

		updateConfig := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      chatId,          
				MessageID:   messageID,      
				InlineMessageID: "",         
			},
			Text:fmt.Sprintf("آیا از حذف دسته بندی %s اطمینان دارید ؟", category.Name), 
		}
		updateConfig.ReplyMarkup = &confermationButtons

		SentMessage, err := b.bot.Send(updateConfig);
		if err != nil {
			panic(err)
		}
			
		res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
		
		if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to orders set: %w", err)
		}
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == AddingMemberToCategory{

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, chatId, messageID)

		msg.Text = "نام کاربری را وارد کنید"
		
		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatId)
		res := b.rdb.Set(context.Background(), categoryIdKey, categoryID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
		
		if err := b.rdb.SAdd(context.Background(), "categoryIDs", categoryID).Err();err !=nil{	
			log.Fatal("failed to add to categories set: %w", err)
		}

		userStates[chatId] = WaitingForUserNameToAdd
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == RemovingMemberToCategory{

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, chatId, messageID)

		msg.Text = "نام کاربری را وارد کنید"
		
		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatId)
		res := b.rdb.Set(context.Background(), categoryIdKey, categoryID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
		
		if err := b.rdb.SAdd(context.Background(), "categoryIDs", categoryID).Err();err !=nil{	
			log.Fatal("failed to add to categories set: %w", err)
		}

		userStates[chatId] = WaitingForUserNameToRemove
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == DeactiveCategory{
		
		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, chatId, messageID)

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatId)
		res := b.rdb.Set(context.Background(), categoryIdKey, categoryID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
		
		if err := b.rdb.SAdd(context.Background(), "categoryIDs", categoryID).Err();err !=nil{	
			log.Fatal("failed to add to categories set: %w", err)
		}

		
		msg.Text = "پیام خود را وارد کنید (این پیام در هنگامی که دسته بندی غیر فعال است به کاربران نمایش داده خواهد شد)"
		msg.ReplyMarkup = keyboards.InProgressKeyboard
		addToredis = true

		userStates[chatId] = WaitingForCategoryInactiveMessage
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == ReactiveCategory{

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()
		DeleteMessage(b,chatId, messageID)	
		
		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))
		category, err := repositories.GetCategoryByID(tx ,db,categoryID)
		if err != nil {
			log.Println(err)
		}
        

        if err != nil {
            log.Println(err)
        }
		values := []interface{}{1}
		columns := []string{"is_active"}
		err = repositories.UpdateCategory(tx ,db, category , values , columns...)
		if err != nil{
			println(err)
		}

		msg.Text = "دسته بندی شما فعال شد"
     
		userStates[chatId] = ""
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == DeactiveCategoryMember{

		categoryID, err := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatId)
		res := b.rdb.Set(context.Background(), categoryIdKey, categoryID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
		
		if err := b.rdb.SAdd(context.Background(), "categoryIDs", categoryID).Err();err !=nil{	
			log.Fatal("failed to add to categories set: %w", err)
		}

		if err != nil {
			log.Println(err)
		}

		var text string

		users, err := repositories.GetActiveCategoryMembers(tx,db, categoryID)

		if err != nil {
			log.Println(err)
		}

		if len(users) == 0 {

			messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()
			DeleteMessage(b,chatId, messageID)	
			msg.Text = "کاربر فعالی برای این دسته بندی وجود ندارد"
		}else{
			for _ , user := range users{
				text += fmt.Sprintf("%s\n", user.Username)
			}
			messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
	
			updateConfig := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:      chatId,          
					MessageID:   messageID,      
					InlineMessageID: "",         
				},
				Text:text,
			}
	
			if _, err := b.bot.Send(updateConfig); err != nil {
				panic(err)
			}
	
			msg.Text = "نام کاربری کاربر مورد نظر خود را وارد کنید"
			userStates[chatId] = WaitingForCategoryMemberUsernameToDeactive
		}
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == ReactiveCategoryMember{

		categoryID, err := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		if err != nil {
			log.Println(err)
		}
		

		var text string

		users, err := repositories.GetInactiveCategoryMembers(tx ,db, categoryID)

		if err != nil {
			log.Println(err)
		}

		if len(users) == 0 {

			messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()
			DeleteMessage(b,chatId, messageID)	
			msg.Text = "کاربر غیرفعالی برای این دسته بندی وجود ندارد"
		}else{
			for _ , user := range users{
				text += fmt.Sprintf("%s\n", user.Username)
			}
			messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
	
			updateConfig := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:      chatId,          
					MessageID:   messageID,      
					InlineMessageID: "",         
				},
				Text:text,
			}
	
			if _, err := b.bot.Send(updateConfig); err != nil {
				panic(err)
			}

			categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatId)
			res := b.rdb.Set(context.Background(), categoryIdKey, categoryID,0)
			if err := res.Err();err != nil {
				log.Fatal("failed to set: %w", err)
			}
			
			if err := b.rdb.SAdd(context.Background(), "categoryIDs", categoryID).Err();err !=nil{	
				log.Fatal("failed to add to categories set: %w", err)
			}

			if err != nil {
				log.Println(err)
		}
	
			msg.Text = "نام کاربری کاربر مورد نظر خود را وارد کنید"
			userStates[chatId] = WaitingForCategoryMemberUsernameToReactive
		}
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "CategorydeleteConfirmation:"){

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, chatId, messageID)

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "CategorydeleteConfirmation:"))
		category, err := repositories.GetCategoryByID(tx ,db, categoryID)
		if err != nil {
			log.Println(err)
		}

		categoryMembers , err := repositories.GetCategoryMembers(tx ,db, category.ID)
			if err != nil {
				log.Println(err)
			}	
		
		err = repositories.DeleteCategory(tx ,db, categoryID)
		if err != nil {
			log.Println(err)
		}
		msg.Text = fmt.Sprintf("دسته بندی %s با موفقیت حذف شد", category.Name)

		for _, member := range categoryMembers{
			
			userCategories , err := repositories.GetUserCategories(tx ,db, member.ID)

			if err != nil {
				log.Println(err)
			}

			if len(userCategories) == 0  && member.Type != "admin"{

				values := []interface{}{"customer"}
				columns := []string{"type"}

				err := repositories.UpdateUser(tx , member, values, columns...)
				if err != nil{
					log.Println(err)
				}
			}
		}
	
		redisKey := fmt.Sprintf("category_name:%s", category.Name)

		if err := b.rdb.Del(context.Background(), redisKey).Err(); err != nil {
			log.Println("Error deleting old key:", err)
			return
		}

		userStates[chatId] = ""
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "deleteCancelation:"){

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, chatId, messageID)

		msg.Text = "لغو درخواست"
		userStates[chatId] = ""
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "botInactiveMessageConfirmation:"){

		redisSetKey := fmt.Sprintf("user:%d:InactiveMessage", chatId)
		In_active_message, _ := b.rdb.Get(context.Background(),redisSetKey).Result()	

		bot_settings , err := repositories.GetTheBotConfigs(db)
        

        if err != nil {
            log.Println(err)
        }
		values := []interface{}{0, In_active_message}
		columns := []string{"is_active", "inactive_message"}
		err = repositories.UpdateBotSettings(db, bot_settings , values , columns...)
		if err != nil{
			println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	

		updateConfig := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      chatId,          
				MessageID:   messageID,      
				InlineMessageID: "",         
			},
			Text:fmt.Sprintf("ربات با موفقیت غیر فعال شد"), 
		}

		if _, err := b.bot.Send(updateConfig); err != nil {
			panic(err)
		}

		userStates[chatId] = ""
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "categoryInactiveMessageConfirmation:"){

	

		redisSetKey := fmt.Sprintf("user:%d:InactiveMessage", chatId)
		In_active_message, _ := b.rdb.Get(context.Background(),redisSetKey).Result()	

		redisGetKey := fmt.Sprintf("user:%d:categoryID", chatId)

		categoryId ,err := b.rdb.Get(context.Background(), redisGetKey).Int()

		if err != nil {
			log.Println(err)
		}

		category, err := repositories.GetCategoryByID(tx ,db,categoryId)
		if err != nil {
			log.Println(err)
		}
        

        if err != nil {
            log.Println(err)
        }
		values := []interface{}{0, In_active_message}
		columns := []string{"is_active", "inactive_message"}
		err = repositories.UpdateCategory(tx ,db, category , values , columns...)
		if err != nil{
			println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	

		updateConfig := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      chatId,          
				MessageID:   messageID,      
				InlineMessageID: "",         
			},
			Text:fmt.Sprintf("دسته بندی %s با موفقیت غیر فعال شد", category.Name), 
		}


		if _, err := b.bot.Send(updateConfig); err != nil {
			panic(err)
		}
	}

	if addToredis{
		SentMessage, err := b.bot.Send(msg)
		if err != nil{
		    log.Println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatId)
		
		res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to orders set: %w", err)
		}
	}else{
		_, err := b.bot.Send(msg)
		if err != nil{
		    log.Println(err)
		}
	}
	// Commit the transaction if using one
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
	}

}