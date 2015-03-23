package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/howeyc/gopass"
	"github.com/mgutz/ansi"
)

func AskPassword(prompt string) Password {
	prettyPrompt := ansi.Color(prompt+": ", "blue")
	fmt.Print(prettyPrompt)

	password := Password{bytes: gopass.GetPasswdMasked()}
	return password
}

func Run(name string, args ...string) (err error) {
	cmd := exec.Command(name, args...)
	heading := ansi.Color("[command]", "magenta")
	fmt.Printf("%s %s %s\n", heading, name, strings.Join(args, " "))

	waitGroup := sync.WaitGroup{}

	type pipe func() (io.ReadCloser, error)
	readAndOutputPipe := func(inputType string, pipe pipe) (err error) {
		input, err := pipe()
		if err != nil {
			return
		}

		waitGroup.Add(1)

		heading := ansi.Color("    [%s]", "yellow")
		heading = fmt.Sprintf(heading, inputType)
		scanner := bufio.NewScanner(input)

		go func() {
			for scanner.Scan() {
				fmt.Println(heading, scanner.Text())
			}
			waitGroup.Done()
		}()

		return
	}

	err = readAndOutputPipe("out", cmd.StdoutPipe)
	if err != nil {
		return
	}

	err = readAndOutputPipe("err", cmd.StderrPipe)
	if err != nil {
		return
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	waitGroup.Wait()
	err = cmd.Wait()

	var exitCode int
	if msg, ok := err.(*exec.ExitError); ok {
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	} else {
		exitCode = 0
	}

	heading = ansi.Color("   [exit]", "cyan")
	fmt.Println(heading, exitCode)

	if err != nil {
		return
	}

	return
}

type Handler (func(args ...string) error)

var (
	commands = map[string]Handler{
		"init":        Initialize,
		"generate-ca": GenerateCA,
	}
)

func Initialize(args ...string) (err error) {
	Run("git", "init")
	return
}

func GenerateCA(args ...string) (err error) {
	password := AskPassword("CA Password")
	openSSLPassword, err := password.WritePasswordFile()
	defer password.CleanUp()

	if err != nil {
		return
	}

	err = Run("openssl", "genrsa", "-aes128", "-passout", openSSLPassword, "-out", "ca.key", "4096")
	if err != nil {
		return
	}

	err = Run("openssl", "req", "-new", "-x509", "-days", "365", "-key", "ca.key", "-passin", openSSLPassword, "-out", "ca.crt", "-subj", "/CN=Heracles Intranet/emailAddress=support@example.com")
	if err != nil {
		return
	}

	return
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: heracles <command> [<args>]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("   init          Creates an empty heracles repository")
		fmt.Println("   generate-ca   Generates a certificate authority, to sign server & client certificates")
		os.Exit(1)
	}

	command := os.Args[1]
	if handler, ok := commands[command]; ok {
		err := handler()
		if err == nil {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	} else {
		fmt.Printf("heracles: '%s' is not a heracles command.\n", command)
	}
}
