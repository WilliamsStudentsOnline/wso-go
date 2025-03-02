#!/bin/bash -e
##### WSO-Backend WSO 2.0 #####
# Execute Prometheus, Grafana, and Node Exporter.

######## SANITY CHECKING
if [ "$(id -u)" = 0 ]; then
  echo "This script should never be run as root." >&2
  exit 1
fi

if ! command -v grafana > /dev/null && ! command -v grafana-server > /dev/null; then
  echo "You need to install Grafana." >&2
  exit 1
fi

if ! command -v prometheus > /dev/null; then
  echo "You need to install Prometheus." >&2
  exit 1
fi

if ! command -v node-exporter > /dev/null && ! command -v prometheus-node-exporter > /dev/null; then
  echo "You need to install Node Exporter." >&2
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
PROMETHEUS_FOLDER="$PROD_DIR/prometheus"
GRAFANA_FOLDER="$PROD_DIR/grafana"

# Ensure directories exist
mkdir -p "$PROMETHEUS_FOLDER"
mkdir -p "$GRAFANA_FOLDER/{logs,data,plugins}"

######## PROMETHEUS
# Start Prometheus
echo "Starting Prometheus..."
# Change this to match the number of cores
GOMAXPROCS=1
prometheus \
  --config.file="$PROD_DIR/prometheus.yml" \
  --storage.tsdb.path="$PROMETHEUS_FOLDER" 2>&1 \
  --storage.tsdb.retention.time=365d \
  --web.config.file="$PROD_DIR/prometheus-basicauth.yml" \
  --web.listen-address=0.0.0.0:9090 \
  | tee "$PROMETHEUS_FOLDER/prometheus.log" &
PROMETHEUS_PID=$!
sleep 3  # Allow Prometheus to initialize
if ! ps -p "$PROMETHEUS_PID" > /dev/null; then
  echo "Prometheus failed to start." >&2
  exit 1
fi
echo "Prometheus is running (PID: $PROMETHEUS_PID)."

######## GRAFANA
# Start Grafana
echo "Starting Grafana..."
grafana-server \
  --config="$PROD_DIR/grafana-config.ini" \
  --pidfile="$GRAFANA_FOLDER/grafana-server.pid"
  --homepath="$GRAFANA_FOLDER" 2>&1 \
  cfg:default.paths.logs="$GRAFANA_FOLDER/logs" \
  cfg.default.paths.data="$GRAFANA_FOLDER/data" \
  cfg.default.paths.plugins="$GRAFANA_FOLDER/plugins"\
  | tee "$GRAFANA_FOLDER/grafana.log" &
GRAFANA_PID=$!
sleep 3  # Allow Grafana to initialize
if ! ps -p "$GRAFANA_PID" > /dev/null; then
  echo "Grafana failed to start." >&2
  exit 1
fi
echo "Grafana is running (PID: $GRAFANA_PID)."

######## NODE EXPORTER
# Start Node Exporter
echo "Starting Node Exporter..."
if command -v node-exporter > /dev/null; then
  node-exporter \
    --web.listen-address="0.0.0.0:9091" \
    --log.level=info 2>&1 \
    | tee "$PROMETHEUS_FOLDER/node-exporter.log" &
  NODE_EXPORTER_PID=$!
elif command -v prometheus-node-exporter > /dev/null; then
  prometheus-node-exporter \
    --web.listen-address="0.0.0.0:9091" \
    --log.level=info 2>&1 \
    | tee "$PROMETHEUS_FOLDER/node-exporter.log" &
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

######## MAIN FUNCTION
cd ..
wait
exit 0
