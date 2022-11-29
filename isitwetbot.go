package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	var err error

	message := "Yes, it's probably wet"
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	chat_id := os.Getenv("CHAT_ID")

	if err = sendMessage(message, telegramToken, chat_id); err != nil {
		log.Println(err)
	}
}

func sendMessage(message string, token string, chat_id string) (err error) {
	messageUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?", token)
	v := url.Values{}
	v.Set("chat_id", chat_id)
	v.Set("text", message)
	messageUrl += v.Encode()
	_, err = http.Get(messageUrl)
	return
}
