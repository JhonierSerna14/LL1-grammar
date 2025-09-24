# Analizador de Gramáticas LL(1)

Este proyecto es una aplicación en Go para analizar gramáticas y determinar si son LL(1). Permite cargar una gramática en formato JSON, realiza el análisis (eliminación de recursión, factorización, cálculo de conjuntos First, Follow y Predicción) y exporta el resultado en un archivo JSON.

## Características
- Interfaz gráfica sencilla (ventanas de selección y mensajes informativos).
- Carga de gramáticas desde archivos JSON.
- Procesamiento automático: eliminación de recursión y factorización.
- Cálculo de conjuntos First, Follow y Predicción.
- Verificación si la gramática es LL(1).
- Exportación del resultado en JSON y apertura automática en el navegador.

## Estructura del Proyecto
- `main.go`: Programa principal y gestión de la interfaz.
- `controllers/GrammarController.go`: Lógica de análisis y procesamiento de la gramática.
- `models/Grammar.go`: Estructura de datos para la gramática.
- `models/Production.go`: Estructura de datos para las producciones.
- `examples/`: Ejemplos de gramáticas en formato JSON.
- `versions/`: Versiones ejecutables del proyecto.

## Uso
1. Ejecuta el programa principal (`main.go`).
2. Selecciona el archivo JSON de la gramática a analizar.
3. El programa realiza el análisis y muestra mensajes informativos.
4. Selecciona la carpeta donde se guardará el resultado.
5. El resultado se guarda como `Result.json` y se abre automáticamente en el navegador.

## Formato de Gramática (JSON)
```json
{
  "initial": "S",
  "terminals": ["a", "b"],
  "nonTerminals": ["S", "A"],
  "productions": [
    { "left": "S", "right": ["A", "a"] },
    { "left": "A", "right": ["b"] }
  ]
}
```

## Requisitos
- Go 1.18+
- [gen2brain/dlgs](https://github.com/gen2brain/dlgs) para ventanas de diálogo.

##
⚡ Desarrollado con Go y ❤️

🌟 ¡Dale una estrella si te gusta el proyecto! ⭐
