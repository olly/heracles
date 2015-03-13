package main

import "fmt"
import "io/ioutil"
import "os"
import "os/user"
import "strconv"
import "github.com/mgutz/ansi"
import "github.com/howeyc/gopass"

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

func main() {
	fmt.Printf("Password: ")
	password := Password{gopass.GetPasswdMasked()}
	fmt.Println(ansi.Color(password.String(), "red"))

	_ = password.WithTempFile(func(path string) {
		fmt.Println(path)
	})

	password.Clear()
	fmt.Printf("%v\n", password)
}
