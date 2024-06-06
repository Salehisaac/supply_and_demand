package keyboards

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var ResponderMainMessageInlibeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("پاسخ به درخواست", "answer_to_request"),
        tgbotapi.NewInlineKeyboardButtonData("دریافت آمار عملکرد شخصی", "getting_the_personalInfo"),
    ),
)


var ResponderRequestMessageInlibeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("ارجاع", "referral"),
    ),
)