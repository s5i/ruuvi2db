# ruuvi2db

ruuvi2db is a service for collecting measurements from RuuviTags and storing them in a database of choice.

## Features

Supported data formats:

- [Data Format 3](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_03.md) (in production)

Output data:

- Temperature (°C)
- Relative humidity (%)
- Air pressure (hPa)
- Battery voltage (mV)
  
Supported databases:

- InfluxDB v1.x

## Requirements

* Linux (should also work with Darwin, untested)
* Bluetooth adapter with BLE support
* Raw capture capabilities for the binary

  ```sh
  sudo setcap "cap_net_raw,cap_net_admin=ep" "$(which ruuvi2db)"
  ```

## Not there yet

Support for the following data formats:

- [Data Formats 2 and 4](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_04.md) (deprecated)
- [Data Format 5](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_05.md) (beta)
- [Data Format 8](https://github.com/ruuvi/ruuvi-sensor-protocols/blob/master/dataformat_08.md) (proposed)

