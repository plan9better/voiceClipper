package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

var (
	isRecording     bool
	recordMutex     sync.Mutex
	voiceConnection *discordgo.VoiceConnection
)

var v *discordgo.VoiceConnection

func handleError(s *discordgo.Session, i *discordgo.InteractionCreate, errorMessage string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Error: %s", errorMessage),
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
		isRecording = true
		handleListChannels(s, i)
		return
	case "stop":
		println("used command 'stop'")
		isRecording = false
		handleStop(s, i)

	default:
		return
	}
}

func handleStop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if v != nil {
		if v.OpusRecv != nil {
			close(v.OpusRecv)
			println("Closed opusrecv")
		}
		v.Close()
		println("Closed v")
	}
	go func() {
		msg := "stopping"
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
		if err != nil {
			log.Panicf("Could not respond to interaction: %s", err)
		}
	}()
}

func handleListChannels(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error
	v, err = s.ChannelVoiceJoin(i.GuildID, "1265656310821818502", true, false)
	if err != nil {
		handleError(s, i, "Could not join voice channel")
	}
	go func() {
		var filename byte
		filename = 0

		for isRecording {
			handleVoice(v.OpusRecv, filename)
			time.Sleep(10 * time.Second)
			close(v.OpusRecv)
			filename += 1
			filename %= 10
		}
		v.Close()
	}()
	// handleVoice(v.OpusRecv)
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

}

func handleVoice(c chan *discordgo.Packet, filename byte) {
	files := make(map[uint32]media.Writer)
	for p := range c {
		file, ok := files[p.SSRC]
		if !ok {
			var err error
			file, err = oggwriter.New(fmt.Sprintf("audio/%d.ogg", filename), 48000, 2)
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
		builder.WriteString("" + author.String() + " says: ")
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
