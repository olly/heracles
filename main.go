package main

import "fmt"
import "github.com/mgutz/ansi"
import "github.com/howeyc/gopass"

func main() {
    fmt.Printf("Password: ")
    passwordInput := gopass.GetPasswdMasked()
    password := string(passwordInput)
    fmt.Println(ansi.Color(password, "red"))
}

