package main

import "fmt"
import "io/ioutil"
import "os"
import "os/user"
import "strconv"
import _ "github.com/mgutz/ansi"
import _ "github.com/howeyc/gopass"

type Password struct {
	bytes []byte
}

func (password Password) Clear() {
	for i, _ := range password.bytes {
		password.bytes[i] = byte(0)
	}
}

func (password Password) String() string {
	return string(password.bytes)
}

func (password Password) WithTempFile(block func(string)) error {
	tmpfile, err := ioutil.TempFile("", "password")
	if err != nil {
		return err
	}

	user, _ := user.Current()
	uid, _ := strconv.Atoi(user.Uid)
	gid, _ := strconv.Atoi(user.Gid)
	tmpfile.Chown(uid, gid)
	tmpfile.Chmod(0600)
	tmpfile.Write(password.bytes)
	tmpfile.Chmod(0400)
	tmpfile.Close()
	block(tmpfile.Name())

	os.Remove(tmpfile.Name())

	return nil
}

var (
	commands = map[string](func(args... string) int){}
)
func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: heracles <command> [<args>]") 
		fmt.Println()
		fmt.Println("Commands:")
		os.Exit(1)
	}

	command := os.Args[1]
	if handler, ok := commands[command]; ok {
		handler()
	} else {
	fmt.Printf("heracles: '%s' is not a heracles command.\n", command)
}
}
