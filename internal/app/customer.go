package app

import(
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/Salehisaac/Supply-and-Demand.git/static/keyboards"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/database/repositories"
    "github.com/Salehisaac/Supply-and-Demand.git/internal/models"
	"database/sql"
	"log"
	"fmt"
	"time"
	"strings"
	"strconv"
	"context"
)



func HandleCustomerUpdateMessage(update tgbotapi.Update, b *Bot , db *sql.DB){

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")


	switch update.Message.Text{
	case "/start":
		text := `سلام به ربات خوش آمدید 👋
		امیدوارم بتونم کمکتون کنم`

		msg.Text = text
		var CustomerMainMessageInlibeKeyboard = keyboards.CustomerMainMessageInlibeKeyboard
		msg.ReplyMarkup = CustomerMainMessageInlibeKeyboard
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



	case "/help":

		msg.Text = "در خدمتتون هستم"
		if _, err := b.bot.Send(msg); err != nil {
			log.Panic(err)
		} 
	case "ثبت درخواست":
		userStates[update.Message.Chat.ID] = StateRequestSubmitted
	case "لغو درخواست":
		userStates[update.Message.Chat.ID] = ""
		msg.Text = "درخواست شما لغو شد"
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			if _, err := b.bot.Send(msg); err != nil {
				log.Panic(err)
			}
		userRequests[update.Message.Chat.ID] = nil
	}

	switch userStates[update.Message.Chat.ID] {

		case StateWaitingForRequest:
			userRequests[update.Message.Chat.ID] = append(userRequests[update.Message.Chat.ID], update.Message.Text)

			go func(chatID int64) {
				<-time.After(30 * time.Second)
				if userStates[chatID] == StateWaitingForRequest {

					var text string
					for _,message :=range userRequests[update.Message.Chat.ID]{
						text += " " + message + "\n"
					}
					user,err := repositories.GetUserByChatID(db, update.Message.Chat.ID)

					if err != nil {
						log.Fatal(err)  
					}

					categoryIdKey := fmt.Sprintf("user:%d:categoryID", update.Message.Chat.ID)
					category_id,err := b.rdb.Get(context.Background(), categoryIdKey).Int()
					if err != nil{
						panic(err)
					}

					request :=  models.Request{
						ResponserID : int64(user.ID),
						CustomerID : int64(user.ID),
						CategoryID : category_id,
						Text : text,
						TrackingCode : generateTrackingCode(),
						RequestID : nil,
						Status : "پاسخ داده نشده",
					}

					err = repositories.InsertRequest(db , request)
					if err != nil {
						log.Fatal(err)  
					}
					

					msg.Text = `درخواست شما در با موفقیت ثبت شد
					در کمترین زمان به آن پاسخ خواهیم داد
					کد رهگیری درخواست شما : ` + request.TrackingCode
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					if _, err := b.bot.Send(msg); err != nil {
						log.Panic(err)
					}
					userStates[update.Message.Chat.ID] = ""
					userRequests[update.Message.Chat.ID] = nil

				}
			}(update.Message.Chat.ID)
				
		case StateRequestSubmitted:
		if len(userRequests[update.Message.Chat.ID]) == 0 {

			msg.Text = "شما درخواستی وارد نکردید لطفا درخواستی وارد کنید"
			if _, err := b.bot.Send(msg); err != nil {
				log.Panic(err)
			}
			userStates[update.Message.Chat.ID] = StateWaitingForRequest
			break
		}
		
		var text string
		for _,message :=range userRequests[update.Message.Chat.ID]{
			text += " " + message + "\n"
		}
		user,err := repositories.GetUserByChatID(db, update.Message.Chat.ID)

		if err != nil {
			log.Fatal(err)  
		}


		categoryIdKey := fmt.Sprintf("user:%d:categoryID", update.Message.Chat.ID)
		category_id,err := b.rdb.Get(context.Background(), categoryIdKey).Int()
		if err != nil{
			panic(categoryIdKey)
		}

		request :=  models.Request{
			ResponserID : int64(user.ID),
			CustomerID : int64(user.ID),
			CategoryID : category_id,
			Text : text,
			TrackingCode : generateTrackingCode(),
			RequestID : nil,
			Status : "پاسخ داده نشده",
		}

		err = repositories.InsertRequest(db , request)
		if err != nil {
			log.Fatal(err)  
		}
		

		msg.Text = `درخواست شما در با موفقیت ثبت شد
		در کمترین زمان به آن پاسخ خواهیم داد
		کد رهگیری درخواست شما : ` + request.TrackingCode
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		if _, err := b.bot.Send(msg); err != nil {
			log.Panic(err)
		}
		userStates[update.Message.Chat.ID] = ""
		userRequests[update.Message.Chat.ID] = nil
		case StateTrackingCode:
			trackingCode := update.Message.Text
			request , err := repositories.GetRequestByTrackingCode(db, trackingCode)
			if err != nil {
			    log.Println(err)
			}

			if request == nil {
			    msg.Text = "کد رهگیری وارد شده صحیح نمی باشد"
			    if _, err := b.bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}else{
				if request.Status == "پاسخ داده نشده" {

				
					currentTime := time.Now()
					timeDifference := currentTime.Sub(request.CreatedAt)
					oneDay := 24 * time.Hour

					

					if timeDifference <= oneDay {
						msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("لغو درخواست", fmt.Sprintf("Request_cancelation:%d",request.ID)),
								tgbotapi.NewInlineKeyboardButtonData("عطف به درخواست", fmt.Sprintf("adding_to_request:%d",request.ID)),
							),
						)
					}else{
						msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("عطف به درخواست", fmt.Sprintf("adding_to_request:%d",request.ID)),
							),
						)
					}

					msg.Text = request.Text + "\n" + "تاریخ ثبت درخواست : " + request.CreatedAt.Format("2006-01-02 15:04:05") + "\n" + "کد رهگیری : " + request.TrackingCode + "\n" + "وضعیت پاسخگویی :  " + request.Status

					SentMessage, err := b.bot.Send(msg);
					
					if  err != nil {
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

					userStates[update.Message.Chat.ID]= ""
				}
			}
		case AddingToTheRequest:
			addingText := update.Message.Text

			requestIdKey := fmt.Sprintf("user:%d:requestId", update.Message.Chat.ID)
			requestID, _ := b.rdb.Get(context.Background(),requestIdKey).Int()
			request, _ := repositories.GetRequestByID(db, requestID)

			newText := request.Text + "\n" + addingText


			values := []interface{}{ newText, time.Now()}
			columns := []string{"text", "updated_at"}
			
			err := repositories.UpdateRequest(db, request, values, columns...)
			if err != nil {
				log.Println(err)
			}

			user,err := repositories.GetUserByChatID(db, update.Message.Chat.ID)

			if err != nil {
				log.Fatal(err)  
			}


			categoryIdKey := fmt.Sprintf("user:%d:categoryID", update.Message.Chat.ID)
			category_id,err := b.rdb.Get(context.Background(), categoryIdKey).Int()
			if err != nil{
				panic(categoryIdKey)
			}

			requestIdint64:= int64(requestID)

			Newrequest :=  models.Request{
				ResponserID : int64(user.ID),
				CustomerID : int64(user.ID),
				CategoryID : category_id,
				Text : addingText,
				TrackingCode : generateTrackingCode(),
				RequestID : &requestIdint64,
				Status : "پاسخ داده نشده",
			}
	
			err = repositories.InsertRequest(db , Newrequest)
			if err != nil {
				log.Fatal(err)  
			}

			msg.Text = "درخواست عطف شما با موفقیت ثبت شد"
			if _, err := b.bot.Send(msg); err != nil {
				log.Panic(err)
			}
			userStates[update.Message.Chat.ID]= ""




	}

}

