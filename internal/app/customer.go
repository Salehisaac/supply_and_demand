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

	chatID := update.Message.Chat.ID
	msg := tgbotapi.NewMessage(chatID, "")
	msg.ReplyMarkup = keyboards.InProgressKeyboard
	var addToredis bool 
	

	switch update.Message.Text{
		case "/start":
			msg.Text = WelcomeMessage
			var CustomerMainMessageInlibeKeyboard = keyboards.CustomerMainMessageInlibeKeyboard
			msg.ReplyMarkup = CustomerMainMessageInlibeKeyboard
			addToredis = true
			userStates[chatID] = ""
			userRequests[chatID] = nil
		case "◀️ بازگشت به منوی اصلی" :
			msg.Text = WelcomeMessage
			var CustomerMainMessageInlibeKeyboard = keyboards.CustomerMainMessageInlibeKeyboard
			msg.ReplyMarkup = CustomerMainMessageInlibeKeyboard
			addToredis = true
			userStates[chatID] = ""
			userRequests[chatID] = nil



		case "ثبت درخواست":
			userStates[chatID] = StateRequestSubmitted
		case "لغو درخواست":
			msg.Text = "درخواست شما لغو شد"
			userStates[chatID] = ""
			userRequests[chatID] = nil
	}

	switch userStates[chatID] {

		case StateWaitingForRequest:
			userRequests[chatID] = append(userRequests[chatID], update.Message.Text)

			go func(chatID int64) {
				<-time.After(5 * time.Minute)
				if userStates[chatID] == StateWaitingForRequest {

					var text string
					for _,message :=range userRequests[chatID]{
						text += " " + message + "\n"
					}
					user,err := repositories.GetUserByChatID(tx ,db, chatID)

					if err != nil {
						log.Fatal(err)  
					}

					categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatID)
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
						Status : "unseen",
					}

					err = repositories.InsertRequest(tx ,db , request)
					if err != nil {
						log.Fatal(err)  
					}
					if tx != nil {
						err = tx.Commit()
						if err != nil {
							log.Println("Error committing transaction:", err)
							return
						}
					}
					

					msg.Text = `درخواست شما در با موفقیت ثبت شد
					در کمترین زمان به آن پاسخ خواهیم داد
					کد رهگیری درخواست شما : ` + request.TrackingCode
					userStates[chatID] = ""
					userRequests[chatID] = nil

				}
			}(chatID)
				
		case StateRequestSubmitted:
		if len(userRequests[chatID]) == 0 {
			msg.Text = "شما درخواستی وارد نکردید لطفا درخواستی وارد کنید"
			userStates[chatID] = StateWaitingForRequest
			break
		}
		
		var text string
		for _,message :=range userRequests[chatID]{
			text += " " + message + "\n"
		}
		user,err := repositories.GetUserByChatID(tx ,db, chatID)
		if err != nil {
			log.Println(err)  
		}

		categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatID)
		category_id,err := b.rdb.Get(context.Background(), categoryIdKey).Int()
		if err != nil{
			log.Println(err)
		}

		responder , err := repositories.GetResponserWithLowestRequest(tx ,db , category_id)
		if err != nil{
			log.Println(err)
		}

		request :=  models.Request{
			ResponserID : int64(responder.ID),
			CustomerID : int64(user.ID),
			CategoryID : category_id,
			Text : text,
			TrackingCode : generateTrackingCode(),
			RequestID : nil,
			Status : "unseen",
		}

		err = repositories.InsertRequest(tx ,db , request)
		if err != nil {
			log.Println(err)  
		}

		if tx != nil {
			err = tx.Commit()
			if err != nil {
				log.Println("Error committing transaction:", err)
				return
			}
		}


		

		msg.Text = `درخواست شما در با موفقیت ثبت شد
		در کمترین زمان به آن پاسخ خواهیم داد
		کد رهگیری درخواست شما : ` + request.TrackingCode
		userStates[chatID] = ""
		userRequests[chatID] = nil

		case StateTrackingCode:
			trackingCode := update.Message.Text
			request , err := repositories.GetRequestByTrackingCode(tx ,db, trackingCode)
			if err != nil {
			    log.Println(err)
			}

			if request == nil {
			    msg.Text = "کد رهگیری وارد شده صحیح نمی باشد"
			    break
			}
			if request.Status == "unseen" {

				
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
				addToredis = true
				userStates[chatID]= ""
				}else if(request.Status == "answerd" ){
					msg.Text = request.Text + "\n" + "تاریخ ثبت درخواست : " + request.CreatedAt.Format("2006-01-02 15:04:05") + "\n" + "کد رهگیری : " + request.TrackingCode + "\n" + "وضعیت پاسخگویی :  " + request.Status
					_, err := b.bot.Send(msg);
					if  err != nil {
						log.Panic(err)
					}

					msg.Text = "پاسخ درخواست 👇🏻"
					_, err = b.bot.Send(msg);
					if  err != nil {
						log.Panic(err)
					}

					answers , err := repositories.GetRequestAnswers(tx , request.ID)
					if  err != nil {
						log.Panic(err)
					}

					for _ , answer := range answers {
						msg.Text = answer.Text
					}

					userStates[chatID]= ""
					}
			
		case AddingToTheRequest:
				addingText := update.Message.Text
				
				requestIdKey := fmt.Sprintf("user:%d:requestId", chatID)
				requestID, err := b.rdb.Get(context.Background(), requestIdKey).Int()
				if err != nil {
					log.Println("Error getting request ID from Redis:", err)
					return
				}
				
				request, err := repositories.GetRequestByID(db, requestID)
				if err != nil {
					log.Println("Error getting request by ID:", err)
					return
				}
				
				request_responser_id := request.ResponserID
				newText := request.Text + "\n" + addingText
				
				values := []interface{}{newText, time.Now()}
				columns := []string{"text", "updated_at"}
				
				err = repositories.UpdateRequest(tx, db, request, values, columns...)
				if err != nil {
					log.Println("Error updating request:", err)
					if tx != nil {
						tx.Rollback()
					}
					return
				}
				
				log.Println("Request Updated successfully")
				
				user, err := repositories.GetUserByChatID(tx, db, chatID)
				if err != nil {
					log.Println("Error getting user by chat ID:", err)
					if tx != nil {
						tx.Rollback()
					}
					return
				}
				
				categoryIdKey := fmt.Sprintf("user:%d:categoryID", chatID)
				category_id, err := b.rdb.Get(context.Background(), categoryIdKey).Int()
				if err != nil {
					log.Println("Error getting category ID from Redis:", err)
					if tx != nil {
						tx.Rollback()
					}
					return
				}
				
				requestIdint64 := int64(requestID)
				Newrequest := models.Request{
					ResponserID:  request_responser_id,
					CustomerID:   int64(user.ID),
					CategoryID:   category_id,
					Text:         addingText,
					TrackingCode: generateTrackingCode(),
					RequestID:    &requestIdint64,
					Status:       "unseen",
				}
				
				err = repositories.InsertRequest(tx, db, Newrequest)
				if err != nil {
					log.Println("Error inserting new request:", err)
					if tx != nil {
						tx.Rollback()
					}
					return
				}
				
				if tx != nil {
					err = tx.Commit()
					if err != nil {
						log.Println("Error committing transaction:", err)
						return
					}
				}
				msg.Text = "درخواست عطف شما با موفقیت ثبت شد"
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

