package main

import "github.com/bwmarrin/discordgo"

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "list",
		Description: "List all the channels in this guild",
	},
	{
		Name:        "stop",
		Description: "stop listening",
	},
	{
		Name:        "join",
		Description: "Join the vc of sender",
	},
	{
		Name:        "echo",
		Description: "Say through a bot",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "message",
				Description: "Contents of the message",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
			{
				Name:        "author",
				Description: "Whether to prepend message's author",
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Required:    false,
			},
		},
	},
}
