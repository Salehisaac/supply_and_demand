package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Salehisaac/Supply-and-Demand.git/internal/database/repositories"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
	"github.com/Salehisaac/Supply-and-Demand.git/static/keyboards"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)



func HandleResponderUpdateMessage(update tgbotapi.Update, b *Bot , db *sql.DB){

	chatID := update.Message.Chat.ID
	msg := tgbotapi.NewMessage(chatID, "")
	var addToredis bool 


	switch update.Message.Text{
		case "/start":
			userStates[chatID] = ""
			msg.Text = WelcomeMessage
			var ResponderMainMessageInlibeKeyboard = keyboards.ResponderMainMessageInlibeKeyboard
			msg.ReplyMarkup = ResponderMainMessageInlibeKeyboard
			addToredis = true
		case "◀️ بازگشت به منوی اصلی" :
			msg.Text = WelcomeMessage
			var ResponderMainMessageInlibeKeyboard = keyboards.ResponderMainMessageInlibeKeyboard
			msg.ReplyMarkup = ResponderMainMessageInlibeKeyboard
			addToredis = true
			userStates[chatID] = ""
		
	}

	switch userStates[chatID]{

		case WaittingForRequestReply:
			answer_text := update.Message.Text
			originalMessageID := update.Message.ReplyToMessage.MessageID
			userKey := fmt.Sprintf("user:%d:request_message_ides", chatID)
			var tracking_code string 
			messageIDs, err := b.rdb.SMembers(context.Background(), userKey).Result()
			if err != nil {
				log.Println(err)
			}

			// Check if the originalMessageID is in the retrieved message IDs
			found := false
			for _, id := range messageIDs {
				parts := strings.Split(id, "_")
				if len(parts) > 0 {
					msgID, err := strconv.Atoi(parts[0])
					if err == nil && msgID == originalMessageID {
						found = true
						tracking_code = parts[1]
						break
					}
				}
			}

			if !found{
				msg.Text = "روی یکی از درخواست ها ریپلای کنید"
				break
			}
			
			// Begin transaction
			tx, err := db.Begin()
			if err != nil {
				log.Println("Failed to begin transaction:", err)
				msg.Text = "خطا در شروع تراکنش"
				break
			}

			// Ensure rollback on failure
			defer func() {
				if err != nil {
					tx.Rollback()
					log.Println("Transaction rolled back due to error:", err)
				}
			}()

			request, err := repositories.GetRequestByTrackingCode(tx,db, tracking_code)
			if err != nil {
				log.Println(err)
				msg.Text = "خطا در بازیابی درخواست"
				break
			}

			if request.RequestID != nil {
				msg.Text = "درخواست عطف قابلیت پاسخگویی ندارد . درخواست دیگیری را امتحان کنید"
				msg.ReplyMarkup = keyboards.InProgressKeyboard
				break
			}

			answer := models.Answer{
				ResponserID:  request.ResponserID,
				CustomerID:   request.CustomerID,
				Text:         answer_text,
				TrackingCode: generateTrackingCode(),
				RequestID:    request.ID,
			}

			err = repositories.InsertAnswer(tx,db, answer)
			if err != nil {
				log.Println(err)
				msg.Text = "خطا در درج پاسخ"
				break
			}

			values := []interface{}{"answerd"}
			columns := []string{"status"}

			err = repositories.UpdateRequest(tx, db ,request, values, columns...)
			if err != nil {
				log.Println(err)
				msg.Text = "خطا در به روز رسانی درخواست"
				break
			}

			user, err := repositories.GetUserByID(tx,db, request.CustomerID)
			if err != nil {
				log.Println(err)
				msg.Text = "خطا در بازیابی اطلاعات کاربر"
				break
			}

			// Commit transaction
			if err := tx.Commit(); err != nil {
				log.Println("Failed to commit transaction:", err)
				msg.Text = "خطا در نهایی سازی تراکنش"
				break
			}

				intChatid, err := strconv.ParseInt(user.ChatID, 10, 64)
				if err != nil {
					log.Println(err)  
				}

				user_respond := tgbotapi.NewMessage(intChatid, "")

				user_respond.Text = fmt.Sprintf("پاسخ ادمین به درخواست با کد رهگیری %s" , tracking_code)

				_, err = b.bot.Send(user_respond)
				if err != nil{
					log.Println(err)
				}

				user_respond.Text = answer_text

				_, err = b.bot.Send(user_respond)
				if err != nil{
					log.Println(err)
				}
					
				msg.Text = "پاسخ شما با موفقیت ثبت شد"

				userStates[chatID] = ""	
				
	}

	if addToredis{
		SentMessage, err := b.bot.Send(msg)
		if err != nil{
		    log.Println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
		
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
}

func HandleResponderUpdateCallBack(update tgbotapi.Update, b *Bot, db *sql.DB) {
	chatID := update.CallbackQuery.Message.Chat.ID
	msg := tgbotapi.NewMessage(chatID, "")
	var addToredis bool

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		msg.Text = "خطا در شروع تراکنش"
		_, _ = b.bot.Send(msg)
		return
	}
	user, _ := repositories.GetUserByChatID(tx , db, chatID)

	// Ensure rollback on failure
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Println("Transaction rolled back due to error:", err)
		}
	}()

	switch update.CallbackQuery.Data {
	case "answer_to_request":
		messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
		messageID, _ := b.rdb.Get(context.Background(), messageIdKey).Int()
		DeleteMessage(b, chatID, messageID)

		responderRequests, err := repositories.GetAllUnseenTheResponderRequests(tx, user.ID)
		if err != nil {
			log.Println(err)
			msg.Text = "خطا در بازیابی درخواست‌ها"
			break
		}

		if len(responderRequests) == 0 {
			msg.Text = "در حال حاضر درخواست پاسخ داده نشده ای برای شما وجود ندارد"
			_, err = b.bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
			tx.Commit() // Commit here since no further DB operations are needed
			return
		}

		userKey := fmt.Sprintf("user:%d:request_message_ides", chatID)
		err = b.rdb.Del(context.Background(), userKey).Err()
		if err != nil {
			log.Fatalf("Failed to delete set: %v", err)
		}

		for _, request := range responderRequests {
			customer, err := repositories.GetUserByID(tx, db,  request.CustomerID)
			if err != nil {
				log.Println(err)
			}
			var ResponderRequestMessageInlibeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("ارجاع", fmt.Sprintf("referral:%s", request.TrackingCode)),
				),
			)

			msg.ReplyMarkup = ResponderRequestMessageInlibeKeyboard
		    
			category, err := repositories.GetCategoryByID(tx, db , request.CategoryID)
			if err != nil {
				log.Println(err)
			}
			text := fmt.Sprintf("متن درخواست : \n%s\n------------------\n توسط : \n%s\n------------------\n تاریخ ایجاد درخواست: \n%s\n------------------\n دسته بندی : \n%s\n------------------\n  کد رهگیری : \n%s\n", request.Text, customer.Username, request.CreatedAt, category.Name, request.TrackingCode)
			msg.Text = text

			SentMessage, err := b.bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
			err = b.rdb.SAdd(context.Background(), userKey, fmt.Sprintf("%d_%s", SentMessage.MessageID, request.TrackingCode)).Err()
			if err != nil {
				log.Panic(err)
			}
		}
		msg.Text = "برای پاسخ داده به درخواست روی آن ریپلای کنید"
		msg.ReplyMarkup = keyboards.InProgressKeyboard
		userStates[chatID] = WaittingForRequestReply

	case "getting_the_personalInfo":
		messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
		messageID, _ := b.rdb.Get(context.Background(), messageIdKey).Int()
		DeleteMessage(b, chatID, messageID)

		user, err := repositories.GetUserByChatID(tx, db , chatID)
	
		if err != nil {
			log.Println(err)
		}
		text, err := repositories.GetResponseStatistics(tx, int64(user.ID))
		if err != nil {
			log.Println(err)
		}
		msg.Text = text
		msg.ReplyMarkup = keyboards.InProgressKeyboard
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "referral:") {
		
		trackingCode := strings.TrimPrefix(update.CallbackQuery.Data, "referral:")
		trackingCodeKey := fmt.Sprintf("user:%d:trackingCode", chatID)

		request, err := repositories.GetRequestByTrackingCode(tx, db , trackingCode)
		if err != nil {
			log.Println(err)
		}

		res := b.rdb.Set(context.Background(), trackingCodeKey, trackingCode, 0)
		if err := res.Err(); err != nil {
			log.Fatal("failed to set: %w", err)
		}

		CategoriesButtons, err := keyboards.GetAllCategoriesExeptOneInButton(tx ,db, request.CategoryID)
		if err != nil {
			log.Panic(err)
		}
		if len(CategoriesButtons.InlineKeyboard) == 0 {
			msg.Text = "دسته بندی دیگری وجود ندارد"
		}else{
			msg.ReplyMarkup = CategoriesButtons
			msg.Text = "لیست دسته بندی ها"
			addToredis = true
		}
		
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") {
		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))
		category, err := repositories.GetCategoryByID(tx,db , categoryID)
		if err != nil {
			log.Println(err)
		}
		if !category.Is_active && category.Inactive_message.String != "" {
			msg := tgbotapi.NewMessage(chatID, category.Inactive_message.String)
			if _, err := b.bot.Send(msg); err != nil {
				log.Println(err)
			}
			tx.Commit() // Commit here since no further DB operations are needed
			return
		}
		messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
		messageID, _ := b.rdb.Get(context.Background(), messageIdKey).Int()
		trackingCodeKey := fmt.Sprintf("user:%d:trackingCode", chatID)
		trackingCode, _ := b.rdb.Get(context.Background(), trackingCodeKey).Result()
		request, err := repositories.GetRequestByTrackingCode(tx,db , trackingCode)
		if err != nil {
			log.Println(err)
		}
		user, err := repositories.GetUserByChatID(tx,db , chatID)
		if err != nil {
			log.Println(err)
		}

		request.Text += fmt.Sprintf("-----------------\n\n\n ارجاع شده توسط %s به دسته بندی %s \n", user.Username, category.Name)

		responder, err := repositories.GetResponserWithLowestRequest(tx, db , categoryID)
		if err != nil {
			log.Println(err)
		}

		values := []interface{}{categoryID, request.Text, responder.ID}
		columns := []string{"category_id", "text", "responser_id"}

		err = repositories.UpdateRequest(tx, db , request, values, columns...)
		if err != nil {
			log.Println(err)
		}

		updateConfig := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:        chatID,
				MessageID:     messageID,
				InlineMessageID: "",
			},
			Text: fmt.Sprintf("با موفقیت به دسته بندی %s ارجاع داده شد", category.Name),
		}

		if _, err := b.bot.Send(updateConfig); err != nil {
			panic(err)
		}
	}

	// Commit the transaction if there are no errors
	if err == nil {
		if err := tx.Commit(); err != nil {
			log.Println("Failed to commit transaction:", err)
			msg.Text = "خطا در نهایی سازی تراکنش"
			_, _ = b.bot.Send(msg)
			return
		}
	}

	if addToredis {
		SentMessage, err := b.bot.Send(msg)
		if err != nil {
			log.Println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)

		res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID, 0)
		if err := res.Err(); err != nil {
			log.Fatal("failed to set: %w", err)
		}

		if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err(); err != nil {
			log.Fatal("failed to add to orders set: %w", err)
		}
	} else {
		_, err := b.bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
