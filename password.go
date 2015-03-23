package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
)

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
