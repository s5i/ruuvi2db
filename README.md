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

### Docker Compose

```sh
# Choose a path for local files.
RUUVI2DB_PATH="/docker/ruuvi2db"

# Change as desired.
tee compose.yaml << EOF > /dev/null
services:
  ruuvi2db:
    container_name: ruuvi2db
    image: shyym/ruuvi2db:latest
    restart: always
    volumes:
      - ${RUUVI2DB_PATH}:/appdata
    network_mode: "host"  # Required for the reader module.
    cap_add:
      - "NET_ADMIN"  # Required for the reader module.
      - "NET_RAW"  # Required for the reader module.
EOF

mkdir -p ${RUUVI2DB_PATH}
sudo docker compose up --pull=always --force-recreate --detach

# Edit the config.
${EDITOR:-vi} "${RUUVI2DB_PATH}/config.yaml"

# Remove the sentinel to start the service.
rm "${RUUVI2DB_PATH}/SENTINEL.readme"

# Set up aliases.
curl "http://localhost:8082/admin/set_alias?addr=AA:AA:AA:AA:AA:AA&name=AA"
```

## Not there yet

Support for the following data formats:

- [Data Formats 2 and 4](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_04.md)
- [Data Format 8](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_08.md)
