package config

const ExampleConfig = `general {
	enable_debug_logs: false
	max_refresh_rate_sec: 300
	buffer_size: 5
	log_to_stdout: true
	log_to_influx: false
	log_to_bolt: false
	enable_http: false
}

bluetooth {
	hci_id: -1
}

devices {
	ruuvi_tag { mac: "89-75-AE-8B-52-D9" human_name: "Living Room" }
	ruuvi_tag { mac: "47-36-DA-4A-9F-F6" human_name: "Bedroom" }
}

http {
	listen: ":80"
	source_db: "bolt"
}

influx_db {
	connection: "http://localhost:8086"
	database: "ruuvi"
	table: "ruuvi"
	username: ""
	password: ""
	precision: "s"
	retention_policy: ""
	write_consistency: ""
}

bolt_db {
	path: "/tmp/ruuvi.boltdb"
	bucket_size_sec: 86400
	retention_sec: 604800
}`
