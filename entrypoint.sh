#!/bin/bash

BINARY="/app/ruuvi2db.app"
EXAMPLE_CONFIG="/app/example_config.yaml"
CONFIG="/appdata/config.yaml"
SENTINEL="/appdata/SENTINEL.readme"

if [[ ! -f ${CONFIG} ]]; then
    echo "Remove this file after making necessary changes to ${CONFIG}" > "${SENTINEL}"
    cp "${EXAMPLE_CONFIG}" "${CONFIG}"
fi

if [[ -f ${SENTINEL} ]]; then
    echo "Remove ${SENTINEL} to start the service."
fi
until [[ ! -f ${SENTINEL} ]]; do
    sleep 5
done

echo "Starting the service..."
${BINARY} --config "${CONFIG}"
