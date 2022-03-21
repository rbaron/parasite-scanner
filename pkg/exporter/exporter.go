package exporter

import (
	"errors"
	"net/http"
	"sync"
	"time"
	"log"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	cfg "github.com/SysdigDan/parasite-scanner/pkg/config"
	"github.com/SysdigDan/parasite-scanner/pkg/data"
)

type PrometheusExporter struct {
	exporter *http.Server
	promData chan *data.ParasiteData
	config   *cfg.PROMConfig
	device   *cfg.DEVICEConfig
}

var (
	tempGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "thermometer",
			Name:      "temperature_celsius",
			Help:      "Temperature in Celsius.",
		},
		[]string{
			"sensor",
			"name",
			"mac",
		},
	)
	humGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "thermometer",
			Name:      "humidity_ratio",
			Help:      "Humidity in percent.",
		},
		[]string{
			"sensor",
			"name",
			"mac",
		},
	)
	battGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "thermometer",
			Name:      "battery_ratio",
			Help:      "Battery in percent.",
		},
		[]string{
			"sensor",
			"name",
			"mac",
		},
	)
	voltGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "thermometer",
			Name:      "battery_volts",
			Help:      "Battery in Volt.",
		},
		[]string{
			"sensor",
			"name",
			"mac",
		},
	)
	frameGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "thermometer",
			Name:      "frame_current",
			Help:      "Current frame number.",
		},
		[]string{
			"sensor",
			"name",
			"mac",
		},
	)
	rssiGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "thermometer",
			Name:      "rssi_dbm",
			Help:      "Received Signal Strength Indication.",
		},
		[]string{
			"sensor",
			"name",
			"mac",
		},
	)
)

const Sensor = "LYWSD03MMC"
const ExpiryAtc = 2.5 * 10 * time.Second
const ExpiryStock = 2.5 * 10 * time.Minute
const ExpiryConn = 2.5 * 10 * time.Second

var expirers = make(map[string]*time.Timer)
var expirersLock sync.Mutex

func bump(name, mac string, expiry time.Duration) {
	expirersLock.Lock()
	if t, ok := expirers[mac]; ok {
		t.Reset(expiry)
	} else {
		expirers[mac] = time.AfterFunc(expiry, func() {
			log.Println("[exporter] Device expiring:", name, mac)
			tempGauge.DeleteLabelValues(Sensor, name, mac)
			humGauge.DeleteLabelValues(Sensor, name, mac)
			battGauge.DeleteLabelValues(Sensor, name, mac)
			voltGauge.DeleteLabelValues(Sensor, name, mac)
			frameGauge.DeleteLabelValues(Sensor, name, mac)
			rssiGauge.DeleteLabelValues(Sensor, name, mac)

			expirersLock.Lock()
			delete(expirers, mac)
			expirersLock.Unlock()
		})
	}
	expirersLock.Unlock()
}

func logTemperature(name, mac string, temp float64) {
	tempGauge.WithLabelValues(Sensor, name, mac).Set(temp)
	log.Printf("[exporter] %s thermometer_temperature_celsius %.1f\n", name, temp)
}

func logHumidity(name, mac string, hum float64) {
	humGauge.WithLabelValues(Sensor, name, mac).Set(hum)
	log.Printf("[exporter] %s thermometer_humidity_ratio %.0f\n", name, hum)
}

func logVoltage(name, mac string, batv float64) {
	voltGauge.WithLabelValues(Sensor, name, mac).Set(batv)
	log.Printf("[exporter] %s thermometer_battery_volts %.3f\n", name, batv)
}

func logRSSI(name, mac string, rssi float64) {
	rssiGauge.WithLabelValues(Sensor, name, mac).Set(rssi)
	log.Printf("[exporter] %s thermometer_rssi_dbm %.3f\n", name, rssi)
}

func logBatteryPercent(name, mac string, batp float64) {
	battGauge.WithLabelValues(Sensor, name, mac).Set(batp)
	log.Printf("[exporter] %s thermometer_battery_ratio %.0f\n", name, batp)
}

func LogStockCharacteristic(name, mac string, temp float64, hum float64, batv float64) {
	bump(name, mac, ExpiryConn)
	logTemperature(name, mac, temp)
	logHumidity(name, mac, hum)
	logVoltage(name, mac, batv)
}

func LogAtcTemp(name, mac string, temp float64) {
	bump(name, mac, ExpiryConn)
	logTemperature(name, mac, temp)
}

func LogAtcHumidity(name, mac string, hum float64) {
	bump(name, mac, ExpiryConn)
	logHumidity(name, mac, hum)
}

func LogAtcBattery(name, mac string, batp float64) {
	bump(name, mac, ExpiryConn)
	logBatteryPercent(name, mac, batp)
}

func metrics(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h.ServeHTTP(w, r)
	}
}

func exporterRouter() *httprouter.Router {
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>parasite-exporter</title></head><body><h1>parasite-exporter</h1><p><a href="/metrics">Metrics</a></p></body></html>`))
	})
	router.GET("/metrics", metrics(promhttp.Handler()))
	return router
}

func MakePROMExporter(device *cfg.DEVICEConfig, cfg *cfg.PROMConfig) *PrometheusExporter {
	return &PrometheusExporter{
		exporter: &http.Server{
			Addr:    cfg.Host,
			Handler: exporterRouter(),
		},
		promData: make(chan *data.ParasiteData),
		config:   cfg,
		device:   device,
	}
}

func (exporter *PrometheusExporter) LogPrometheusData(deviceConfig *cfg.ParasiteConfig, data *data.ParasiteData) {
	bump(deviceConfig.Name, data.Key, ExpiryAtc)
	logTemperature(deviceConfig.Name, data.Key, data.TempCelcius)
	logHumidity(deviceConfig.Name, data.Key, data.Humidity)
	logBatteryPercent(deviceConfig.Name, data.Key, data.BatteryPercentage)
	logVoltage(deviceConfig.Name, data.Key, data.BatteryVoltage)
	logRSSI(deviceConfig.Name, data.Key, float64(data.RSSI))
}

func (exporter *PrometheusExporter) Ingest(data *data.ParasiteData) {
	exporter.promData <- data
}

func (exporter *PrometheusExporter) Start() error {
	log.Println("[exporter] Starting Prometheus exporter...")

	if len(exporter.exporter.Addr) == 0 {
		return errors.New("Exporter missing address")
	}

	if exporter.exporter.Handler == nil {
		return errors.New("Exporter missing handler")
	}

	return exporter.exporter.ListenAndServe()
}

func (exporter *PrometheusExporter) Run() {
	go func() {
		if err := exporter.Start(); err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("[exporter] Prometheus exporter listening on %s\n", exporter.config.Host)
	}()

	for data := range exporter.promData {
		deviceConfig, exists := exporter.device.Registry[cfg.MACAddr(data.Key)]
		if !exists {
			log.Printf("[mqtt] Received valid BLE broadcast from %s, but it's not configured for MQTT\n", data.Key)
			continue
		}
		exporter.LogPrometheusData(deviceConfig, data)
	}
}
