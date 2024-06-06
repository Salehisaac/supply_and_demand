package keyboards

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Salehisaac/Supply-and-Demand.git/internal/database/repositories"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)




var InProgressKeyboard = 
	tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("◀️ بازگشت به منوی اصلی")),
	)


func GetAllCategoriesInButton(tx *sql.Tx ,db *sql.DB)(tgbotapi.InlineKeyboardMarkup, error){
	categories, err := repositories.GetAllCategories(tx ,db)
	if err != nil{
		return tgbotapi.InlineKeyboardMarkup{}, err
	}

	var rows []tgbotapi.InlineKeyboardButton

	for _, category := range categories {

		responsers , err := repositories.GetCategoryMembers(tx ,db, category.ID)
		

		if err != nil{
			return tgbotapi.InlineKeyboardMarkup{}, err
		}
		button_text := fmt.Sprintf("نام : %v  _  تعداد اعضا : %d", category.Name, len(responsers))
		button := tgbotapi.NewInlineKeyboardButtonData(button_text, fmt.Sprintf("category_id:%d", category.ID))
		row := tgbotapi.NewInlineKeyboardRow(button)
		rows = append(rows, row...)
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, btn := range rows {
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{btn})
	}


	
	CategoriesButtons := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	return CategoriesButtons , nil

}

func GetAllCategoriesExeptOneInButton(tx *sql.Tx , db *sql.DB , categoryId int)(tgbotapi.InlineKeyboardMarkup, error){

	categories, err := repositories.GetAllCategories(tx , db)
	if err != nil{
		return tgbotapi.InlineKeyboardMarkup{}, err
	}

	var rows []tgbotapi.InlineKeyboardButton

	for _, category := range categories {

		if category.ID == categoryId {
			continue 
		}

		responsers , err := repositories.GetCategoryMembers(tx ,db, category.ID)
		

		if err != nil{
			return tgbotapi.InlineKeyboardMarkup{}, err
		}
		button_text := fmt.Sprintf("نام : %v  _  تعداد اعضا : %d", category.Name, len(responsers))
		button := tgbotapi.NewInlineKeyboardButtonData(button_text, fmt.Sprintf("category_id:%d", category.ID))
		row := tgbotapi.NewInlineKeyboardRow(button)
		rows = append(rows, row...)
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, btn := range rows {
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{btn})
	}


	
	CategoriesButtons := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	return CategoriesButtons , nil

}

func GetActiveCategoriesInButton(tx *sql.Tx ,db *sql.DB)(tgbotapi.InlineKeyboardMarkup, error){
	categories, err := repositories.GetActiveCategories(tx ,db)
	if err != nil{
		return tgbotapi.InlineKeyboardMarkup{}, err
	}

	var rows []tgbotapi.InlineKeyboardButton

	for _, category := range categories {

		responsers , err := repositories.GetCategoryMembers(tx ,db, category.ID)
		

		if err != nil{
			return tgbotapi.InlineKeyboardMarkup{}, err
		}
		button_text := fmt.Sprintf("نام : %v  _  تعداد اعضا : %d", category.Name, len(responsers))
		button := tgbotapi.NewInlineKeyboardButtonData(button_text, fmt.Sprintf("category_id:%d", category.ID))
		row := tgbotapi.NewInlineKeyboardRow(button)
		rows = append(rows, row...)
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, btn := range rows {
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{btn})
	}


	
	CategoriesButtons := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	return CategoriesButtons , nil

}

func GetInactiveCategoriesInButton(tx *sql.Tx , db *sql.DB)(tgbotapi.InlineKeyboardMarkup, error){
	categories, err := repositories.GetInactiveCategories(tx,db)
	if err != nil{
		return tgbotapi.InlineKeyboardMarkup{}, err
	}

	var rows []tgbotapi.InlineKeyboardButton

	for _, category := range categories {

		responsers , err := repositories.GetCategoryMembers(tx ,db, category.ID)
		

		if err != nil{
			return tgbotapi.InlineKeyboardMarkup{}, err
		}
		button_text := fmt.Sprintf("نام : %v  _  تعداد اعضا : %d", category.Name, len(responsers))
		button := tgbotapi.NewInlineKeyboardButtonData(button_text, fmt.Sprintf("category_id:%d", category.ID))
		row := tgbotapi.NewInlineKeyboardRow(button)
		rows = append(rows, row...)
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, btn := range rows {
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{btn})
	}


	
	CategoriesButtons := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	return CategoriesButtons , nil

}

func GetAllCategoryMembersInlineButton(db *sql.DB , users []models.User)(tgbotapi.InlineKeyboardMarkup, error){

	var rows []tgbotapi.InlineKeyboardButton

	for _, user := range users {

		button_text := user.Username
		button := tgbotapi.NewInlineKeyboardButtonData(button_text, fmt.Sprintf("user_id:%d", user.ID))
		row := tgbotapi.NewInlineKeyboardRow(button)
		rows = append(rows, row...)
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, btn := range rows {
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{btn})
	}


	
	CategoryUsersButtons := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	return CategoryUsersButtons , nil

}

func GetAdminMainMessageKeyboard(db *sql.DB)(tgbotapi.ReplyKeyboardMarkup, error) {
	botSettings, err := repositories.GetTheBotConfigs(db)

	if err != nil {
		log.Println(err)
		return tgbotapi.ReplyKeyboardMarkup{}, err
	}

	
	var botStatusButtonText string
	if botSettings.Is_active {
		botStatusButtonText = "غیر فعال کردن موقت بات"
	} else {
		botStatusButtonText = "فعال کردن بات"
	}
	var AdminMainMessageKeyboard = 
	tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("ایجاد دسته بندی"), tgbotapi.NewKeyboardButton("لیست دسته بندی ها")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("ویرایش دسته بندی"),tgbotapi.NewKeyboardButton("حذف دسته بندی")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("افزودن عضو به دسته بندی"), tgbotapi.NewKeyboardButton("حذف عضو از دسته بندی")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("دریافت آمار"), tgbotapi.NewKeyboardButton(botStatusButtonText)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("فعال کردن دسته بندی"), tgbotapi.NewKeyboardButton("غیرفعال کردن موقت یک دسته بندی")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("فعال کردن پاسخگو"), tgbotapi.NewKeyboardButton("غیرفعال کردن موقت پاسخگو")),
	)

	return AdminMainMessageKeyboard , nil

}

