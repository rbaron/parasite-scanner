# Fatal Error on macos BigSur
[Solution](https://github.com/tinygo-org/bluetooth/issues/48) is to add `iterm` to "System Preferences" —> "Security & Privacy" —> "Privacy" -> "Bluetooth".

# How to get advertisement data?
- tinygo-org/bluetooth implementation [doesn't use advFields](https://github.com/tinygo-org/bluetooth/blob/41f73176384665d9351a498e908a02cc64efee6d/adapter_darwin.go#L90)

# Home Assistant
- [Available device classes](https://www.home-assistant.io/integrations/sensor#device-class) for the `sensor` component

# Replacing tinygo-org/bluetooth with a custom fork
Assuming you want to replace the original module with the tip of the `release` branch of your fork:
```
replace tinygo.org/x/bluetooth v0.3.0 => github.com/rbaron/bluetooth release
```

Upon calling `go build`, the line will be replaced with a new pseudo-version.