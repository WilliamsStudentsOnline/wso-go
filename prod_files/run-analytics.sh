#!/bin/bash -e
##### WSO-Backend WSO 2.0 #####
# Execute Grafana and Prometheus.
if [ $(id -u) = 0 ]; then
   echo "This script should never be run as root." >&2
   exit 1
fi
if ! [[ command -v "grafana" > /dev/null || command -v "grafana-server" > /dev/null ]]; then
  echo "You need to install grafana." >&2
  exit 1
fi
if ! command -v "prometheus" > /dev/null; then
  echo "You need to install prometheus." >&2
  exit 1
fi
echo "Executing Prometheus and Grafana..."
cd prod_files
prometheus --config.file=prometheus.yml &
grafana --config grafana-config.ini &
cd ..
wait
exit 0

