package main

import (
	"log"

	"cryptocore/internal/client/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		log.Println(err)
	}
}
