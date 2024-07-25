package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

func handleError(s *discordgo.Session, i *discordgo.InteractionCreate, errorMessage string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("**Error:** %s", errorMessage),
		},
	})
	if err != nil {
		log.Panicf("Could not respond to interaction: %s", err)
	}
}

func newInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	switch data.Name {
	case "echo":
		println("Used command 'echo'")
		handleEcho(s, i, parseOptions(data.Options))
		return
	case "join":
		println("Used command 'join'")
		handleJoin(s, i)
		return
	case "list":
		println("Used command 'list'")
		handleListChannels(s, i)
		return
	default:
		return
	}
}

func handleListChannels(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// println("Handling list channels command")
	// guild, err := s.Guild(i.GuildID)
	// if err != nil {
	// 	handleError(s, i, "Could not find Guild ID")
	// }
	// voiceStates := guild.VoiceStates
	// println(len(voiceStates))
	// channels := make(map[string]int, len(voiceStates))
	// for _, vs := range voiceStates {
	// 	println(vs.ChannelID, vs.Member.User.GlobalName)
	// 	channels[vs.ChannelID] = 1
	// 	if channels[vs.ChannelID] >= 1 {
	// 		channels[vs.ChannelID] += 1
	// 	}
	// }

	// maxChan := ""
	// maxUsers := 0
	// for ch, n := range channels {
	// 	if n > maxUsers {
	// 		maxChan = ch
	// 	}
	// }

	// if maxChan != "" {
	v, err := s.ChannelVoiceJoin(i.GuildID, "1265656310821818502", true, false)
	if err != nil {
		handleError(s, i, "Could not join voice channel")
	}
	go func() {
		time.Sleep(10 * time.Second)
		close(v.OpusRecv)
		v.Close()
	}()
	go func() {
		msg := "Joining"
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
		if err != nil {
			log.Panicf("Could not respond to interaction: %s", err)
		}
	}()
	handleVoice(v.OpusRecv)

	// }
}

func handleVoice(c chan *discordgo.Packet) {
	files := make(map[uint32]media.Writer)
	for p := range c {
		file, ok := files[p.SSRC]
		if !ok {
			var err error
			file, err = oggwriter.New(fmt.Sprintf("audio/%d.ogg", p.SSRC), 48000, 2)
			if err != nil {
				fmt.Printf("failed to create file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
				return
			}
			files[p.SSRC] = file
		}
		// Construct pion RTP packet from DiscordGo's type.
		rtp := createPionRTPPacket(p)
		err := file.WriteRTP(rtp)
		if err != nil {
			fmt.Printf("failed to write to file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
		}
	}
}

func handleEcho(s *discordgo.Session, i *discordgo.InteractionCreate, opts optionMap) {
	builder := new(strings.Builder)
	if v, ok := opts["author"]; ok && v.BoolValue() {
		author := interactionAuthor(i.Interaction)
		builder.WriteString("**" + author.String() + "** says: ")
	}
	builder.WriteString(opts["message"].StringValue())

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: builder.String(),
		},
	})

	if err != nil {
		log.Panicf("could not respond to interaction: %s", err)
	}
}

func handleJoin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// s.VoiceConnections
	// msg := fmt.Sprintf("%v", ())
	msg := "Contents: \n"
	for i, v := range s.VoiceConnections {
		msg += fmt.Sprintf("%s, %v\n", i, v)
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
	if err != nil {
		log.Panicf("Could not respond to interaction: %s", err)
	}
}
