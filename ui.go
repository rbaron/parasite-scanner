package main

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func UIRun(inChan chan string, db *DB) {
	if err := ui.Init(); err != nil {
		panic("Failed to initialize termui: " + err.Error())
	}
	defer ui.Close()

	p1 := widgets.NewPlot()
	p1.Title = "dot-mode line Chart"
	p1.Marker = widgets.MarkerDot
	p1.SetRect(0, 0, 200, 50)
	p1.DotMarkerRune = '+'
	p1.AxesColor = ui.ColorWhite
	p1.LineColors[0] = ui.ColorRed
	p1.DrawDirection = widgets.DrawRight
	p1.MaxVal = 100.

	ui.Render(p1)

	refreshData := func(key string) {
		series := []float64{}
		r := (*db)[key]
		for p := r.Next(); p != r; p = p.Next() {
			if p.Value != nil {
				series = append(series, float64(p.Value.(ParasiteData).SoilMoisture))
			}
		}
		p1.Data = [][]float64{series}
		ui.Render(p1)
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
