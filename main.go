// Proyecto Analizador de Gramáticas LL(1)
// Autor: JhonierSerna14
// Este archivo contiene el programa principal y la interfaz de usuario.
package main

import (
	"Project/controllers"
	"Project/models"
	"encoding/json"
	"os"
	"os/exec"

	"github.com/gen2brain/dlgs"
)

// main es el punto de entrada del programa.
// Muestra una ventana de bienvenida, solicita la gramática, realiza el análisis y exporta el resultado.
func main() {
	dlgs.Info("Bienvenido", "Analizador de gramáticas LL(1) - Proyecto Lenguajes")
	var grammar models.Grammar
	var err error

	// Abrir archivo de gramática
	grammar = openJson()

	// Crear controlador y procesar la gramática
	controller := controllers.NewGrammarController()
	controller.SetGrammar(grammar)

	// Exportar resultado y mostrar mensaje
	err = exportToJSON(controller.GetGrammar())
	if err != nil {
		dlgs.Error("Error", "No se pudo exportar el resultado: "+err.Error())
	} else {
		dlgs.Info("Éxito", "El análisis se completó y el resultado fue guardado correctamente.")
	}
}

// openJson abre una ventana para seleccionar el archivo de gramática en formato JSON.
// Devuelve la estructura Grammar cargada o vacía en caso de error.
func openJson() models.Grammar {
	filename, _, err := dlgs.File("Selecciona el archivo de gramática (JSON)", "", false)
	if err != nil {
		dlgs.Error("Error", "No se pudo abrir el selector de archivos: "+err.Error())
		return models.Grammar{}
	}

	// Abrir el archivo seleccionado
	file, err := os.Open(filename)
	if err != nil {
		dlgs.Error("Error", "No se pudo abrir el archivo seleccionado: "+err.Error())
		return models.Grammar{}
	}
	defer file.Close()

	// Decodificar el JSON en la estructura Grammar
	var grammar models.Grammar
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&grammar)
	if err != nil {
		dlgs.Error("Error", "No se pudo decodificar el archivo JSON: "+err.Error())
		return models.Grammar{}
	}
	return grammar
}

// exportToJSON guarda la gramática procesada en un archivo JSON en la carpeta seleccionada.
// Además, intenta abrir el archivo generado en el navegador.
// Devuelve error si ocurre algún problema en el proceso.
func exportToJSON(grammar models.Grammar) error {
	fileName, _, err := dlgs.File("Selecciona la carpeta para guardar el resultado", "", true)
	if err != nil {
		return err
	}

	// Crear archivo de resultado
	file, err := os.Create(fileName + "/Result.json")
	if err != nil {
		return err
	}
	defer file.Close()

	// Codificar la gramática en formato JSON
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(grammar)
	if err != nil {
		return err
	}

	// Intentar abrir el archivo en el navegador (no es crítico si falla)
	var command = exec.Command("cmd", "/C", "start", "msedge", fileName+"/Result.json")
	_ = command.Start()
	return nil
}
