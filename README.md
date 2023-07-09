# ruuvi2db

ruuvi2db is a service for collecting measurements from RuuviTags, storing them
in a database of choice, and displaying the data via HTTP.

Tested on Raspberry Pi Zero W.

## Features

Supported data formats:

- [Data Format 3](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_03.md)
- [Data Format 5](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_05.md)

Output data:

- Temperature (°C)
- Relative humidity (%)
- Air pressure (hPa)
- Battery voltage (mV)

Supported databases:

- BoltDB (custom format)

## Requirements

* Linux 2.6.23+
* Bluetooth adapter with BLE support
* Raw capture capabilities for the binary

## Installation

```sh
wget https://github.com/s5i/ruuvi2db/releases/latest/download/ruuvi2db
chmod +x ./ruuvi2db

sudo mkdir -p /usr/local/ruuvi2db
sudo mv ./ruuvi2db /usr/local/ruuvi2db
sudo chown root:root /usr/local/ruuvi2db/ruuvi2db

sudo tee /etc/systemd/system/ruuvi2db.service << EOF > /dev/null
[Unit]
Description=ruuvi2db Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/ruuvi2db/ruuvi2db
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo /usr/local/ruuvi2db/ruuvi2db --create_config

# Change as desired.
sudo vim /usr/local/ruuvi2db/config.yaml

sudo systemctl enable ruuvi2db
sudo systemctl start ruuvi2db
```

ruuvi2db can also be ran without root privileges.

* Default BoltDB path requires root; change as desired.
* Default HTTP port is 8080. Can be overridden in config.
* Raw capture capabilities need to be granted.

```sh
sudo setcap "cap_net_raw,cap_net_admin=ep" "$(which ruuvi2db)"
```

## Not there yet

Support for the following data formats:

- [Data Formats 2 and 4](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_04.md)
- [Data Format 8](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_08.md)
