package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

type WeatherForecast struct {
	Summary    Summary
	Link       string
	MobileLink string
}

type Summary struct {
	Phrase string
	Type   string
	TypeId int
}

func main() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("08:15").Do(run)
	s.StartBlocking()
}

func run() {
	var err error
	var currentForecast WeatherForecast

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	chat_id := os.Getenv("CHAT_ID")
	accuweatherToken := os.Getenv("ACCUWEATHER_TOKEN")
	weatherUrl, ok := os.LookupEnv("WEATHER_URL")
	if !ok {
		weatherUrl = "http://dataservice.accuweather.com/forecasts/v1/minute?"
	}

	if currentForecast, err = getWeather(weatherUrl, accuweatherToken); err != nil {
		log.Println(err)
	}

	if err = sendMessage(currentForecast.Summary.Phrase, telegramToken, chat_id); err != nil {
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

func getWeather(weatherUrl string, token string) (currentForecast WeatherForecast, err error) {
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
		err = fmt.Errorf("Error with API: HTTP code = %v\n", respCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &currentForecast)

	return
}
