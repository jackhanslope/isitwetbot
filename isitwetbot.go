package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-co-op/gocron"
	tele "gopkg.in/telebot.v3"
)

type WeatherForecast struct {
	Summary    Summary
	Link       string
	MobileLink string
}

func (forecast WeatherForecast) sendString() string {
	if forecast.Summary.Phrase == "No precipitation for at least 120 min" {
		return forecast.Summary.Phrase
	} else {
		return fmt.Sprintf("[%s\n](%s)", forecast.Summary.Phrase, forecast.Link)
	}
}

type Summary struct {
	Phrase string
	Type   string
	TypeId int
}

type Config struct {
	TelegramToken    string `env:"TELEGRAM_TOKEN,notEmpty"`
	ChatId           int64  `env:"CHAT_ID,notEmpty"`
	AccuweatherToken string `env:"ACCUWEATHER_TOKEN,notEmpty"`
	WeatherUrl       string `env:"WEATHER_URL" envDefault:"http://dataservice.accuweather.com/forecasts/v1/minute?"`
	NtfyTopic        string `env:"NTFY_TOPIC,notEmpty"`
}

func main() {
	conf, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	teleSettings := tele.Settings{
		Token:  conf.TelegramToken,
		Poller: &tele.LongPoller{Timeout: time.Minute},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		responderAgent(teleSettings, conf)
	}()
	go func() {
		defer wg.Done()
		schedulerAgent(teleSettings, conf)
	}()

	wg.Wait()
}

func responderAgent(teleSettings tele.Settings, conf Config) (err error) {
	bot, err := tele.NewBot(teleSettings)
	if err != nil {
		return err
	}
	bot.Handle("/weather", func(context tele.Context) (err error) {
		user := context.Sender()
		if user.ID != conf.ChatId {
			unauthString := fmt.Sprintf("Unauthorized attempted access from user: %s", user.Username)

			log.Printf(unauthString)

			req, _ := http.NewRequest(
				"POST",
				fmt.Sprintf("https://ntfy.sh/%s", conf.NtfyTopic),
				strings.NewReader(unauthString),
			)
			req.Header.Set("Title", "isitwetbot")
			http.DefaultClient.Do(req)

			return context.Send("You are not authorized to use this bot.")
		} else {
			log.Printf("Weather requested from authorized user %s", user.Username)
		}

		forecast, err := conf.getWeather()
		if err != nil {
			return err
		}

		err = context.Send(forecast.sendString(), "Markdown", tele.NoPreview)
		log.Printf("Sent requested forecast to %s", user.Username)
		return
	})
	bot.Start()
	return
}

func schedulerAgent(teleSettings tele.Settings, conf Config) (err error) {
	scheduler := gocron.NewScheduler(time.UTC)
	job, err := scheduler.Every(1).Day().At("08:15").Do(sendScheduled, teleSettings, conf)
	if err != nil {
		log.Fatal(err)

	}
	scheduler.StartAsync()
	nextRun := job.NextRun()
	log.Println("Next scheduled send at", nextRun.Format("15:04:05 on Mon 02 Jan 2006"))
	scheduler.StartBlocking()
	return
}

func sendScheduled(teleSettings tele.Settings, conf Config) (err error) {
	log.Println("Starting scheduled send")

	bot, err := tele.NewBot(teleSettings)
	if err != nil {
		return
	}

	user := tele.User{ID: conf.ChatId}

	forecast, err := conf.getWeather()
	if err != nil {
		return
	}

	message, err := bot.Send(&user, forecast.sendString(), "Markdown", tele.NoPreview)
	log.Printf("Sent scheduled forecast to %s", message.Chat.Username)
	return
}

func loadConfig() (conf Config, err error) {
	log.Println("Loading config")
	err = env.Parse(&conf)
	return
}

func (conf Config) getWeather() (WeatherForecast, error) {
	return getWeather(conf.WeatherUrl, conf.AccuweatherToken)
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
		err = fmt.Errorf("error with API: HTTP code = %v\n", respCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &currentForecast)

	return
}
