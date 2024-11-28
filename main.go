package main

import (
	"github.com/biosecret/go-iot/app"
	_ "github.com/biosecret/go-iot/docs"
)

func main() {
	// setup and run app
	err := app.SetupAndRunApp()
	if err != nil {
		panic(err)
	}
}
