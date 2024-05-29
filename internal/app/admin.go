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

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	chatId := update.Message.Chat.ID




	switch userStates[update.Message.Chat.ID] {

		
		case WaitingForEnteringCategoryName:
			category_name := update.Message.Text
			redisKey := fmt.Sprintf("category_name:%s", category_name)
			redisExists, err := b.rdb.Exists(context.Background(), redisKey).Result()
			if err != nil {
				log.Printf("Error checking category status in Redis: %v", err)
				return
			}

			if redisExists == 1 {
				msg.Text = "  این دسته بندی قبلا ایجاد شده است یک نام دیگر وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				if _,err := b.bot.Send(msg);err != nil{
					log.Panic(err)
				}
				return
			}
			
			res := b.rdb.Set(context.Background(), redisKey, category_name,0)
			if err := res.Err();err != nil {
			    log.Println(err)
			}
			category :=  models.Category{
				Name    : category_name,
			}

			err = repositories.InsertCategory(db , category)
			if err != nil {
				log.Println(err)  
			}

			msg.Text = "دسته بندی با نام" + category_name + "ایجاد شد "
			var AdminMainMessageKeyboard = keyboards.AdminMainMessageKeyboard
			msg.ReplyMarkup = AdminMainMessageKeyboard

			if _, err := b.bot.Send(msg); err != nil {
				log.Println(err)
			}
			userStates[update.Message.Chat.ID] = ""

		case WaitingForCategoryUpdate:
			category_name := update.Message.Text

			redisKey := fmt.Sprintf("category_name:%s", category_name)
			redisExists, err := b.rdb.Exists(context.Background(), redisKey).Result()
			if err != nil {
				log.Printf("Error checking category status in Redis: %v", err)
				return
			}

			if redisExists == 1 {
				msg.Text = "  این دسته بندی قبلا ایجاد شده است یک نام دیگر وارد کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				if _,err := b.bot.Send(msg);err != nil{
					log.Panic(err)
				}
				return
			}


			redisGetKey := fmt.Sprintf("user:%d:categoryID", chatId)

			categoryId ,err := b.rdb.Get(context.Background(), redisGetKey).Int()

			if err != nil {
				log.Println(err)
			}

			category, err := repositories.GetCategoryByID(db,categoryId)
			oldCategoryName := category.Name

			if err != nil {
				log.Println(err)
			}

			if err := b.rdb.Del(context.Background(), oldCategoryName).Err(); err != nil {
				log.Println("Error deleting old key:", err)
				return
			}


			UpdateKey := fmt.Sprintf("category_name:%s", category_name)

			res := b.rdb.Set(context.Background(), UpdateKey, category_name,0)
			if err := res.Err();err != nil {
			    log.Println(err)
			}

			values := []interface{}{category_name, time.Now()}
			columns := []string{"name", "updated_at"}

			err = repositories.UpdateCategory(db, &category, values, columns...) 

			if err != nil {
				log.Println(err)
			}

			msg.Text = fmt.Sprintf("عنوان دسته بندی %s به %s تغییر یافت" , oldCategoryName , category_name )
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				if _,err := b.bot.Send(msg);err != nil{
					log.Panic(err)
			}
		userStates[chatId] = ""		
	}

	switch update.Message.Text{
	case "/start":
		text := `سلام به ربات خوش آمدید 👋
		امیدوارم بتونم کمکتون کنم`

		msg.Text = text
		var AdminMainMessageKeyboard = keyboards.AdminMainMessageKeyboard
		msg.ReplyMarkup = AdminMainMessageKeyboard
		SentMessage, err := b.bot.Send(msg)
		if err != nil {
			log.Panic(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", update.Message.Chat.ID)
		res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to orders set: %w", err)
		}

	case "ایجاد دسته بندی":
		msg.Text = "لطفا نام دسته بندی را وارد کنید"
		msg.ReplyMarkup = keyboards.InProgressKeyboard
		if _,err := b.bot.Send(msg);err != nil{
			log.Panic(err)
		}
		userStates[update.Message.Chat.ID] = WaitingForEnteringCategoryName

	case "لیست دسته بندی ها":
		
		CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(db)
		if err != nil {
			log.Panic(err)
		}
		msg.ReplyMarkup = CategoriesButtons
		msg.Text = "لیست دسته بندی ها"
	
		SentMessage, err := b.bot.Send(msg)
		if err != nil{
		    log.Println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", update.Message.Chat.ID)
		
		res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to orders set: %w", err)
		}
		userStates[update.Message.Chat.ID] = CategoryList
	
	case "ویرایش دسته بندی":

		CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(db)
		if err != nil {
			log.Panic(err)
		}
		msg.ReplyMarkup = CategoriesButtons
		msg.Text = "لیست دسته بندی ها"
	
		SentMessage, err := b.bot.Send(msg)
		if err != nil{
		    log.Println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", update.Message.Chat.ID)
		
		res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to orders set: %w", err)
		}
		userStates[update.Message.Chat.ID] = UpdateCategory
	
	case "حذف دسته بندی":

		CategoriesButtons , err:= keyboards.GetAllCategoriesInButton(db)
		if err != nil {
			log.Panic(err)
		}
		msg.ReplyMarkup = CategoriesButtons
		msg.Text = "لیست دسته بندی ها"
	
		SentMessage, err := b.bot.Send(msg)
		if err != nil{
		    log.Println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", update.Message.Chat.ID)
		
		res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to orders set: %w", err)
		}
		userStates[update.Message.Chat.ID] = DeleteCategory
	}

	
}

