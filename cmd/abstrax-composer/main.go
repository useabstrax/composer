package main

import (
	"os"

	"github.com/useabstrax/abstrax/plugins/composer/internal/commands"
)

func main() {
	os.Exit(commands.Execute())
}
