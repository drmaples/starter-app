package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/drmaples/starter-app/app/controller"
)

func main() {
	if len(os.Args) != 2 {
		panic("missing email address")
	}
	email := os.Args[1]

	token, err := controller.NewSignedToken(email, "", "", time.Hour)
	if err != nil {
		panic(err)
	}

	fmt.Printf("JWT:\n%[1]s\n%[2]s\n%[1]s\n", strings.Repeat("-", 40), token)
}
