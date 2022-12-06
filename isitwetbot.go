package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/caarlos0/env/v6"
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

type Config struct {
	TelegramToken    string `env:"TELEGRAM_TOKEN,notEmpty"`
	ChatId           string `env:"CHAT_ID,notEmpty"`
	AccuweatherToken string `env:"ACCUWEATHER_TOKEN,notEmpty"`
	WeatherUrl       string `env:"WEATHER_URL" envDefault:"http://dataservice.accuweather.com/forecasts/v1/minute?"`
}

func main() {
	conf, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("08:15").Do(run, conf)
	s.StartBlocking()
}

func run(conf Config) {
	var err error
	var currentForecast WeatherForecast

	if currentForecast, err = getWeather(conf.WeatherUrl, conf.AccuweatherToken); err != nil {
		log.Println(err)
	}

	if err = sendMessage(currentForecast.Summary.Phrase, conf.TelegramToken, conf.ChatId); err != nil {
		log.Println(err)
	}
}

func loadConfig() (conf Config, err error) {
	err = env.Parse(&conf)
	return
}

func sendMessage(message string, token string, chatId string) (err error) {
	log.Printf("Sending message to %v", chatId)
	messageUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?", token)
	v := url.Values{}
	v.Set("chat_id", chatId)
	v.Set("text", message)
	messageUrl += v.Encode()
	_, err = http.Get(messageUrl)
	return
}

func getWeather(weatherUrl string, token string) (currentForecast WeatherForecast, err error) {
	log.Printf("Fetching weather from %v\n", weatherUrl)
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
