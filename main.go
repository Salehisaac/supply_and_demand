package main

import (
   
    "log"
    "github.com/Salehisaac/Supply-and-Demand.git/internal/app"
     _"github.com/go-telegram-bot-api/telegram-bot-api/v5"


)



func main() {
    botToken := "6427555441:AAEgmiblShsDaDyVlkSclFeOVqS1XCbOe-8"
    bot, err := app.NewBot(botToken)
    if err != nil {
        log.Fatalf("Error initializing bot: %v", err)
    }

    
    log.Println("Starting bot...")
    if err := bot.Start(); err != nil {
        log.Fatalf("Error starting bot: %v", err)
    }
}
