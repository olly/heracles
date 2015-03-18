package main

import "bufio"
import "fmt"
import "io"
import "io/ioutil"
import "os"
import "os/exec"
import "os/user"
import "strconv"
import "strings"
import "sync"
import "syscall"
import "github.com/mgutz/ansi"
import "github.com/howeyc/gopass"

type Password struct {
	bytes        []byte
	passwordFile *string
}

func (password Password) Clear() {
	for i, _ := range password.bytes {
		password.bytes[i] = byte(0)
	}
}

func (password Password) String() string {
	return string(password.bytes)
}

func (password *Password) WritePasswordFile() (string, error) {
	tmpfile, err := ioutil.TempFile("", "password")
	if err != nil {
		return "", err
	}

	user, _ := user.Current()
	uid, _ := strconv.Atoi(user.Uid)
	gid, _ := strconv.Atoi(user.Gid)
	tmpfile.Chown(uid, gid)
	tmpfile.Chmod(0600)
	tmpfile.Write(password.bytes)
	tmpfile.Chmod(0400)
	tmpfile.Close()

	name := tmpfile.Name()
	password.passwordFile = &name
	openSSLPasswordArg := "file:" + name

	return openSSLPasswordArg, nil
}

func (password *Password) CleanUp() {
	if password.passwordFile != nil {
		os.Remove(*password.passwordFile)
		password.passwordFile = nil
	}
}

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
