package main

import (
	"fmt"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/wasuken/gorunabi-bot/api"
	"github.com/wasuken/gorunabi-bot/masterAPI"
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
	var converted_message string
	converted_message = strings.Replace(message, "：", ":", -1)
	converted_message = strings.Replace(converted_message, "　", " ", -1)
	if converted_message == "help" {
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
		if strings.Count(converted_message, ":") != 1 {
			return "", fmt.Errorf("%s is invalid format", converted_message)
		} else {
			params := url.Values{}
			kvs := strings.Split(converted_message, ":")
			if kvs[0] == "検索" && strings.TrimSpace(kvs[0]) != "" {
				params.Add("freeword", strings.Join(kvs[1:], ""))
				added_map := masterAPI.SearchMasterDataMakeKeyValues(strings.Join(kvs[1:], ""))
				for _, v := range masterAPI.DB_NAME_LIST {
					if added_map[masterAPI.DB_NAME_API_KEY_MAP[v]] != "" {
						params.Add(masterAPI.DB_NAME_API_KEY_MAP[v], added_map[masterAPI.DB_NAME_API_KEY_MAP[v]])
						break
					}
				}
				return api.GetGurunabiJSONResult(api_base_url, params.Encode()), nil
			}
			return "", fmt.Errorf("%s is invalid format", converted_message)
		}
	}
}
