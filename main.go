package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/line/line-bot-sdk-go/linebot"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
key:valueを設定していき、最終的に設定した値で検索し、結果のメッセージをあなたへ送信します。
また、[key:value,key:value...]のように入力することも可能。
現在サポートしているkey一覧を知りたくばkey-allを入力してください。
※(なお、keyおよびvalue中に,や:を入力した場合、確実にparse errorになる上に検索に必要であるとは想定してないのでいちいち入力しないでください)
`, nil
	} else if message == "key-all" {
		return "key 一覧(工事中)", nil
	} else if !regexp.MustCompile(`,`).MatchString(message) {
		// 単体のkey:valueと想定。
		if strings.Count(message, ":") != 1 {
			return "", fmt.Errorf("%s is invalid format", message)
		} else {
			params := url.Values{}
			params.Add("apikey", os.Getenv("GURUNABI_SECRET"))
			kvs := strings.Split(message, ":")
			if ok := params.Get(kvs[0]); ok != "" {
				params.Del(kvs[0])
			}
			params.Add(kvs[0], kvs[1])
			return getGurunabiJSONResult(params.Encode()), nil
		}
	} else {
		params := parseKvs(message)
		params.Add("apikey", os.Getenv("GURUNABI_SECRET"))
		return getGurunabiJSONResult(params.Encode()), nil
	}
}

// レストラン検索の想定
func getGurunabiJSONResult(paramsStr string) string {
	resp, _ := http.Get("https://api.gnavi.co.jp/RestSearchAPI/v3/?" + paramsStr)
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	js, err := simplejson.NewJson(byteArray)
	if err != nil {
		log.Fatal(err)
	}
	result := ""
	fmt.Println(js.Get("rest").Map())
	// for rest := range js.Get("rest").MustArray() {
	// 	result += rest.Get("id").String() + "\n" +
	// 		rest.Get("name").String() + "\n" +
	// 		rest.Get("url_mobile").String() + "\n"
	// }
	return result
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
