package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/colemickens/azkube/cmd"
)

const (
	ClientID = "azkube-client-id"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		log.Fatalln(err)
	}
}
