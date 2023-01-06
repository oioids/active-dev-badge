package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var BotToken string

// If you don't add a guild ID, the commands will be global. Just invite the bot to one server and you'll be fine
var GuildID string

// If you want to remove the commands when the bot is stopped, set this to true
var RemoveCommands bool = true

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	if _, err := os.Stat("token.txt"); err == nil {
		f, err := os.Open("token.txt")
		if err != nil {
			log.Fatalf("Cannot open file: %v", err)
		}
		defer f.Close()
		_, err = fmt.Fscanln(f, &BotToken)
		if err != nil {
			log.Fatalf("Cannot read from file: %v", err)
		}
	} else if os.IsNotExist(err) {

		log.Println("Token file does not exist. Creating one...")
		log.Println("Enter the bot's token:")
		fmt.Scanln(&BotToken)

		f, err2 := os.Create("token.txt")
		if err2 != nil {
			log.Fatalf("Cannot create file: %v", err2)
		}
		defer f.Close()
		_, err2 = f.WriteString(BotToken)
		if err2 != nil {
			log.Fatalf("Cannot write to file: %v", err2)
		}

	} else {
		// Some other error
		log.Fatalf("Cannot open file: %v", err)
	}

	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "finish",
			Description: "Run this command to get your Active Developer Badge",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"finish": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey there!" +
						"\nClick this link to get your active developer badge." +
						"\nhttps://discord.com/developers/active-developer" +
						"\nKeep in mind that you may need to wait 24 hours before you can get your badge." +
						"\n\nRemember that your bot **must have \"Use data to improve Discord\" turned on. `User Settings > Privacy`**",
				},
			})
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Println("If you see this message, it means that you have not entered the bot's token. Please, enter it in the token.txt file OR delete the file entirely.")
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if RemoveCommands {
		log.Println("Removing commands...")

		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
