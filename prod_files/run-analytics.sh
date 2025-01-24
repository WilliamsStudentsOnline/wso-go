#!/bin/bash -e
##### WSO-Backend WSO 2.0 #####
# Execute Prometheus, Grafana, and Node Exporter.

if [ "$(id -u)" = 0 ]; then
  echo "This script should never be run as root." >&2
  exit 1
fi

if ! command -v grafana > /dev/null && ! command -v grafana-server > /dev/null; then
  echo "You need to install grafana." >&2
  exit 1
fi

if ! command -v prometheus > /dev/null; then
  echo "You need to install prometheus." >&2
  exit 1
fi

if ! command -v node-exporter > /dev/null && ! command -v prometheus-node-exporter > /dev/null; then
  echo "You need to install node exporter." >&2
  exit 1
fi

echo "Starting Prometheus, Grafana, and Node Exporter..."

# Check if the current directory is prod_files
if [ "$(basename "$PWD")" != "prod_files" ]; then
  if [ -d prod_files ]; then
    cd prod_files || exit 1
  else
    echo "prod_files directory not found." >&2
    exit 1
  fi
fi

# Generate absolute paths
PROD_DIR="$(pwd)"
PROMETHEUS_CONFIG="$PROD_DIR/prometheus.yml"
PROMETHEUS_STORAGE_PATH="$PROD_DIR/prometheus-data"
GRAFANA_CONFIG="$PROD_DIR/grafana-config.ini"
GRAFANA_DATAPATH="$PROD_DIR/grafana-data"
GRAFANA_DB_PATH="$PROD_DIR/grafana.db"  # Path to your preexisting grafana.db
NODE_EXPORTER_LOG_PATH="$PROD_DIR/node-exporter-logs/node-exporter.log"

# Ensure directories exist
mkdir -p "$PROMETHEUS_STORAGE_PATH"
mkdir -p "$GRAFANA_DATAPATH"
mkdir -p "$(dirname "$NODE_EXPORTER_LOG_PATH")"

# Set Grafana's database path
export GF_DATABASE_PATH="$GRAFANA_DB_PATH"

# Start Prometheus
echo "Starting Prometheus..."
prometheus --config.file="$PROMETHEUS_CONFIG" --storage.tsdb.path="$PROMETHEUS_STORAGE_PATH" 2>&1 | tee "$PROD_DIR/prometheus.log" &
PROMETHEUS_PID=$!
sleep 3  # Allow Prometheus to initialize
if ! ps -p "$PROMETHEUS_PID" > /dev/null; then
  echo "Prometheus failed to start." >&2
  exit 1
fi
echo "Prometheus is running (PID: $PROMETHEUS_PID)."

# Set Grafana data path environment variable
export GF_PATHS_DATA="$GRAFANA_DATAPATH"
export GF_PATHS_LOGS="$PROD_DIR/grafana-logs"  # optional: log path
export GF_PATHS_PLUGINS="$PROD_DIR/grafana-plugins"  # optional: plugins path

# Start Grafana
echo "Starting Grafana..."
grafana-server --config="$GRAFANA_CONFIG" --homepath="/usr/share/grafana" 2>&1 | tee "$PROD_DIR/grafana.log" &
GRAFANA_PID=$!
sleep 3  # Allow Grafana to initialize
if ! ps -p "$GRAFANA_PID" > /dev/null; then
  echo "Grafana failed to start." >&2
  exit 1
fi
echo "Grafana is running (PID: $GRAFANA_PID)."

# Start Node Exporter
echo "Starting Node Exporter..."
if command -v node-exporter > /dev/null; then
  node-exporter --web.listen-address="0.0.0.0:9120" --log.level=info 2>&1 | tee "$NODE_EXPORTER_LOG_PATH" &
  NODE_EXPORTER_PID=$!
elif command -v prometheus-node-exporter > /dev/null; then
  prometheus-node-exporter --web.listen-address="0.0.0.0:9120" --log.level=info 2>&1 | tee "$NODE_EXPORTER_LOG_PATH" &
  NODE_EXPORTER_PID=$!
else
  echo "Node Exporter failed to start." >&2
  exit 1
fi
sleep 3  # Allow Node Exporter to initialize
if ! ps -p "$NODE_EXPORTER_PID" > /dev/null; then
  echo "Node Exporter failed to start." >&2
  exit 1
fi
echo "Node Exporter is running (PID: $NODE_EXPORTER_PID)."

cd ..
wait
exit 0
