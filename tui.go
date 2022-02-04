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
	header            *widgets.Paragraph
	help              *widgets.Paragraph
	temp              *widgets.Plot
	humidity          *widgets.Plot
	batteryPercentage *widgets.Plot
	batteryVoltage    *widgets.Plot
	rssi              *widgets.Plot
	table             *widgets.Table
}

// TUI is a "Terminal User Interface".
// It implements the DataSubscriber interface, and will re-render whenever new data
// arrives, or when user input is detected.
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
	tui.Render()
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

	batteryPercentageChart := widgets.NewPlot()
	batteryPercentageChart.Title = "Battery Percentage (%)"
	batteryPercentageChart.Marker = widgets.MarkerDot
	batteryPercentageChart.SetRect(100, 24+kHeaderHeight, 200, 36+kHeaderHeight)
	batteryPercentageChart.DotMarkerRune = '+'
	batteryPercentageChart.LineColors[0] = ui.ColorRed

	batteryVoltageChart := widgets.NewPlot()
	batteryVoltageChart.Title = "Battery Voltage (V)"
	batteryVoltageChart.Marker = widgets.MarkerDot
	batteryVoltageChart.SetRect(100, 24+kHeaderHeight, 200, 36+kHeaderHeight)
	batteryVoltageChart.DotMarkerRune = '+'
	batteryVoltageChart.LineColors[0] = ui.ColorRed

	table := widgets.NewTable()
	table.Title = "Discovered b-parasites"
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.SetRect(0, 36+kHeaderHeight, 200, 60+kHeaderHeight)
	table.FillRow = true
	table.Rows = [][]string{{"UUID", "Soil Moisture", "Temperature", "Humidity", "Battery Voltage"}}
	table.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorClear, ui.ModifierBold)

	return &Widgets{
		header:            header,
		help:              help,
		temp:              tempChart,
		humidity:          humidityChart,
		rssi:              rssiChart,
		batteryPercentage: batteryPercentageChart,
		batteryVoltage:    batteryVoltageChart,
		table:             table,
	}
}

func (tui *TUI) Close() {
	ui.Close()
}

func (tui *TUI) refreshData() {
	r := (*(tui.db))[tui.seenKeys[tui.selectedKeyIndex]]
	plotRecentData(tui.widgets.temp, r, func(data *ParasiteData) float64 { return float64(data.TempCelcius) })
	plotRecentData(tui.widgets.humidity, r, func(data *ParasiteData) float64 { return float64(data.Humidity) })
	plotRecentData(tui.widgets.batteryPercentage, r, func(data *ParasiteData) float64 { return float64(data.BatteryPercentage) })
	plotRecentData(tui.widgets.batteryVoltage, r, func(data *ParasiteData) float64 { return float64(data.BatteryVoltage) })
	plotRecentData(tui.widgets.rssi, r, func(data *ParasiteData) float64 { return -float64(data.RSSI) })

	table := tui.widgets.table
	table.Rows = [][]string{}
	table.Rows = [][]string{{"UUID", "Temperature", "Humidity", "Battery Percentage", "Battery Voltage", "RSSI", "Time"}}
	for i, k := range tui.seenKeys {
		var last = (*tui.db)[k].Prev().Value.(*ParasiteData)
		table.Rows = append(table.Rows, []string{
			last.Key,
			fmt.Sprintf("%5.1fC", last.TempCelcius),
			fmt.Sprintf("%5.1f%%", last.Humidity),
			fmt.Sprintf("%5.2fV", last.BatteryPercentage),
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

func (tui *TUI) Ingest(data *ParasiteData) {
	tui.dataChan <- data
}

func (tui *TUI) Render() {
	widgets := tui.widgets
	ui.Render(widgets.header,
		widgets.help,
		widgets.temp,
		widgets.humidity,
		widgets.batteryPercentage,
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
	logger.Println("Dumping DB:")
	keys := make([]string, 0)
	for k := range *db {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		var last = (*db)[k].Prev().Value.(*ParasiteData)
		logger.Println(last)
	}
	logger.Println("-----------------------------")
}
