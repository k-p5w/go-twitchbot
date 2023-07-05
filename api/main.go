package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gempir/go-twitch-irc/v4"
)

func main() {

	configPath := "config.json"

	// 設定ファイルを読み込む
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config file: %s", err.Error())
	}

	// Twitchのチャットボットの設定
	botUsername := config.botName
	channel := config.channelName
	oauthToken := config.oauthToken

	fmt.Println("start-bot!")

	// or client := twitch.NewAnonymousClient() for an anonymous user (no write capabilities)
	client := twitch.NewClient(botUsername, oauthToken)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// おーここでメッセージがとれる
		fmt.Printf("bot:%v:%v\n", message.Emotes, message.Message)
		// つまり、これをHTML表示できれば解決だな？
	})

	// チャンネルにjoinする
	client.Join(channel)

	err := client.Connect()
	if err != nil {
		panic(err)
	}

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
