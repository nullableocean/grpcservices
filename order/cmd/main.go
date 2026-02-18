package main

import "log"

func main() {
	err := start()

	if err != nil {
		log.Fatalln(err)
	}
}
