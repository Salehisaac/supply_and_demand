package keyboards

import (
	"database/sql"
	"fmt"
	"github.com/Salehisaac/Supply-and-Demand.git/internal/database/repositories"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var AdminMainMessageKeyboard = 
	tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("ایجاد دسته بندی"), tgbotapi.NewKeyboardButton("لیست دسته بندی ها")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("ویرایش دسته بندی")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("انتساب افراد به دسته بندی"), tgbotapi.NewKeyboardButton("حذف فرد از دسته بندی")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("حذف دسته بندی"), tgbotapi.NewKeyboardButton("مشاهده اعضای دسته بندی (پاسخگویان)")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("دریافت آمار"), tgbotapi.NewKeyboardButton("غیر فعال کردن موقت بات")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("غیرفعال کردن موقت یک دسته بندی"), tgbotapi.NewKeyboardButton("غیرفعال کردن موقت پاسخگو")),
	)

var InProgressKeyboard = 
	tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("◀ بازگشت به منوی اصلی")),
	)


func GetAllCategoriesInButton(db *sql.DB)(tgbotapi.InlineKeyboardMarkup, error){
	categories, err := repositories.GetAllCategories(db)
	if err != nil{
		return tgbotapi.InlineKeyboardMarkup{}, err
	}

	var rows []tgbotapi.InlineKeyboardButton

	for _, category := range categories {

		responsers , err := repositories.GetCategoryMembers(db, category.ID)
		

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

