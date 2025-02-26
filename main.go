package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SushyDev/cobalt-discord-app/cobalt"
	"github.com/SushyDev/cobalt-discord-app/commands"
	"github.com/SushyDev/cobalt-discord-app/config"
	"github.com/bwmarrin/discordgo"
)

var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - app registers commands globally")
	AppToken       = flag.String("token", "", "Discord application access token")
	CobaltURL      = flag.String("cobalt", "https://api.cobalt.tools", "Base URL for Cobalt API")
	CobaltAuthKey  = flag.String("apikey", "", "API Key for Cobalt API authentication")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutting down or not")
)
var applicationCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "video",
		Description: "Download video from a social media URL",
		IntegrationTypes: &[]discordgo.ApplicationIntegrationType{discordgo.ApplicationIntegrationGuildInstall, discordgo.ApplicationIntegrationUserInstall},
		Contexts: &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild, discordgo.InteractionContextBotDM, discordgo.InteractionContextPrivateChannel},
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "URL of the video to download",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "quality",
				Description: "Video quality (144, 240, 360, 480, 720, 1080, 1440, 2160, 4320, max)",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "144p",
						Value: "144",
					},
					{
						Name:  "240p",
						Value: "240",
					},
					{
						Name:  "360p",
						Value: "360",
					},
					{
						Name:  "480p",
						Value: "480",
					},
					{
						Name:  "720p",
						Value: "720",
					},
					{
						Name:  "1080p",
						Value: "1080",
					},
					{
						Name:  "1440p (2K)",
						Value: "1440",
					},
					{
						Name:  "2160p (4K)",
						Value: "2160",
					},
					{
						Name:  "4320p (8K)",
						Value: "4320",
					},
					{
						Name:  "Maximum quality",
						Value: "max",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "mode",
				Description: "Download mode",
				Required:    false,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Auto (default)",
						Value: "auto",
					},
					{
						Name:  "Audio only",
						Value: "audio",
					},
					{
						Name:  "No audio (muted video)",
						Value: "mute",
					},
				},
			},
		},
	},
}

func init() {
	flag.Parse()
}

func main() {
	// Initialize configuration
	cfg := config.NewConfig(*AppToken, *CobaltURL, *CobaltAuthKey)

	// Create cobalt client
	cobaltClient := cobalt.NewClient(cfg.CobaltURL, cfg.CobaltAuthKey)

	// Initialize Discord session
	s, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	s.GuildIntegrationCreate("1344356700919955539", "0", "")

	// Register handlers
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Register command handler - functional approach using separate handlers module
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Only handle ApplicationCommand interactions
		if i.Type == discordgo.InteractionApplicationCommand {
			// Route to appropriate command handler based on command name
			switch i.ApplicationCommandData().Name {
			case "video":
				fmt.Println("Video command")
				commands.VideoCommand(s, i, cobaltClient)
			}
		}
	})

	// Open Discord connection
	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open Discord session: %v", err)
	}

	// Register commands
	log.Println("Registering commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(applicationCommands))
	for i, v := range applicationCommands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// Set up graceful shutdown
	fmt.Println("Cobalt Discord application is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Clean up before exiting
	if *RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	// Close Discord session
	s.Close()
}
