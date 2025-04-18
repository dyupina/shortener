package main

import (
	"os"
)

func main() {
	os.Exit(0) // want "osexitcheck os.Exit cannot be called in main function of main package"
}
