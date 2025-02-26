package commands

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/SushyDev/cobalt-discord-app/cobalt"
	"github.com/SushyDev/cobalt-discord-app/utils"
	"github.com/bwmarrin/discordgo"
)

// VideoCommand handles the /video slash command
func VideoCommand(s *discordgo.Session, i *discordgo.InteractionCreate, client *cobalt.Client) {
	// Acknowledge the interaction immediately
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Processing your video request...",
		},
	})
	
	if err != nil {
		utils.LogError("Error responding to interaction", err)
		return
	}
	
	// Extract command options
	options := i.ApplicationCommandData().Options
	
	// Get URL (required option)
	// url := options[0].StringValue()
	
	// Build request options from command options
	requestOptions := buildRequestOptions(options)
	
	// Process the video download
	processAndSendVideo(s, i, client, requestOptions)
}

// buildRequestOptions constructs cobalt.RequestOptions from command options
func buildRequestOptions(options []*discordgo.ApplicationCommandInteractionDataOption) cobalt.RequestOptions {
	// Get URL (required option)
	url := options[0].StringValue()
	
	// Set default values for optional parameters
	requestOptions := cobalt.RequestOptions{
		URL:          url,
		VideoQuality: "1080",
		DownloadMode: "auto",
	}
	
	// Get optional values if provided
	for _, opt := range options {
		switch opt.Name {
		case "quality":
			if opt.Value != nil {
				requestOptions.VideoQuality = opt.StringValue()
			}
		case "mode":
			if opt.Value != nil {
				requestOptions.DownloadMode = opt.StringValue()
			}
		}
	}
	
	return requestOptions
}

// processAndSendVideo processes a video request and sends the response
func processAndSendVideo(s *discordgo.Session, i *discordgo.InteractionCreate, client *cobalt.Client, options cobalt.RequestOptions) {
	// Process the video download
	result, err := client.ProcessVideo(options)
	if err != nil {
		// Send error response
		editResponseWithError(s, i, "Error processing video", err)
		return
	}
	
	// Handle the result based on the status
	switch result.Status {
	case "redirect", "tunnel":
		handleDownloadableMedia(s, i, client, result)
	case "picker":
		handleMediaPicker(s, i, result)
	case "error":
		handleErrorResponse(s, i, result)
	default:
		// Handle unexpected response
		editResponse(s, i, fmt.Sprintf("Received unexpected response status: %s", result.Status))
	}
}

// handleDownloadableMedia handles redirect and tunnel responses
func handleDownloadableMedia(s *discordgo.Session, i *discordgo.InteractionCreate, client *cobalt.Client, result *cobalt.CobaltResponse) {
	// Download the file
	videoData, filename, err := client.DownloadFile(result.URL, result.Filename)
	if err != nil {
		editResponseWithError(s, i, "Error downloading video", err)
		return
	}
	
	// Try to send the file
	err = sendVideoFile(s, i, videoData, filename)
	if err != nil {
		// If sending fails, provide a direct link
		editResponse(s, i, fmt.Sprintf("The video is too large to send directly. You can download it here: %s", result.URL))
	}
}

// handleMediaPicker handles picker responses with multiple media items
func handleMediaPicker(s *discordgo.Session, i *discordgo.InteractionCreate, result *cobalt.CobaltResponse) {
	content := "Multiple media items found. Use the links below to download:\n\n"
	
	// Check if there's common audio
	if result.Audio != "" {
		content += fmt.Sprintf("**Common Audio**: [%s](%s)\n\n", result.AudioFilename, result.Audio)
	}
	
	// List all media items
	for i, item := range result.Picker {
		content += fmt.Sprintf("%d. %s: [Download](%s)\n", i+1, strings.ToUpper(item.Type), item.URL)
	}
	
	editResponse(s, i, content)
}

// handleErrorResponse handles error responses from cobalt API
func handleErrorResponse(s *discordgo.Session, i *discordgo.InteractionCreate, result *cobalt.CobaltResponse) {
	errorMsg := fmt.Sprintf("Error: %s", result.Error.Code)
	
	// Add service info if available
	if result.Error.Context != nil && result.Error.Context.Service != "" {
		errorMsg += fmt.Sprintf(" (Service: %s)", result.Error.Context.Service)
	}
	
	// Add limit info if available
	if result.Error.Context != nil && result.Error.Context.Limit > 0 {
		errorMsg += fmt.Sprintf(" (Limit: %d)", result.Error.Context.Limit)
	}
	
	editResponse(s, i, errorMsg)
}

// sendVideoFile sends a video file as a response
func sendVideoFile(s *discordgo.Session, i *discordgo.InteractionCreate, videoData []byte, filename string) error {
	reader := bytes.NewReader(videoData)
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr("Here's your video:"),
		Files: []*discordgo.File{
			{
				Name:   filename,
				Reader: reader,
			},
		},
	})
	
	if err != nil {
		log.Printf("Error editing interaction with file: %v", err)
		
		// Try to send as a followup if the file is too large for an edit
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Here's your video:",
			Files: []*discordgo.File{
				{
					Name:   filename,
					Reader: bytes.NewReader(videoData),
				},
			},
		})
	}
	
	return err
}

// editResponse edits the interaction response with a string message
func editResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr(content),
	})
	
	if err != nil {
		log.Printf("Error editing interaction response: %v", err)
	}
}

// editResponseWithError edits the interaction response with an error message
func editResponseWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string, err error) {
	log.Printf("%s: %v", message, err)
	
	_, respErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr(fmt.Sprintf("%s: %v", message, err)),
	})
	
	if respErr != nil {
		log.Printf("Error editing interaction response: %v", respErr)
	}
}

// Helper function to get pointer to string
func stringPtr(s string) *string {
	return &s
}
