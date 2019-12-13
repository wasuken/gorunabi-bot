package main

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
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
		return `基本的に[key:value]で入力することになります。例:freeword:ラーメン
また、[key:value,key:value...]のように入力することも可能。なお、keyおよびvalue中に,や:を入力した場合、
確実にparse errorになる上に検索に必要であるとは想定してないのでいちいち入力しないでください。
現在サポートしているkey一覧を知りたくばkey-allを入力してください。
`, nil
	} else if message == "key-all" {
		return "key 一覧(工事中)", nil
	} else if !regexp.MustCompile(`,`).MatchString(message) {
		// 単体のkey:valueと想定。
		if strings.Count(message, ":") != 1 {
			return "", fmt.Errorf("%s is invalid format", message)
		} else {
			return "工事中です", nil
		}
	} else {
		return "工事中です", nil
	}
}
