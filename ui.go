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

	p1 := widgets.NewPlot()
	p1.Title = "Soil Moisture"
	p1.Marker = widgets.MarkerDot
	p1.SetRect(0, 0, 200, 40)
	p1.DotMarkerRune = '+'
	p1.AxesColor = ui.ColorWhite
	p1.LineColors[0] = ui.ColorRed
	p1.DrawDirection = widgets.DrawRight
	p1.MaxVal = 100.

	// list := widgets.NewList()
	// list.Title = "b-parasites"
	// // list.Rows = []string{
	// // 	"Item 1",
	// // 	"Item 2",
	// // 	"Item 3",
	// // }
	// list.TextStyle = ui.NewStyle(ui.ColorWhite)
	// list.SelectedRowStyle = ui.NewStyle(ui.ColorBlue)
	// list.WrapText = true
	// list.SetRect(200, 0, 250, 50)

	table := widgets.NewTable()
	// table.Rows = [][]string{
	// 	[]string{"header1", "header2", "header3"},
	// 	[]string{"AAA", "BBB", "CCC"},
	// 	[]string{"DDD", "EEE", "FFF"},
	// 	[]string{"GGG", "HHH", "III"},
	// }
	table.Title = "b-parasites"
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	// table.BorderStyle = ui.NewStyle(ui.ColorGreen)
	table.SetRect(0, 40, 200, 60)
	table.FillRow = true
	table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humitiy", "Battery Voltage"}}
	// table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorClear, ui.ModifierBold)
	// table.RowStyles[2] = ui.NewStyle(ui.ColorWhite, ui.ColorRed, ui.ModifierBold)
	// table.RowStyles[3] = ui.NewStyle(ui.ColorYellow)

	ui.Render(p1, table)

	refreshData := func(key string) {
		series := []float64{}
		r := (*db)[key]
		for p := r.Next(); p != r; p = p.Next() {
			if p.Value != nil {
				series = append(series, float64(p.Value.(ParasiteData).SoilMoisture))
			}
		}
		p1.Data = [][]float64{series}

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

		ui.Render(p1, table)
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
