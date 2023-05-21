package views

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/xremming/abborre/esox"
	"golang.org/x/oauth2"
)

func URLs(cfg aws.Config, tableName string, oauth2Config oauth2.Config) esox.URLs {
	return esox.URLs{{
		Name:    "home",
		Handler: Home(),
		Path:    "/",
	}, {
		Name:    "events.list",
		Handler: EventsList(cfg, tableName),
		Path:    "/events",
	}, {
		Name:    "events.list.ics",
		Handler: EventsListICS(cfg, tableName),
		Path:    "/events.ics",
	}, {
		Name:    "events.update",
		Handler: EventsUpdate(cfg, tableName),
		Path:    "/events/update",
	}, {
		Name:    "events.create",
		Handler: EventsCreate(cfg, tableName),
		Path:    "/events/create",
	}, {
		Name:    "discord.login",
		Handler: DiscordLogin(oauth2Config),
		Path:    "/discord/login",
	}, {
		Name:    "discord.callback",
		Handler: DiscordCallback(oauth2Config),
		Path:    "/discord/callback",
	}}
}
