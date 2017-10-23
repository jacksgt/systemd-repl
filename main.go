package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"os/exec"
	"strings"
)

type service struct {
	name        string
	description string
	state       string
	load        string
	sub         string
}

var SERVICE string

func serviceCompleter(d prompt.Document) []prompt.Suggest {
	services := getAllServices()
	var s []prompt.Suggest
	for _, e := range services {
		s = append(s, prompt.Suggest{
			Text:        e.name,
			Description: fmt.Sprintf("%s [%s]", e.description, e.state),
		})
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func getAllServices() []service {
	var services []service

	cmd := exec.Command("systemctl", "list-units", "--plain")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get services from `systemctl list-units`: %s\n", err)
		return nil
	}

	for _, line := range strings.Split(strings.TrimSuffix(string(out), "\n"), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			fmt.Printf("Error, too few fields in line: %s\n", line)
			continue
		}
		unit := fields[0]
		load := fields[1]
		active := fields[2]
		sub := fields[3]
		description := strings.Join(fields[4:], " ")

		var s = service{
			name:        unit,
			load:        load,
			state:       active,
			sub:         sub,
			description: description,
		}

		services = append(services, s)
	}

	return services
}

func executor(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}

	if s == "logs" {
		fmt.Print(journalctl(SERVICE))
		return
	}

	fmt.Print(systemctlRun(SERVICE, s))
	return
}

func actionCompleter(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "status", Description: "Get status of service"},
		{Text: "start", Description: "Launch service"},
		{Text: "stop", Description: "Terminate service"},
		{Text: "logs", Description: "Show service logs"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	fmt.Println("Welcome to systemd-repl. Quit with Ctrl+D")

	fmt.Println("Select service")
	SERVICE = prompt.Input("> ", serviceCompleter)
	fmt.Println("Selected " + SERVICE)

	p := prompt.New(
		executor,
		actionCompleter,
		prompt.OptionPrefix("> "),
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
