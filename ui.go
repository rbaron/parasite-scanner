package main

import (
	"container/ring"
	"fmt"
	"sort"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const kRingSize = 1000
const kHeaderHeight = 5

type DB map[string]*ring.Ring

type Widgets struct {
	header         *widgets.Paragraph
	help           *widgets.Paragraph
	soilMoisture   *widgets.Plot
	temp           *widgets.Plot
	humidity       *widgets.Plot
	batteryVoltage *widgets.Plot
	rssi           *widgets.Plot
	table          *widgets.Table
}

type TUI struct {
	dataChan         chan *ParasiteData
	seenKeys         []string
	selectedKeyIndex int
	db               *DB
	widgets          *Widgets
}

func InitUI() *TUI {
	if err := ui.Init(); err != nil {
		panic("Failed to initialize termui: " + err.Error())
	}

	tui := &TUI{}
	tui.dataChan = make(chan *ParasiteData)
	tui.seenKeys = []string{}
	tui.selectedKeyIndex = -1
	tui.db = &DB{}
	tui.widgets = initWidgets()
	return tui
}

func initWidgets() *Widgets {
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
	table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humidity", "Battery Voltage"}}
	table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorClear, ui.ModifierBold)

	return &Widgets{
		header:         header,
		help:           help,
		soilMoisture:   soilMoistureChart,
		temp:           tempChart,
		humidity:       humidityChart,
		rssi:           rssiChart,
		batteryVoltage: batteryChart,
		table:          table,
	}
}

func (tui *TUI) Close() {
	ui.Close()
}

func (tui *TUI) refreshData() {
	r := (*(tui.db))[tui.seenKeys[tui.selectedKeyIndex]]
	plotRecentData(tui.widgets.soilMoisture, r, func(data *ParasiteData) float64 { return float64(data.SoilMoisture) })
	plotRecentData(tui.widgets.temp, r, func(data *ParasiteData) float64 { return float64(data.TempCelcius) })
	plotRecentData(tui.widgets.humidity, r, func(data *ParasiteData) float64 { return float64(data.Humidity) })
	plotRecentData(tui.widgets.batteryVoltage, r, func(data *ParasiteData) float64 { return float64(data.BatteryVoltage) })
	plotRecentData(tui.widgets.rssi, r, func(data *ParasiteData) float64 { return -float64(data.RSSI) })

	table := tui.widgets.table
	table.Rows = [][]string{}
	table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humidity", "Battery Voltage", "RSSI", "Time"}}
	for i, k := range tui.seenKeys {
		var last = (*tui.db)[k].Prev().Value.(*ParasiteData)
		table.Rows = append(table.Rows, []string{
			last.Key,
			fmt.Sprintf("%5.1f%%", last.SoilMoisture),
			fmt.Sprintf("%5.1fC", last.TempCelcius),
			fmt.Sprintf("%5.1f%%", last.Humidity),
			fmt.Sprintf("%5.2fV", last.BatteryVoltage),
			fmt.Sprintf("%ddBm", last.RSSI),
			fmt.Sprintf("%.0fs ago", time.Since(last.Time).Seconds()),
		})
		if i == tui.selectedKeyIndex {
			// Skip the header.
			table.RowStyles[i+1] = ui.NewStyle(ui.ColorYellow)
		} else {
			// Skip the header.
			table.RowStyles[i+1] = ui.NewStyle(ui.ColorWhite)
		}
	}
}

func (tui *TUI) Run() {
	defer ui.Close()
	uiEvents := ui.PollEvents()
	for {
		select {
		case uiEvent := <-uiEvents:
			switch uiEvent.ID {
			case "q", "<C-c>":
				return
			case "j":
				if tui.selectedKeyIndex < len(tui.seenKeys)-1 {
					tui.selectedKeyIndex++
					tui.refreshData()
					tui.Render()
				}
			case "k":
				if tui.selectedKeyIndex > 0 {
					tui.selectedKeyIndex--
					tui.refreshData()
					tui.Render()
				}
			}
		case data := <-tui.dataChan:
			r, exists := (*tui.db)[data.Key]
			if !exists {
				r = ring.New(kRingSize)
			} else {
				// Avoid reprocessing repeated packets.
				if r.Prev().Value.(*ParasiteData).Counter == data.Counter {
					continue
				}
			}
			r.Value = data
			(*tui.db)[data.Key] = r.Next()
			if tui.selectedKeyIndex == -1 {
				tui.selectedKeyIndex = 0
			}
			exists = false
			for _, k := range tui.seenKeys {
				if data.Key == k {
					exists = true
				}
			}
			if !exists {
				tui.seenKeys = append(tui.seenKeys, data.Key)
			}
			tui.refreshData()
			tui.Render()
		}
	}

}

func (tui *TUI) Render() {
	widgets := tui.widgets
	ui.Render(widgets.header,
		widgets.help,
		widgets.soilMoisture,
		widgets.temp,
		widgets.humidity,
		widgets.batteryVoltage,
		widgets.rssi,
		widgets.table)
}

func plotRecentData(plot *widgets.Plot, r *ring.Ring, getter func(datapoint *ParasiteData) float64) {
	// Collect all valid points from ring.
	serie := []float64{}
	serieLen := 0
	for p := r.Next(); p != r; p = p.Next() {
		if p.Value != nil {
			data := p.Value.(*ParasiteData)
			serie = append(serie, getter(data))
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

func DumpDB(db *DB) {
	fmt.Println("Dumping DB:")
	keys := make([]string, 0)
	for k, _ := range *db {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		var last = (*db)[k].Prev().Value.(*ParasiteData)
		fmt.Println(last)
	}
	fmt.Println("-----------------------------")
}
