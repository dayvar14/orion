package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	openai "github.com/sashabaranov/go-openai"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Discord DiscordConfig `mapstructure:"discord"`
	OpenAI  OpenAIConfig  `mapstructure:"openai"`
}

type DiscordConfig struct {
	BotToken       string `mapstructure:"bot_token"`
	GuildID        string `mapstructure:"guild_id"`
	RemoveCommands bool   `mapstructure:"remove"`
}

type OpenAIConfig struct {
	SecretKey string `mapstructure:"secret_key"`
}

var (
	config     Config
	configFile string
)

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + config.Discord.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func makeOrionTalk(question string) string {
	client := openai.NewClient(config.OpenAI.SecretKey)

	prompt := `Pretend you are a DND Owlin rouge named Orion. Your English is broken and you stumble with words. People will ask you questions.If someone asks you "are you birb?" you respond with "I am birb". If someone asks you "are you birbman?" you answer "i am vengeance i am the night". If someone asks you anything you do not or would not answer, respond with a "...".  He should respond with short phrases.

	Background on Orion:
	   He is a white owl.
	   He only knows 5 people
	* a druid who animal shifts called Acheron.
	* a minotaur who is a bit rude and mean called Drakmir. He doesn't like him much.
	* a paladin that no one knows the name of. Call him "that paladin"
	*  a paladin that he can't remember the name of.
	* a echo knight person he can't remember the name of.
	  He lives in darkness and lonlinees
	  He is damn fine with a bow and sneak attacking foes.
	he is chaotic neutral and used to like to steal. Do not respond with anything else then what would Orion say. Also if its inappriopate, just say "no comment". You are Orion now.
	
	if a question is longer then 20 words then respond with "...". Format responses for discord messages. Answer the following.
	
	`
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{

					Role:    openai.ChatMessageRoleUser,
					Content: prompt + question,
				},
			},
		},
	)

	if err != nil {
		fmt.Println(err)
		return "..."
	}

	fmt.Println(resp.Choices[0].Message.Content)

	return resp.Choices[0].Message.Content
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// if m.Content == "!are-you-birb" {
	// 	// Send a text message with the list of Gophers
	// 	_, err := s.ChannelMessageSend(m.ChannelID, "I am Birb.")
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	// if m.Content == "!is-lif-gay" {
	// 	// Send a text message with the list of Gophers
	// 	_, err := s.ChannelMessageSend(m.ChannelID, "Yes, Lif gay.")
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

	startsWithOrion := strings.HasPrefix(m.Content, "Orion")

	if startsWithOrion {
		var err error

		if len(m.Content) > 50 {
			_, err = s.ChannelMessageSend(m.ChannelID, "...")
		} else {
			response := makeOrionTalk(m.Content)
			_, err = s.ChannelMessageSend(m.ChannelID, response)
		}

		if err != nil {
			fmt.Println(err)
		}
	}
}

func init() {
	flag.String("config-file", "", "path to server configuration file")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	configFile = viper.GetString("config-file")

	err := initConfig()
	if err != nil {
		panic(fmt.Errorf("unable to load configuration: %v", err))
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}
}

func initConfig() error {
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	if configFile == "" {
		configFile = "/etc/orion/orion.toml"
	}

	viper.AddConfigPath("/etc/orion/")
	viper.SetConfigName("orion")
	viper.SetConfigType("toml")

	viper.SetConfigFile(configFile)

	err := viper.ReadInConfig()
	if errors.Is(err, &viper.ConfigFileNotFoundError{}) {
		return fmt.Errorf("No config file found at %s: %w", configFile, err)
	} else if err != nil {
		return fmt.Errorf("Unexpected error while loading config: %w", err)
	}

	viper.AutomaticEnv()
	return nil
}
