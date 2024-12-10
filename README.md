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
# Create a non-privileged user.
sudo useradd -m ruuvi2db
sudo su - ruuvi2db

# Download the binary.
wget https://github.com/s5i/ruuvi2db/releases/latest/download/ruuvi2db
chmod +x ./ruuvi2db

# Grant the necessary capabilities (raw bluetooth, privileged ports).
sudo setcap "cap_net_raw,cap_net_admin,cap_net_bind_service=ep" ./ruuvi2db

# Set up the config.
mkdir -p ~/.ruuvi2db/
wget https://raw.githubusercontent.com/s5i/ruuvi2db/refs/heads/master/example.cfg -O ~/.ruuvi2db/ruuvi2db.cfg
"${EDITOR:-vi}" ~/.ruuvi2db/ruuvi2db.cfg

# Set up the systemd service.
sudo tee /etc/systemd/system/ruuvi2db.service << EOF > /dev/null
[Unit]
Description=ruuvi2db Service
Requires=network.target
Requires=bluetooth.target

[Service]
User=ruuvi2db
Type=simple
ExecStartPre=sleep 10
ExecStart=/home/ruuvi2db/ruuvi2db
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable ruuvi2db
sudo systemctl start ruuvi2db

# Set up aliases.
curl "http://localhost:8082/admin/set_alias?addr=AA:AA:AA:AA:AA:AA&name=AA"
```
### Docker (experimental)

```sh
# Choose the docker mountpoint.
export RUUVI2DB_CONFIG_DIR="/docker/ruuvi2db/cfg"
export RUUVI2DB_DATA_DIR="/docker/ruuvi2db/data"

# Create and start the container.
docker run -d \
  --name ruuvi2db \
  -v "${RUUVI2DB_CONFIG_DIR}:/cfg" \
  -v "${RUUVI2DB_DATA_DIR}:/data" \
  --restart=always \
  --net=host --cap-add NET_RAW --cap-add NET_ADMIN \
  shyym/ruuvi2db

# Edit the skeleton config.
${EDITOR:-vi} ${RUUVI2DB_CONFIG_DIR}/ruuvi2db.cfg
```

Note: `--net=host --cap-add NET_RAW --cap-add NET_ADMIN` are only required for
the `reader` module.

## Not there yet

Support for the following data formats:

- [Data Formats 2 and 4](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_04.md)
- [Data Format 8](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_08.md)
