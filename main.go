package main

import (
	"fmt"
	"sort"

	"container/ring"
)

const kRingSize = 1000

type DB map[string]*ring.Ring

var db = DB{}

func DumpDB() {
	keys := make([]string, 0)
	for k, _ := range db {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		var last = db[k].Prev().Value.(ParasiteData)
		fmt.Println(last)
	}
	fmt.Println("-----------------------------")
}

func main() {
	fmt.Println("Waiting for parasites' BLE advertisements...")
	ch := make(chan ParasiteData)
	go StartScanning(ch)

	ui_chan := make(chan string)
	go UIRun(ui_chan, &db)

	for data := range ch {
		r, exists := db[data.Key]
		if !exists {
			r = ring.New(kRingSize)
		} else {
			// Avoid reprocessing repeated packets.
			if r.Value.(ParasiteData).Counter == data.Counter {
				continue
			}
		}
		r.Value = data
		db[data.Key] = r.Next()
		ui_chan <- data.Key
	}
}
