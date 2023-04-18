package main

import (
	"Project/controllers"
	"Project/models"
	"encoding/json"
	"github.com/gen2brain/dlgs"
	"log"
	"os"
	"os/exec"
)

func main() {
	var grammar = openJson()
	controller := controllers.NewGrammarController()
	controller.SetGrammar(grammar)
	exportToJSON(controller.GetGrammar())
}

func openJson() models.Grammar {
	// Open window to select file
	filename, _, err := dlgs.File("Choose a file", "", false)
	if err != nil {
		log.Fatal(err)
	}

	// Open selected file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var grammar models.Grammar
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&grammar)
	if err != nil {
		log.Fatal(err)
	}
	return grammar
}

func exportToJSON(grammar models.Grammar) error {
	fileName, _, err := dlgs.File("Save JSON file", "", true)
	if err != nil {
		return err
	}
	file, err := os.Create(fileName + "/Result.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(grammar)
	var command = exec.Command("cmd", "/C", "start", "msedge", fileName+"/Result.json")
	err = command.Start()
	if err != nil {
		return err
	}
	return nil
}
