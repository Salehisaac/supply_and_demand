package keyboards

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var CustomerMainMessageInlibeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
    tgbotapi.NewInlineKeyboardRow(
        tgbotapi.NewInlineKeyboardButtonData("ثبت درخواست", "Request_registration"),
        tgbotapi.NewInlineKeyboardButtonData("پیگیری درخواست", "tracking_request"),
    ),
)



