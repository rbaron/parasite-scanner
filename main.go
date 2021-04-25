package main

func main() {
	ui := InitUI()
	ch := make(chan *ParasiteData)
	go StartScanning(ch)

	go ui.Run()

	for data := range ch {
		ui.dataChan <- data
	}
}
