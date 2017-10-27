package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/coreos/go-systemd/dbus"
)

type service struct {
	name        string
	description string
	state       string
	load        string
	sub         string
}

var SERVICE string

var UNIT string

var DBUSCONN *dbus.Conn

func serviceCompleter(d prompt.Document) []prompt.Suggest {
	units := getAllUnits()
	var s []prompt.Suggest
	for _, u := range units {
		s = append(s, prompt.Suggest{
			Text:        u.Name,
			Description: fmt.Sprintf("%s [%s]", u.Description, u.SubState),
		})
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func connectToDbus() error {
	conn, err := dbus.NewSystemConnection()
	if err != nil {
		return err
	}

	DBUSCONN = conn
	return nil
}

func getAllUnits() []dbus.UnitStatus {
	u, err := DBUSCONN.ListUnits()
	if err != nil {
		fmt.Errorf("Error getting all units fro dbus: %s\n", err)
		return nil
	}
	return u
}

func executor(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}

	if s == "logs" {
		fmt.Print(journalctl(UNIT))
		return
	}

	if s == "followlogs" {
		followlogs(UNIT)
		return
	}

	fmt.Print(systemctlRun(UNIT, s))
	return
}

func actionCompleter(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "status", Description: "Get status of service"},
		{Text: "start", Description: "Launch service"},
		{Text: "stop", Description: "Terminate service"},
		{Text: "logs", Description: "Show service logs"},
		{Text: "followlogs", Description: "Follow service logs"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func listUnits() {
	units := getAllUnits()
	for _, u := range units {
		fmt.Printf("%s [%s] %s\n", u.Name, u.SubState, u.Description)
	}
	return
}

func unitAvailable(name string) bool {
	return true
}

func main() {
	err := connectToDbus()
	if err != nil {
		fmt.Printf("Failed to connect to systemd: %s\n", err)
		os.Exit(1)
	}
	defer DBUSCONN.Close()

	fmt.Println("Welcome to systemd-repl. Connected to systemd. Quit with Ctrl+D")

	fmt.Println("Select service")

	var in string
	for UNIT == "" {
		in = prompt.Input("> ", serviceCompleter)
		if in == "list-units" {
			listUnits()
		} else if unitAvailable(in) {
			UNIT = in
		} else {
			fmt.Printf("Unit '%s' unavailable\n", in)
		}
	}

	fmt.Println("Selected " + UNIT)

	p := prompt.New(
		executor,
		actionCompleter,
		prompt.OptionPrefix(UNIT+"> "),
		prompt.OptionTitle("systemd REPL"),
	)
	p.Run()

}

func systemctlRun(service, action string) string {
	cmd := exec.Command("systemctl", action, service)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return string(out)
}

func journalctl(service string) string {
	cmd := exec.Command("journalctl", "-u", service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return string(out)
}

func followlogs(service string) {
	cmd := exec.Command("journalctl", "-fu", service)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()
	return
}
