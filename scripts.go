package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"bytes"
)

type CommandLogger interface {
	LogCommand(server string, channel string, from string, cmd string)
}

type ScriptsConfig struct {
	AdminScripts   string
	PublicScripts  string
	PrivateScripts string
	LocalPort      string
}

// avoid characters such as "../" to disallow commands like "!../admin/kick"
//var re_cmd = regexp.MustCompile("^!([a-zA-Z0-9]+)( .*)?")
var re_cmd = regexp.MustCompile("^([a-zA-Z0-9]+)(,?:? ?.*)?")
var re_bz  = regexp.MustCompile(`https\:\/\/bugzilla.redhat.com\/show_bug.cgi\?id=(\w+)|#`)
var re_bug = regexp.MustCompile(`bug(.+)`)

func fileExists(cmd string) bool {
	_, err := os.Stat(cmd)
	return err == nil
}

func cmdPath(config ScriptsConfig, cmd string, admin bool, private bool) string {
	if private {
		path := fmt.Sprintf("%s/%s.cmd", config.PrivateScripts, cmd)
		if fileExists(path) {
			return path
		}
		return ""
	}
	if admin {
		path := fmt.Sprintf("%s/%s.cmd", config.AdminScripts, cmd)
		if fileExists(path) {
			return path
		}
	}
	path := fmt.Sprintf("%s/%s.cmd", config.PublicScripts, cmd)
	if fileExists(path) {
		fmt.Sprintf("cmdPath %s is not found", path)
		return path
	}
	return ""
}

func execCmd(config ScriptsConfig, path string, ev Event) {
	log.Printf("Executing [%s]\n", path)

	in_params := strings.Split(ev.Data, " ")
	dynamic_hostname := strings.Split(config.LocalPort, ":")
        var stderr bytes.Buffer

	command := exec.Command(path,
		dynamic_hostname[1],
		ev.Server,
		ev.Channel,
		ev.User)

	for _, v := range in_params[1:] {
		command.Args = append(command.Args, v)
	}

        command.Stderr = &stderr
	err := command.Run()
        if err != nil {
                log.Printf("\x1b[1;31mError:\n%s\n%s\x1b[0m\n", stderr.String(), err)
        } else {
                log.Printf("Success to execute\n")
        }
}

func execCmdArgs(config ScriptsConfig, path string, ev Event, args []string) {
        log.Printf("Executing [%s]\n", path)

        dynamic_hostname := strings.Split(config.LocalPort, ":")
	var stderr bytes.Buffer

        command := exec.Command(path,
                dynamic_hostname[1],
                ev.Server,
                ev.Channel,
                ev.User)

        for _, v := range args {
                command.Args = append(command.Args, v)
        }

	command.Stderr = &stderr
        err := command.Run()
        if err != nil {
                log.Printf("\x1b[1;31mError:\n%s\n%s\x1b[0m\n", stderr.String(), err)
        } else {
                log.Printf("Success to execute\n")
	}
}


// creates a new action from what was sent on the admin port
func netAdminCraftAction(output string) Action {
	var a Action
	shellapi := strings.SplitN(output, " ", 3)
	a.Type = A_RAW
	if len(shellapi) == 3 {
		a.Server = shellapi[0]
		a.Priority, _ = strconv.Atoi(shellapi[1])
		if a.Priority != PRIORITY_LOW &&
			a.Priority != PRIORITY_MEDIUM &&
			a.Priority != PRIORITY_HIGH {
			a.Priority = PRIORITY_LOW
		}
		a.Data = shellapi[2]
	} else {
		a.Data = output
		a.Priority = PRIORITY_LOW
	}
	return a
}

// shell commands can send several commands in the same connection
// (using \r\n)
func netAdminReadFromCon(con *net.TCPConn, chac chan Action) {
	const NBUF = 512
	var rawcmd []byte
	var buf [NBUF]byte

	for {
		n, err := con.Read(buf[0:])
		rawcmd = append(rawcmd, buf[0:n]...)
		if err != nil {
			break
		}
	}
	con.Close()
	msgs := strings.Split(string(rawcmd), "\n")
	for i := 0; i < len(msgs); i++ {
		if len(msgs[i]) > 0 {
			s := strings.TrimRight(msgs[i], " \r\n\t")
			chac <- netAdminCraftAction(s)
		}
	}
}

// opens the admin port and directly send RAW commands to grobot
func netAdmin(config ScriptsConfig, chac chan Action) {
	a, err := net.ResolveTCPAddr("tcp", config.LocalPort)
	if err != nil {
		log.Fatalf("Can't resolve: %v\n", err)
	}
	listener, err := net.ListenTCP("tcp", a)
	if err != nil {
		log.Fatalf("Can't open admin port: %v\n", err)
	}
	for {
		con, err := listener.AcceptTCP()
		if err == nil {
			go netAdminReadFromCon(con, chac)
		}
	}
}

func Scripts(chac chan Action, chev chan Event, logger CommandLogger, config ScriptsConfig) {
	go netAdmin(config, chac)
	for {
		e, ok := <-chev

		if !ok {
			log.Printf("Channel closed")
			return
		}

		switch e.Type {
		case E_PRIVMSG:
			if e.CmdId != 0 {
				if m := re_cmd.FindStringSubmatch(e.Data); len(m) > 0 {
					path := cmdPath(config, m[1],
						e.AdminCmd,
						len(e.Channel) == 0)
					args := []string{strings.TrimLeft(m[2], ":, \t")}
					if len(path) > 0 {
						logger.LogCommand(e.Server, e.Channel, e.User, m[1])
						go execCmdArgs(config, path, e, args)
					}
				}
			}
			if m := re_bz.FindStringSubmatch(e.Data); len(m) > 0 {
				path := cmdPath(config, "bugzilla",
					e.AdminCmd,
					len(e.Channel) == 0)
				if len(path) > 0 {
					logger.LogCommand(e.Server, e.Channel, e.User, m[1])
					go execCmd(config, path, e)
				}
			} else if m := re_bug.FindStringSubmatch(e.Data); len(m) > 0 {
				n := m
				var res []string

				for {
					n = regexp.MustCompile(`(\d+)(.*)`).FindStringSubmatch(n[len(n)-1])
					if len(n) > 1 {
						res = append(res, n[1])
					} else {
						break
					}
				}

				if len(res) == 0 {
					continue
				}
				path := cmdPath(config, "bugzilla",
					e.AdminCmd,
					len(e.Channel) == 0)
				if len(path) > 0 {
					logger.LogCommand(e.Server, e.Channel, e.User, m[1])
					go execCmdArgs(config, path, e, res)
				}
                        }
		}
	}
}
