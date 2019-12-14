package main

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"gorunabi-bot/api"
	"gorunabi-bot/masterAPI"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	api_base_url string = "https://api.gnavi.co.jp"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "master" {
			masterAllUpdate()
		} else if os.Args[1] == "create" {
			masterCreate()
		}
	} else {
		server()
	}
}
func masterAllUpdate() {
	masterAPI.GetGAreaSmallSearchResponse(api_base_url)
}
func masterCreate() {
	masterAPI.CreateTables()
}

func server() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)

	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					sendMsg, e := parse(message.Text)
					if e != nil {
						log.Print(e)
					}
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(sendMsg)).Do(); err != nil {
						log.Print(err)
					}
				}
			} else if event.Type == linebot.EventTypeFollow {
				sendMsg, _ := parse("help")
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(sendMsg)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	})
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

func parse(message string) (string, error) {
	if message == "help" {
		f, err := os.Open("help.txt")
		if err != nil {
			log.Fatal("error")
		}
		defer f.Close()
		// 一気に全部読み取り
		b, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal("error")
		}
		return string(b), nil
	} else {
		if strings.Count(message, ":") != 1 {
			return "", fmt.Errorf("%s is invalid format", message)
		} else {
			params := url.Values{}
			kvs := strings.Split(message, ":")
			if kvs[0] == "検索" {
				params.Add("freeword", strings.Join(kvs[1:], ""))
				return api.GetGurunabiJSONResult(api_base_url, params.Encode()), nil
			}
			return "", fmt.Errorf("%s is invalid format", message)
		}
	}
}

func parseKvs(kvsStr string) url.Values {
	result := url.Values{}
	for _, kvsStr := range strings.Split(kvsStr, ",") {
		kvs := strings.Split(kvsStr, ":")
		if ok := result.Get(kvs[0]); ok != "" {
			result.Del(kvs[0])
		}
		result.Add(kvs[0], kvs[1])
	}
	return result
}
