package gorobot

import (
	"botapi"
	"os"
	"netchan"
	"fmt"
	"log"
)

type GoRobot struct {
	Config *Config
	LogMap map[string] *os.File
	Irc* Irc
	Exp *netchan.Exporter
	Modules map[string] chan botapi.Event
	Actions chan botapi.Action
}

// Creates a new robot from a configuration file, automatically
// connect to servers listed in the configuration file
func NewGoRobot(config string) *GoRobot {
	robot := GoRobot{
		Config: NewConfig(config),
		LogMap: make(map[string] *os.File),
		Irc: NewIrc(),
		Modules: make(map[string] chan botapi.Event),
	}
	robot.Exp = botapi.InitExport(robot.Config.Module.Interface)
	robot.Actions = botapi.ExportActions(robot.Exp)
	for  k, v := range robot.Config.Servers {
		v.Name = k
		robot.Irc.Connect(v)
	}
	return &robot
}

func (robot *GoRobot) SendEvent(event *botapi.Event) {
	for _, chev := range robot.Modules {
		chev <- *event
	}
	robot.LogEvent(event)
}

// Based on PING events from servers, ugly but enough for now
func (robot *GoRobot) Cron() {
	robot.LogStatistics()
}

// Autojoin channels on a given server
func (robot *GoRobot) AutoJoin(s string) {
	serv := robot.Irc.GetServer(s)
	if serv != nil {
		for k, _ := range serv.Config.Channels {
			serv.JoinChannel(k)
		}
	}
}

// Handle a notice
func (robot *GoRobot) HandleNotice(serv *Server, event *botapi.Event) {
	switch event.CmdId {
	case 1:
		robot.AutoJoin(serv.Config.Name)
	}
}

func (robot *GoRobot) HandleEvent(serv *Server, event *botapi.Event) {
	switch event.Type {
	case botapi.E_KICK :
		if serv.Config.Nickname == event.Data && robot.Config.AutoRejoinOnKick {
			serv.JoinChannel(event.Channel)
		}
	case botapi.E_PING :
		serv.SendMeRaw[botapi.PRIORITY_HIGH] <- fmt.Sprintf("PONG :%s\r\n", event.Data)
		robot.Cron()
	case botapi.E_NOTICE :
		robot.HandleNotice(serv, event)
	case botapi.E_PRIVMSG :
		if _, ok := serv.Config.Channels[event.Channel]; ok == true {
			event.AdminCmd = serv.Config.Channels[event.Channel].Master
		}
	}
	robot.SendEvent(event)
}

func (robot *GoRobot) NewModule(ac *botapi.Action) {
	robot.Modules[ac.Data] = botapi.ExportEvents(robot.Exp, ac.Data)
}

func (robot *GoRobot) HandleAction(ac *botapi.Action) {
	// if the command is RAW, we need to parse it first to be able
	// to correctly handle it.
	if ac.Type == botapi.A_RAW {
		new_action := ExtractAction(ac)
		if new_action != nil {
			p := ac.Priority
			*ac = *new_action
			ac.Priority = p
		} else {
			log.Printf("raw command ignored [%s]", ac.Raw)
			return
		}
	}

	switch ac.Type {
	case botapi.A_NEWMODULE:
		robot.NewModule(ac)
	case botapi.A_SAY:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.Say(ac)
		}
	case botapi.A_JOIN:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.JoinChannel(ac.Channel)
		}
	case botapi.A_PART:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.LeaveChannel(ac.Channel, ac.Data)
		}
	case botapi.A_KICK:
		if serv := robot.Irc.GetServer(ac.Server); serv != nil {
			serv.KickUser(ac.Channel, ac.User, ac.Data)
		}
	}
}

func (robot *GoRobot) Run() {
	for {
		select {
		case action := <-robot.Actions:
			robot.HandleAction(&action)
		case event := <-robot.Irc.Events:
			srv := robot.Irc.GetServer(event.Server)
			if srv != nil {
				robot.HandleEvent(srv, &event)
			}
		}
	}
}