func HandleCustomerUpdateCallBack(update tgbotapi.Update, b *Bot , db *sql.DB){

	
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	user, _ := repositories.GetUserByChatID(db, update.CallbackQuery.Message.Chat.ID)
	if _, err := b.bot.Request(callback); err != nil {
		panic(err)
	}


	switch update.CallbackQuery.Data{
		
		case "Request_registration":
			categories, err := repositories.GetAllCategories(db)

			var rows []tgbotapi.InlineKeyboardButton

			for _, category := range categories {
				button := tgbotapi.NewInlineKeyboardButtonData(category.Name, fmt.Sprintf("category_id:%d", category.ID))
				row := tgbotapi.NewInlineKeyboardRow(button)
				rows = append(rows, row...)
			}

			var keyboard [][]tgbotapi.InlineKeyboardButton
			for _, btn := range rows {
				keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{btn})
			}



			CategoriesButtons := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

			

			if err != nil {
				log.Panic(err) 
			}


			messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
			messageIDstr, _ := b.rdb.Get(context.Background(),messageIdKey).Result()
			messageID, _ := strconv.Atoi(messageIDstr)

			updateConfig := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:      update.CallbackQuery.Message.Chat.ID,          
					MessageID:   messageID,      
					InlineMessageID: "",         
				},
				Text:      "دسته بندی خود را انتخاب کنید", 
			}
			updateConfig.ReplyMarkup = &CategoriesButtons
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
		case "tracking_request":
			requests, err := repositories.GetLastFiveRequests(db , user.ID)
			if err != nil {
			    panic(err)
			}
			var text string
			for _,request := range requests {
			    text += "کد رهگیری : " + request.TrackingCode + "\n" + "درخواست : " + request.Text + "\n" + "وضعیت : " + request.Status + "\n" + "-------------------------------------------------------------" + "\n"
			}

			if text == "" {

				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "درخواستی برای برسی وجود ندارد")
				if _, err := b.bot.Send(msg); err != nil {
					panic(err)
				} 
			}else{
				messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
				messageIDstr, _ := b.rdb.Get(context.Background(),messageIdKey).Result()
				messageID, _ := strconv.Atoi(messageIDstr)
	
				DeleteMessage(b, update.CallbackQuery.Message.Chat.ID, messageID)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
				if _, err := b.bot.Send(msg); err != nil {
					panic(err)
				}
				msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "برای برسی درخواست خود کد پیگیری آن را وارد کنید")
				if _, err := b.bot.Send(msg); err != nil {
					panic(err)
				}
				userStates[update.CallbackQuery.Message.Chat.ID]= StateTrackingCode
			}

		
	}


	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") {

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		messageIdKey := fmt.Sprintf("user:%d:categoryID", update.CallbackQuery.Message.Chat.ID)
		res := b.rdb.Set(context.Background(), messageIdKey, categoryID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "categoryIDs", messageIdKey).Err();err !=nil{	
			log.Fatal("failed to add to categories set: %w", err)
		}

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
			

			messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
			messageIDstr, _ := b.rdb.Get(context.Background(),messageIdKey).Result()
			messageID, _ := strconv.Atoi(messageIDstr)


			DeleteMessage(b, update.CallbackQuery.Message.Chat.ID, messageID)
			endRequestButton := tgbotapi.NewKeyboardButton("ثبت درخواست")
			cancelRequestButton := tgbotapi.NewKeyboardButton("لغو درخواست")
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(endRequestButton, cancelRequestButton),
			)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "لطفا درخواست خود را وارد کنید و دکمه ی ثبت درخواست را بزنید : ")
			msg.ReplyMarkup = keyboard
			if _, err := b.bot.Send(msg); err != nil {
				panic(err)
			}
			userStates[update.CallbackQuery.Message.Chat.ID] = StateWaitingForRequest
		}
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "Request_cancelation:") {

		requestID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "Request_cancelation:"))
		request, err := repositories.GetRequestByID(db, requestID)
		

		if err != nil {
			log.Println(err)
		} 

		values := []interface{}{"لغو شده", time.Now()}
		columns := []string{"status", "updated_at"}
		
		err = repositories.UpdateRequest(db, request, values, columns...)
		if err != nil {
			log.Println(err)
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
		messageIDstr, _ := b.rdb.Get(context.Background(),messageIdKey).Result()
		messageID, _ := strconv.Atoi(messageIDstr)

		updateConfig := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      update.CallbackQuery.Message.Chat.ID,          
				MessageID:   messageID,      
				InlineMessageID: "",         
			},
			Text:      "درخواست شما لغو شد", 
		}
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

	if strings.HasPrefix(update.CallbackQuery.Data, "adding_to_request:") {

		requestID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "adding_to_request:"))
		request, err := repositories.GetRequestByID(db, requestID)
		if err != nil {
			log.Println(err)
		} 
		
		messageIdKey := fmt.Sprintf("user:%d:messageId", update.CallbackQuery.Message.Chat.ID)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()		
		DeleteMessage(b, update.CallbackQuery.Message.Chat.ID, messageID)

		
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, request.Text)
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
		
			if _, err := b.bot.Send(msg); err != nil {
				panic(err)
			}
			userStates[update.CallbackQuery.Message.Chat.ID]= AddingToTheRequest
		
		requestIdKey := fmt.Sprintf("user:%d:requestId", update.CallbackQuery.Message.Chat.ID)
		res := b.rdb.Set(context.Background(), requestIdKey, requestID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "requestIDs", requestIdKey).Err();err !=nil{	
			log.Fatal("failed to add to requests set: %w", err)
		}
	}
}