func HandleAdminUpdateCallBack(update tgbotapi.Update, b *Bot , db *sql.DB){

	chatId := update.CallbackQuery.Message.Chat.ID

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == CategoryList {

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))
		
		category, err := repositories.GetCategoryByID(db, categoryID)

		if err != nil {
			log.Println(err)
		} 

		if !category.Is_active && category.Inactive_message.String != "" {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, category.Inactive_message.String)
			if _, err := b.bot.Send(msg); err != nil {
				panic(err)
			}
		}else{
			
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "")
			messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
			DeleteMessage(b, update.CallbackQuery.Message.Chat.ID, messageID)

			users , err := repositories.GetCategoryMembers(db , categoryID)
			if err !=nil {
				log.Println(err)
			}

			if len(users) <= 0 {
				msg.Text = "این دسته بندی فاقد عضو میباشد"
			}else{
				var text string

				for _, user := range users {
					text += fmt.Sprintf("🧑‍💼 chat_id : %s  name : %s", user.ChatID, user.Username) + "\n"
				}
	
				msg.Text = text
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				
				
			}
			if _, err := b.bot.Send(msg); err != nil {
				log.Println(err)
			}
			
		}
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == UpdateCategory{

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, update.CallbackQuery.Message.Chat.ID, messageID)


		msg := tgbotapi.NewMessage(chatId, "نام جدید این دسته بندی را وارد کنید")
			if _, err := b.bot.Send(msg); err != nil {
				panic(err)
			}

		
		categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatId)
		res := b.rdb.Set(context.Background(), categoryIdKey, categoryID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
		
		if err := b.rdb.SAdd(context.Background(), "categoryIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to categories set: %w", err)
		}

		userStates[chatId] = WaitingForCategoryUpdate
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") && userStates[chatId] == DeleteCategory{



		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))
		category, err := repositories.GetCategoryByID(db, categoryID)
		if err != nil {
			log.Println(err)
		}
		confermationButtons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("بله", fmt.Sprintf("deleteConfermation:%d", categoryID)),
			tgbotapi.NewInlineKeyboardButtonData("انصراف ", fmt.Sprintf("deleteCancelation:%d", categoryID)),
		),)
		
			
		messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
		messageIDstr, _ := b.rdb.Get(context.Background(),messageIdKey).Result()
		messageID, _ := strconv.Atoi(messageIDstr)

		updateConfig := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      update.CallbackQuery.Message.Chat.ID,          
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

	if strings.HasPrefix(update.CallbackQuery.Data, "deleteConfermation:"){

		messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, update.CallbackQuery.Message.Chat.ID, messageID)

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "deleteConfermation:"))
		category, err := repositories.GetCategoryByID(db, categoryID)
		if err != nil {
			log.Println(err)
		}


		err = repositories.DeleteCategory(db, categoryID)
		if err != nil {
			log.Println(err)
		}
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("دسته بندی %s با موفقیت حذف شد", category.Name))
		if _, err := b.bot.Send(msg); err != nil {
			panic(err)
		}
		userStates[chatId] = ""
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "deleteCancelation:"){
		messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()	
		DeleteMessage(b, update.CallbackQuery.Message.Chat.ID, messageID)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "انصراف از حذف")
		if _, err := b.bot.Send(msg); err != nil {
			panic(err)
		}
		userStates[chatId] = ""
	}

	


}