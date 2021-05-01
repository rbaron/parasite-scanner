module github.com/rbaron/parasite-scanner

go 1.16

require (
	github.com/eclipse/paho.mqtt.golang v1.3.3
	github.com/gizak/termui/v3 v3.1.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	tinygo.org/x/bluetooth v0.3.0
)

replace tinygo.org/x/bluetooth v0.3.0 => github.com/rbaron/bluetooth v0.3.1-0.20210501180115-a5ddbbc48845