func HandleCustomerUpdateCallBack(update tgbotapi.Update, b *Bot , db *sql.DB){

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

	chatID := update.CallbackQuery.Message.Chat.ID
	user, _ := repositories.GetUserByChatID(tx ,db, chatID)
	msg := tgbotapi.NewMessage(chatID, "")
	msg.ReplyMarkup = keyboards.InProgressKeyboard
	var addToredis bool 



	switch update.CallbackQuery.Data{
		
		case "Request_registration":
			categories, err := repositories.GetAllCategories(tx ,db)

			if len(categories) == 0 {

			messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()
			
			updateConfig := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:      chatID,          
					MessageID:   messageID,      
					InlineMessageID: "",         
				},
				Text:      "دسته بندی موجود نمیباشد", 
			}
			
		 	_,err := b.bot.Send(updateConfig);
			if err != nil {
				log.Println(err)
			}
			}else{
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
				log.Println(err) 
			}

			messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()

			updateConfig := tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					ChatID:      chatID,          
					MessageID:   messageID,      
					InlineMessageID: "",         
				},
				Text:      "دسته بندی خود را انتخاب کنید", 
			}
			updateConfig.ReplyMarkup = &CategoriesButtons
			SentMessage, err := b.bot.Send(updateConfig);
			if err != nil {
				log.Println(err) 
			}
			
			res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
			if err := res.Err();err != nil {
				log.Println("failed to set: %w", err)
			}
		
			if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
				log.Println("failed to add to orders set: %w", err)
			}
		}
		case "tracking_request":
			requests, err := repositories.GetLastFiveRequests(db , user.ID)
			if err != nil {
			    log.Println(err) 
			}
			var text string
			for _,request := range requests {
			    text += "کد رهگیری : " + request.TrackingCode + "\n" + "درخواست : " + request.Text + "\n" + "وضعیت : " + request.Status + "\n" + "-------------------------------------------------------------" + "\n"
			}

			if text == "" {

				msg := tgbotapi.NewMessage(chatID, "درخواستی برای برسی وجود ندارد")
				if _, err := b.bot.Send(msg); err != nil {
					log.Println(err) 
				} 
			}else{
				messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
				messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()

				updateConfig := tgbotapi.EditMessageTextConfig{
					BaseEdit: tgbotapi.BaseEdit{
						ChatID:      chatID,          
						MessageID:   messageID,      
						InlineMessageID: "",         
					},
					Text: text, 
				}
				if _, err := b.bot.Send(updateConfig); err != nil {
					log.Println(err) 
				}
				msg.Text = "برای برسی درخواست خود کد پیگیری آن را وارد کنید"
				userStates[chatID]= StateTrackingCode
			}

		
	}


	if strings.HasPrefix(update.CallbackQuery.Data, "category_id:") {

		categoryID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "category_id:"))

		messageIdKey := fmt.Sprintf("user:%d:categoryID", chatID)
		res := b.rdb.Set(context.Background(), messageIdKey, categoryID,0)
		if err := res.Err();err != nil {
			log.Println("failed to set: %w", err)
		}
		if err := b.rdb.SAdd(context.Background(), "categoryIDs", messageIdKey).Err();err !=nil{	
			log.Println("failed to add to categories set: %w", err)
		}

		category, err := repositories.GetCategoryByID(tx ,db, categoryID)
		if err != nil {
			log.Println(err)
		} 

		if !category.Is_active && category.Inactive_message.String != "" {
			msg := tgbotapi.NewMessage(chatID, category.Inactive_message.String)
			if _, err := b.bot.Send(msg); err != nil {
				log.Println(err)
			}
		}else{
			

			messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
			messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()
			DeleteMessage(b, chatID, messageID)

			endRequestButton := tgbotapi.NewKeyboardButton("ثبت درخواست")
			cancelRequestButton := tgbotapi.NewKeyboardButton("لغو درخواست")
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(endRequestButton, cancelRequestButton),
			)
			msg.Text = "لطفا درخواست خود را وارد کنید و دکمه ی ثبت درخواست را بزنید : "
			msg.ReplyMarkup = keyboard
			userStates[chatID] = StateWaitingForRequest
		}
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "Request_cancelation:") {

		requestID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "Request_cancelation:"))
		request, err := repositories.GetRequestByID(db, requestID)
		
		if err != nil {
			log.Println(err)
		} 

		values := []interface{}{"canceld", time.Now()}
		columns := []string{"status", "updated_at"}
		
		err = repositories.UpdateRequest(tx ,db, request, values, columns...)
		if err != nil {
			log.Println(err)
		}

		if tx != nil {
			err = tx.Commit()
			if err != nil {
				log.Println("Error committing transaction:", err)
				return
			}
		}

		messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()



		updateConfig := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      chatID,          
				MessageID:   messageID,      
				InlineMessageID: "",         
			},
			Text:      "درخواست شما لغو شد", 
		}
		SentMessage, err := b.bot.Send(updateConfig);
			if err != nil {
				log.Println(err)
			}
			res := b.rdb.Set(context.Background(), messageIdKey, SentMessage.MessageID,0)
			if err := res.Err();err != nil {
				log.Println("failed to set: %w", err)
			}
		
			if err := b.rdb.SAdd(context.Background(), "messsageIDs", messageIdKey).Err();err !=nil{	
				log.Println("failed to add to orders set: %w", err)
			}
	}

	if strings.HasPrefix(update.CallbackQuery.Data, "adding_to_request:") {

		requestID, _ := strconv.Atoi(strings.TrimPrefix(update.CallbackQuery.Data, "adding_to_request:"))
		request, err := repositories.GetRequestByID(db, requestID)
		if err != nil {
			log.Println(err)
		} 
		
		messageIdKey := fmt.Sprintf("user:%d:messageId", chatID)
		messageID, _ := b.rdb.Get(context.Background(),messageIdKey).Int()		
		DeleteMessage(b, chatID, messageID)

		
		msg.Text = request.Text
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
		userStates[chatID]= AddingToTheRequest
		
		requestIdKey := fmt.Sprintf("user:%d:requestId", chatID)
		res := b.rdb.Set(context.Background(), requestIdKey, requestID,0)
		if err := res.Err();err != nil {
			log.Fatal("failed to set: %w", err)
		}
	
		if err := b.rdb.SAdd(context.Background(), "requestIDs", requestIdKey).Err();err !=nil{	
			log.Fatal("failed to add to requests set: %w", err)
		}
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