package main

import (
	"fmt"
	"io/ioutil"
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
	accuweatherToken := os.Getenv("ACCUWEATHER_TOKEN")

	if err = sendMessage(message, telegramToken, chat_id); err != nil {
		log.Println(err)
	}

	if err = getWeather(accuweatherToken); err != nil {
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

func getWeather(token string) (err error) {
	weatherUrl := "http://dataservice.accuweather.com/forecasts/v1/minute?"
	v := url.Values{}
	v.Set("q", "51.46,-2.6")
	v.Set("apikey", token)
	v.Set("language", "en-GB")
	weatherUrl += v.Encode()

	resp, err := http.Get(weatherUrl)
	if err != nil {
		return
	}

	if respCode := resp.StatusCode; respCode >= 400 {
		err = fmt.Errorf("Error with API: HTTP code = %v", respCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	return
}
