package main

import (
	"container/ring"
	"fmt"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var keys = []string{}
var selectedKey int = -1

func plotRecentData(plot *widgets.Plot, r *ring.Ring, getter func(datapoint *ParasiteData) float64) {
	// Collect all valid points from ring.
	serie := []float64{}
	serieLen := 0
	for p := r.Next(); p != r; p = p.Next() {
		if p.Value != nil {
			data := p.Value.(ParasiteData)
			serie = append(serie, getter(&data))
			serieLen++
		}
	}

	// How many points will fit?
	// TODO: Is 5 always a good safe zone?
	nPoints := plot.Inner.Size().X - 5
	if nPoints >= serieLen {
		nPoints = serieLen
	}

	serie = serie[(serieLen - nPoints):serieLen]
	plot.Data = [][]float64{serie}
}

const kHeaderHeight = 5

func UIRun(inChan chan string, db *DB) {
	if err := ui.Init(); err != nil {
		panic("Failed to initialize termui: " + err.Error())
	}
	defer ui.Close()

	header := widgets.NewParagraph()
	header.Title = "b-parasite scanner"
	header.Text = "\nListening to BLE advertisements and looking for b-parasites..."
	header.SetRect(0, 0, 100, kHeaderHeight)

	help := widgets.NewParagraph()
	help.Title = "Controls"
	help.Text = "j: next\nk: previous"
	help.SetRect(100, 0, 200, kHeaderHeight)

	soilMoistureChart := widgets.NewPlot()
	soilMoistureChart.Title = "Soil Moisture (%)"
	soilMoistureChart.Marker = widgets.MarkerDot
	soilMoistureChart.SetRect(0, 0+kHeaderHeight, 100, 18+kHeaderHeight)
	soilMoistureChart.DotMarkerRune = '+'
	soilMoistureChart.LineColors[0] = ui.ColorBlue
	soilMoistureChart.MaxVal = 100.

	tempChart := widgets.NewPlot()
	tempChart.Title = "Temperature (C)"
	tempChart.Marker = widgets.MarkerDot
	tempChart.SetRect(0, 18+kHeaderHeight, 100, 36+kHeaderHeight)
	tempChart.DotMarkerRune = '+'
	tempChart.LineColors[0] = ui.ColorYellow

	humidityChart := widgets.NewPlot()
	humidityChart.Title = "Humidity (%)"
	humidityChart.Marker = widgets.MarkerDot
	humidityChart.SetRect(100, 0+kHeaderHeight, 200, 12+kHeaderHeight)
	humidityChart.DotMarkerRune = '+'
	humidityChart.LineColors[0] = ui.ColorGreen

	rssiChart := widgets.NewPlot()
	rssiChart.Title = "RSSI (-dBm)"
	rssiChart.Marker = widgets.MarkerDot
	rssiChart.SetRect(100, 12+kHeaderHeight, 200, 24+kHeaderHeight)
	rssiChart.DotMarkerRune = '+'
	rssiChart.LineColors[0] = ui.ColorWhite

	batteryChart := widgets.NewPlot()
	batteryChart.Title = "Battery Voltage (V)"
	batteryChart.Marker = widgets.MarkerDot
	batteryChart.SetRect(100, 24+kHeaderHeight, 200, 36+kHeaderHeight)
	batteryChart.DotMarkerRune = '+'
	batteryChart.LineColors[0] = ui.ColorRed

	table := widgets.NewTable()
	table.Title = "Discovered b-parasites"
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.SetRect(0, 36+kHeaderHeight, 200, 60+kHeaderHeight)
	table.FillRow = true
	table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humitiy", "Battery Voltage"}}
	table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorClear, ui.ModifierBold)

	render := func() {
		ui.Render(header, help, soilMoistureChart, tempChart, humidityChart, batteryChart, rssiChart, table)
	}
	render()

	refreshData := func() {
		r := (*db)[keys[selectedKey]]
		plotRecentData(soilMoistureChart, r, func(data *ParasiteData) float64 { return float64(data.SoilMoisture) })
		plotRecentData(tempChart, r, func(data *ParasiteData) float64 { return float64(data.TempCelcius) })
		plotRecentData(humidityChart, r, func(data *ParasiteData) float64 { return float64(data.Humidity) })
		plotRecentData(batteryChart, r, func(data *ParasiteData) float64 { return float64(data.BatteryVoltage) })
		plotRecentData(rssiChart, r, func(data *ParasiteData) float64 { return -float64(data.RSSI) })

		table.Rows = [][]string{}
		table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humidity", "Battery Voltage", "RSSI", "Time"}}
		for i, k := range keys {
			var last = (*db)[k].Prev().Value.(ParasiteData)
			table.Rows = append(table.Rows, []string{
				strings.Split(last.Key, "-")[4],
				fmt.Sprintf("%5.1f%%", last.SoilMoisture),
				fmt.Sprintf("%5.1fC", last.TempCelcius),
				fmt.Sprintf("%5.1f%%", last.Humidity),
				fmt.Sprintf("%5.2fV", last.BatteryVoltage),
				fmt.Sprintf("%ddBm", last.RSSI),
				fmt.Sprintf("%.0fs ago", time.Since(last.Time).Seconds()),
			})
			if i == selectedKey {
				// Skip the header.
				table.RowStyles[i+1] = ui.NewStyle(ui.ColorYellow)
			} else {
				// Skip the header.
				table.RowStyles[i+1] = ui.NewStyle(ui.ColorWhite)
			}
		}
		render()
	}

	uiEvents := ui.PollEvents()
	for {
		select {
		case uiEvent := <-uiEvents:
			switch uiEvent.ID {
			case "q", "<C-c>":
				return
			case "j":
				if selectedKey < len(keys)-1 {
					selectedKey++
					refreshData()
				}
			case "k":
				if selectedKey > 0 {
					selectedKey--
					refreshData()
				}
			}
		case dbEvent := <-inChan:
			if selectedKey == -1 {
				selectedKey = 0
			}
			exists := false
			for _, k := range keys {
				if dbEvent == k {
					exists = true
				}
			}
			if !exists {
				keys = append(keys, dbEvent)
			}
			refreshData()
		}
	}
}
