module github.com/SysdigDan/parasite-scanner

go 1.18

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/julienschmidt/httprouter v1.3.0
	github.com/prometheus/client_golang v1.12.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	tinygo.org/x/bluetooth v0.3.0
)

require (
	github.com/JuulLabs-OSS/cbgo v0.0.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/muka/go-bluetooth v0.0.0-20220319164423-1763af51ee1a // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace tinygo.org/x/bluetooth v0.3.0 => github.com/rbaron/bluetooth v0.3.1-0.20210501180115-a5ddbbc48845
