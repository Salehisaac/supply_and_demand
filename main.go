package main

import (
   
    "log"
    "github.com/Salehisaac/Supply-and-Demand.git/internal/app"
     _"github.com/go-telegram-bot-api/telegram-bot-api/v5"


)



func main() {

    //Creating new Instance of Bot
    bot, err := app.NewBot()
    if err != nil {
        log.Fatalf("Error initializing bot: %v", err)
    }

    // Starting the Bot
    log.Println("Starting bot...")
    if err := bot.Start(); err != nil {
        log.Fatalf("Error starting bot: %v", err)
    }
}
