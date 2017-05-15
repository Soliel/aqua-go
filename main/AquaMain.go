package main

import (
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"strings"
	log "github.com/inconshreveable/log15"
)

/*const (
	Space for constant variables here
)*/

var (
	conf       *config
	BotID      string
	CmdHandler *commandHandler
)

type config struct {
	BotToken  string `json:"bot_token"`
	BotPrefix string `json:"bot_prefix"`
}

func main() {

	conf = loadConfig("config.json")

	//create discord session, create a database connection, and check for errors.
	dg, err := discordgo.New("Bot " + conf.BotToken)

	if err != nil {
		log.Error("Unable to establish connection to discord","Error", err)
		return
	}

	CmdHandler = newCommandHandler()
	registerCommands()

	//select bot user.
	u, err := dg.User("@me")
	if err != nil {
		log.Error("Could not get bot username.","Error", err)
	}

	dg.AddHandler(onMessageReceived)

	err = dg.Open()
	if err != nil {
		log.Error("Unable to open discord.","Error", err)
		return
	}

	log.Info("Bot is now running", "User", u.Username)

	<-make(chan struct{})
	return
}

func onMessageReceived(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == BotID {
		return
	}

	if len(m.Content) < len(conf.BotPrefix) {
		return
	}

	if m.Content[:len(conf.BotPrefix)] != conf.BotPrefix {
		return
	}

	content := m.Content[len(conf.BotPrefix):]
	if len(content) < 1 {
		return
	}

	content = strings.ToLower(content)

	var args []string

	//String math to allow Spaces in between quotations.
	if strings.Contains(content, "\"") {
		tempArgs := strings.Split(content, "\"")
		for s := range tempArgs {
			tempArgs[s] = strings.TrimSpace(tempArgs[s])
			if tempArgs[s] != "" {
				args = append(args, tempArgs[s])
			}
		}

		args[0] = strings.TrimPrefix(args[0], " ")
		args[0] = strings.TrimSuffix(args[0], " ")
		if strings.Contains(args[0], " ") {
			firstArgs := strings.Fields(args[0])
			args = append(firstArgs[:], args[1:]...)
		}
	} else {
		args = strings.Fields(content)
	}
	name := args[0]

	command, found := CmdHandler.get(name)
	if !found {
		return
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		fmt.Println("Error getting channel, ", err)
		return
	}

	//set up my context to pass to whatever function is called.
	ctx := new(context)
	ctx.Args = args[1:]
	ctx.Session = s
	ctx.Msg = m
	ctx.Channel = channel

	guild, err := s.State.Guild(channel.GuildID)
	if err == nil {
		ctx.Guild = guild
	}


	//pass command pointer and run the function
	c := command.CmdFunc
	go c(*ctx)
}

func registerCommands() {
	CmdHandler.register("ping", testcommand)
}

func loadConfig(filename string) *config {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error loading config, ", err)
		return nil
	}

	var confData config
	err = json.Unmarshal(body, &confData)
	if err != nil {
		fmt.Println("Error parsing JSON data, ", err)
		return nil
	}
	return &confData
}

func testcommand(ctx context) {
	ctx.Session.ChannelMessageSend(ctx.Channel.ID, "Pong!")
}