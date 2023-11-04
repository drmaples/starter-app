package main

import (
	"fmt"
	"os"
	"time"

	"github.com/drmaples/starter-app/app/controller"
	"github.com/drmaples/starter-app/app/platform"
)

func main() {
	if len(os.Args) != 2 {
		panic("missing email address")
	}
	email := os.Args[1]

	cfg, err := platform.NewConfig()
	if err != nil {
		panic(err)
	}
	token, err := controller.NewSignedToken(cfg.JWTSignKey, email, "", "", time.Hour)
	if err != nil {
		panic(err)
	}

	fmt.Printf("JWT:\n%s\n", token)
}
