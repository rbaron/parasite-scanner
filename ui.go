package main

import (
	"fmt"
	"sort"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func UIRun(inChan chan string, db *DB) {
	if err := ui.Init(); err != nil {
		panic("Failed to initialize termui: " + err.Error())
	}
	defer ui.Close()

	soilMoistureChart := widgets.NewPlot()
	soilMoistureChart.Title = "Soil Moisture (%)"
	soilMoistureChart.Marker = widgets.MarkerDot
	soilMoistureChart.SetRect(0, 0, 150, 40)
	soilMoistureChart.DotMarkerRune = '+'
	soilMoistureChart.LineColors[0] = ui.ColorBlue
	soilMoistureChart.MaxVal = 100.

	tempChart := widgets.NewPlot()
	tempChart.Title = "Temperature (C)"
	tempChart.Marker = widgets.MarkerDot
	tempChart.SetRect(150, 0, 200, 10)
	tempChart.DotMarkerRune = '+'
	tempChart.LineColors[0] = ui.ColorYellow

	humidityChart := widgets.NewPlot()
	humidityChart.Title = "Humidity (%)"
	humidityChart.Marker = widgets.MarkerDot
	humidityChart.SetRect(150, 10, 200, 20)
	humidityChart.DotMarkerRune = '+'
	humidityChart.LineColors[0] = ui.ColorGreen

	batteryChart := widgets.NewPlot()
	batteryChart.Title = "Battery Voltage (V)"
	batteryChart.Marker = widgets.MarkerDot
	batteryChart.SetRect(150, 20, 200, 30)
	batteryChart.DotMarkerRune = '+'
	batteryChart.LineColors[0] = ui.ColorRed

	table := widgets.NewTable()
	table.Title = "Discovered b-parasites"
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.SetRect(0, 40, 200, 60)
	table.FillRow = true
	table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humitiy", "Battery Voltage"}}
	table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorClear, ui.ModifierBold)

	render := func() {
		ui.Render(soilMoistureChart, tempChart, humidityChart, batteryChart, table)
	}
	render()

	refreshData := func(key string) {
		soilMoistureSeries := []float64{}
		tempSeries := []float64{}
		humiditySeries := []float64{}
		battySeries := []float64{}
		r := (*db)[key]
		for p := r.Next(); p != r; p = p.Next() {
			if p.Value != nil {
				soilMoistureSeries = append(soilMoistureSeries, float64(p.Value.(ParasiteData).SoilMoisture))
				tempSeries = append(tempSeries, float64(p.Value.(ParasiteData).TempCelcius))
				humiditySeries = append(humiditySeries, float64(p.Value.(ParasiteData).Humidity))
				battySeries = append(battySeries, float64(p.Value.(ParasiteData).BatteryVoltage))
			}
		}
		soilMoistureChart.Data = [][]float64{soilMoistureSeries}
		tempChart.Data = [][]float64{tempSeries}
		humidityChart.Data = [][]float64{humiditySeries}
		batteryChart.Data = [][]float64{battySeries}

		table.Rows = [][]string{}
		keys := make([]string, 0)
		for k, _ := range *db {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humitiy", "Battery Voltage"}}
		for _, k := range keys {
			var last = (*db)[k].Prev().Value.(ParasiteData)
			table.Rows = append(table.Rows, []string{last.Key,
				fmt.Sprintf("%5.1f%%", last.SoilMoisture),
				fmt.Sprintf("%4.1fC", last.TempCelcius),
				fmt.Sprintf("%5.1f%%", last.Humidity),
				fmt.Sprintf("%3.1fV", last.BatteryVoltage),
			})
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
			}
		case dbEvent := <-inChan:
			_ = dbEvent
			refreshData(dbEvent)
		}
	}
}
