package main

import (
	"fmt"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var keys = []string{}
var selectedKey int = -1

func UIRun(inChan chan string, db *DB) {
	if err := ui.Init(); err != nil {
		panic("Failed to initialize termui: " + err.Error())
	}
	defer ui.Close()

	soilMoistureChart := widgets.NewPlot()
	soilMoistureChart.Title = "Soil Moisture (%)"
	soilMoistureChart.Marker = widgets.MarkerDot
	soilMoistureChart.SetRect(0, 0, 100, 18)
	soilMoistureChart.DotMarkerRune = '+'
	soilMoistureChart.LineColors[0] = ui.ColorBlue
	soilMoistureChart.MaxVal = 100.

	tempChart := widgets.NewPlot()
	tempChart.Title = "Temperature (C)"
	tempChart.Marker = widgets.MarkerDot
	tempChart.SetRect(0, 18, 100, 36)
	tempChart.DotMarkerRune = '+'
	tempChart.LineColors[0] = ui.ColorYellow

	humidityChart := widgets.NewPlot()
	humidityChart.Title = "Humidity (%)"
	humidityChart.Marker = widgets.MarkerDot
	humidityChart.SetRect(100, 0, 200, 12)
	humidityChart.DotMarkerRune = '+'
	humidityChart.LineColors[0] = ui.ColorGreen

	rssiChart := widgets.NewPlot()
	rssiChart.Title = "RSSI (-dBm)"
	rssiChart.Marker = widgets.MarkerDot
	rssiChart.SetRect(100, 12, 200, 24)
	rssiChart.DotMarkerRune = '+'
	rssiChart.LineColors[0] = ui.ColorWhite

	batteryChart := widgets.NewPlot()
	batteryChart.Title = "Battery Voltage (V)"
	batteryChart.Marker = widgets.MarkerDot
	batteryChart.SetRect(100, 24, 200, 36)
	batteryChart.DotMarkerRune = '+'
	batteryChart.LineColors[0] = ui.ColorRed

	table := widgets.NewTable()
	table.Title = "Discovered b-parasites"
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.SetRect(0, 36, 200, 60)
	table.FillRow = true
	table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humitiy", "Battery Voltage"}}
	table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorClear, ui.ModifierBold)

	render := func() {
		ui.Render(soilMoistureChart, tempChart, humidityChart, batteryChart, rssiChart, table)
	}
	render()

	refreshData := func() {
		soilMoistureSeries := []float64{}
		tempSeries := []float64{}
		humiditySeries := []float64{}
		battySeries := []float64{}
		rssiSeries := []float64{}
		r := (*db)[keys[selectedKey]]
		for p := r.Next(); p != r; p = p.Next() {
			if p.Value != nil {
				soilMoistureSeries = append(soilMoistureSeries, float64(p.Value.(ParasiteData).SoilMoisture))
				tempSeries = append(tempSeries, float64(p.Value.(ParasiteData).TempCelcius))
				humiditySeries = append(humiditySeries, float64(p.Value.(ParasiteData).Humidity))
				battySeries = append(battySeries, float64(p.Value.(ParasiteData).BatteryVoltage))
				rssiSeries = append(rssiSeries, -1*float64(p.Value.(ParasiteData).RSSI))
			}
		}
		soilMoistureChart.Data = [][]float64{soilMoistureSeries}
		tempChart.Data = [][]float64{tempSeries}
		humidityChart.Data = [][]float64{humiditySeries}
		batteryChart.Data = [][]float64{battySeries}
		rssiChart.Data = [][]float64{rssiSeries}

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
