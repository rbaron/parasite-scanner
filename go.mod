module github.com/rbaron/parasite-scanner

go 1.16

require (
	github.com/gizak/termui/v3 v3.1.0
	tinygo.org/x/bluetooth v0.3.0
)

replace tinygo.org/x/bluetooth v0.3.0 => github.com/rbaron/bluetooth v0.3.1-0.20210425094756-e6fb45ea1175
