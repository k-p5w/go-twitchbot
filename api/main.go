package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/translate"
	"github.com/gempir/go-twitch-irc/v4"
	"golang.org/x/text/language"
)

// ChatMsgInfo is チャットメッセージを保存しておくところ
type ChatMsgInfo struct {
	MsgOrg          twitch.PrivateMessage
	IsTranslateText bool
}

var RecMsg map[time.Time]ChatMsgInfo

var TextLine string

// 設定ファイルの構造体
type Config struct {
	BotName     string `json:"botName"`
	ChannelName string `json:"channelName"`
	OauthToken  string `json:"oauthToken"`
}

// 設定ファイルを読み込む関数
func loadConfig(path string) (*Config, error) {
	// 設定ファイルを読み込む
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// JSONデコード
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	fmt.Println("start-bot!")
	RecMsg = map[time.Time]ChatMsgInfo{}

	// 使用するCPUの数を指定
	runtime.GOMAXPROCS(2)

	// 並行処理の開始
	go chatBot()

	fmt.Println("HandleFunc start.")
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}

	http.HandleFunc("/", Handler)

	go func() {
		for {
			time.Sleep(5 * time.Second)
			// SVGを定期的に更新する処理
		}
	}()

	// http.ListenAndServe(":"+port, nil)
	// debugのときはこれでファイアウォールの設定がでなくなるらしい
	http.ListenAndServe("localhost:"+port, nil)

}

// Handler is /APIから呼ばれる
func Handler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Handler start.")
	// getパラメータの解析
	q := r.URL.Query()

	qColor := ""
	svgBGcolor := "black"

	qColor = q.Get("color")

	if strings.HasSuffix(qColor, ".html") {
		qColor = strings.Replace(qColor, ".html", "", -1)

	} else {
		return
	}

	fmt.Printf("qColor=%v \n", qColor)
	// 3桁か6桁なら色扱いにする（ホントはカラーコードに変換できるかチェックいるんだろうなあ）
	if len(qColor) == 3 {
		svgBGcolor = fmt.Sprintf("#%v", qColor)
	} else if len(qColor) == 6 {
		svgBGcolor = fmt.Sprintf("#%v", qColor)
	}

	// どういう状況で閲覧されているかで描画内容を変えるためにロジックをいる
	TextLineTranslateText := ""
	for _, v := range RecMsg {

		TextLine = v.MsgOrg.Message

		if !v.IsTranslateText {
			got, err := translateText("ja", TextLine)
			if err != nil {
				fmt.Printf("translateText: %v", err)
			}
			TextLineTranslateText = got
		}

	}

	fmt.Printf("MSG:%v \n", TextLine)
	fontsize := 30
	refreshCycle := 10
	text1 := TextLine

	text2 := TextLineTranslateText

	// 出来た。チャットメッセージをHTMLで取得して表示する
	// SVGにしなくていいんじゃない？？HTMLでそのまま出せば。。
	svgPageUniversal := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="ja">
<head>
  <title>SVG</title>
<style>
body {
  background-color: black;
}
h1, h2 {
	color: %v;
	font-family: "しねきゃぷしょん", "メイリオ";
	font-size: %vpx;
	margin: 0;
	padding: 0;
	line-height: 1; /* 行の高さを制御 */
	height: 30px; /* 高さを指定 */
  }
</style>
</head>
<body>
<h1>%v</h1>
<h2>%v</h2>
</body>
  <script>
  function refreshPage(seconds) {
	  setTimeout(function () {
		  location.reload();
	  }, seconds * 1000);
  }
  refreshPage(%v);
</script>
</body>
</html>
		`, svgBGcolor, fontsize, text1, text2, refreshCycle)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")
	fmt.Println(svgBGcolor)
	// fmt.Println(svgPageUniversal)
	fmt.Fprint(w, svgPageUniversal)

}

func chatBot() {
	fmt.Println("chatBot-start!")
	configPath := "config.json"

	// 設定ファイルを読み込む
	config, loaderr := loadConfig(configPath)
	if loaderr != nil {
		log.Fatalf("Failed to load config file: %s", loaderr.Error())
	}

	// Twitchのチャットボットの設定
	botUsername := config.BotName
	channel := config.ChannelName
	oauthToken := config.OauthToken
	fmt.Println("start-NewClient.")
	// or client := twitch.NewAnonymousClient() for an anonymous user (no write capabilities)
	client := twitch.NewClient(botUsername, oauthToken)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// おーここでメッセージがとれる
		fmt.Printf("bot:%v:%v\n", message.Emotes, message.Message)
		TextLine = message.Message
		var recItem ChatMsgInfo
		recItem.MsgOrg = message
		recItem.IsTranslateText = false
		RecMsg[message.Time] = recItem
		// fmt.Println(message)
		// timeLine:= message.Time
		// chatLog := message
		// recMsg = append(recMsg, chatLog)
		// つまり、これをHTML表示できれば解決だな？
		// 表示はできたので、これを配列に詰めて翻訳も含めたいところ。
	})

	// チャンネルにjoinする
	client.Join(channel)

	err := client.Connect()
	if err != nil {
		panic(err)
	}
}

func translateText(targetLanguage, text string) (string, error) {
	// text := "The Go Gopher is cute"
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %w", err)
	}

	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	resp, err := client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("Translate: %w", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("Translate returned empty response to text: %s", text)
	}
	return resp[0].Text, nil
}
