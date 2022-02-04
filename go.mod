module github.com/rbaron/parasite-scanner

go 1.16

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/gizak/termui/v3 v3.1.0
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/julienschmidt/httprouter v1.3.0
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/prometheus/client_golang v1.12.1
	github.com/sirupsen/logrus v1.8.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	tinygo.org/x/bluetooth v0.3.0
)

replace tinygo.org/x/bluetooth v0.3.0 => github.com/rbaron/bluetooth v0.3.1-0.20210501180115-a5ddbbc48845
