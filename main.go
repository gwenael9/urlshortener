package main

import (
	cmd2 "github.com/axellelanca/urlshortener/cmd"
	_ "github.com/axellelanca/urlshortener/cmd/cli"
	_ "github.com/axellelanca/urlshortener/cmd/server"
)

func main() {
	cmd2.Execute()
}
