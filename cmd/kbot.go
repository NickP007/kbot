/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	telebot "gopkg.in/telebot.v3"
)

var (
	// TeleToken bot
	TeleToken = os.Getenv("TELE_TOKEN")
)

// kbotCmd represents the kbot command
var kbotCmd = &cobra.Command{
	Use:     "kbot",
	Aliases: []string{"start"},
	Short:   "Start a bot",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s started\n", AppVersion)

		kbot, err := telebot.NewBot(telebot.Settings{
			URL:    "",
			Token:  TeleToken,
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		})

		if err != nil {
			log.Fatalf("Plaese check TELE_TOKEN env variable. %s", err)
			return
		}

		kbot.Handle(telebot.OnText, func(m telebot.Context) error {
			payload := m.Message().Payload
			msg_text := m.Text()
			log.Print(payload, " - ", msg_text)

			switch payload {
			case "hello":
				err = m.Send(fmt.Sprintf("<b>Hello, %s</b>\nI'm %s!", m.Sender().FirstName, AppVersion), telebot.ModeHTML)
			case "":
				switch msg_text {
				case "/start":
					err = m.Send("<b>Usage:</b>\n /help - for help message\n hello - to view 'hello message'\n ping - get 'Pong' response", telebot.ModeHTML)
				case "/help":
					err = m.Send("NP Kbot help page... be soon")
				case "/hello", "hello":
					err = m.Send(fmt.Sprintf("<b>Hello, %s</b>\nI'm %s!", m.Sender().FirstName, AppVersion), telebot.ModeHTML)
				case "ping":
					err = m.Send("Pong")
				}
			default:
				err = m.Send("<b>Usage:</b>\n /help - for help message\n hello - to view 'hello message'\n ping - get 'Pong' response", telebot.ModeHTML)
			}

			return err
		})
		kbot.Start()
	},
}

func init() {
	rootCmd.AddCommand(kbotCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// kbotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// kbotCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
