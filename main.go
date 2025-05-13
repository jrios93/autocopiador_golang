package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

var running bool

func main() {
	// Inicializar la aplicación de Fyne
	myApp := app.New()
	myWindow := myApp.NewWindow("AutoCopiador de Series")
	myWindow.Resize(fyne.NewSize(500, 600))

	// Elementos de la interfaz
	seriesInput := widget.NewMultiLineEntry()
	seriesInput.SetPlaceHolder("Pega las series aquí separadas por espacio...")

	// Campo para la fecha
	dateEntry := widget.NewEntry()
	dateEntry.SetPlaceHolder("Ej: 13052025")

	statusLabel := widget.NewLabel("Esperando...")
	copiedCountLabel := widget.NewLabel("Copiadas: 0")
	var startButton *widget.Button
	startButton = widget.NewButton("Iniciar", func() {
		// Validar que la fecha no esté vacía
		if dateEntry.Text == "" {
			statusLabel.SetText("Por favor, ingresa una fecha.")
			return
		}

		if running {
			return
		}

		// Comenzar el proceso de copia
		running = true
		statusLabel.SetText("Ejecutando...")
		startButton.Disable()
		copiedCountLabel.SetText("Copiadas: 0")
		go startCopying(seriesInput.Text, dateEntry.Text, copiedCountLabel, statusLabel, startButton)
	})
	stopButton := widget.NewButton("Detener", func() {
		// Detener el proceso
		running = false
		statusLabel.SetText("Cancelado")
		startButton.Enable()
	})

	// Disponer los widgets
	content := container.NewVBox(
		seriesInput,
		dateEntry, // Agregar el campo de fecha
		startButton,
		stopButton,
		statusLabel,
		copiedCountLabel,
	)

	myWindow.SetContent(content)

	// Configurar ESC para detener el proceso de manera global
	go listenForEscapeKey()

	myWindow.ShowAndRun()
}

func listenForEscapeKey() {
	// Escuchar tecla ESC globalmente
	evChan := hook.Start()
	defer hook.End()

	for ev := range evChan {
		if ev.Kind == hook.KeyDown && ev.Keychar == 27 { // 27 es ESC
			if running {
				running = false
				fmt.Println("Cancelado con ESC")
			}
		}
	}
}

func startCopying(rawSeries, date string, copiedCountLabel, statusLabel *widget.Label, startButton *widget.Button) {
	// Espera antes de comenzar para cambiarse al programa
	time.Sleep(3 * time.Second)

	// Dividir las series en un slice
	series := strings.Fields(rawSeries)
	count := 0
	failedSeries := []string{} // Lista de series fallidas

	// Establecer el tiempo de retraso entre las acciones
	delay := 100 * time.Millisecond // Ajusta este valor para la velocidad

	// Empezar a copiar las series
	for _, s := range series {
		if !running {
			break
		}

		// Escribir la serie
		robotgo.TypeStr(s)
		time.Sleep(delay)

		// Bajar a la siguiente fila
		robotgo.KeyTap("tab")
		time.Sleep(delay)

		// Escribir la fecha
		robotgo.TypeStr(date)
		robotgo.KeyTap("tab")
		time.Sleep(delay)

		// Bajar a la siguiente línea
		robotgo.KeyTap("down")
		time.Sleep(delay)

		// Actualizar el contador de series copiadas
		count++
		copiedCountLabel.SetText(fmt.Sprintf("Copiadas: %d", count))

		// Actualizar la interfaz
		statusLabel.SetText(fmt.Sprintf("Copiando... (%d/%d)", count, len(series)))
	}

	// Guardar las series fallidas en un archivo
	if len(failedSeries) > 0 {
		saveFailedSeries(failedSeries)
	}

	if running {
		statusLabel.SetText("Finalizado")
	} else {
		statusLabel.SetText("Cancelado")
	}

	// Habilitar el botón "Iniciar" después de que termine el proceso
	startButton.Enable()
}

func saveFailedSeries(failedSeries []string) {
	file, err := os.Create("failed_series.txt")
	if err != nil {
		fmt.Println("Error al crear el archivo:", err)
		return
	}
	defer file.Close()

	for _, series := range failedSeries {
		_, err := file.WriteString(series + "\n")
		if err != nil {
			fmt.Println("Error al escribir en el archivo:", err)
			return
		}
	}

	fmt.Println("Series fallidas guardadas en 'failed_series.txt'")
}
