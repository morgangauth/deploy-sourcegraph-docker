#!/usr/bin/env bash
set -e
source ./replicas.sh

# Description: Serves the frontend of Sourcegraph via HTTP(S).
#
# Disk: 128GB / non-persistent SSD
# Network: 100mbps
# Liveness probe: HTTP GET http://sourcegraph-frontend:3080/healthz
# Ports exposed to other Sourcegraph services: 6060/TCP
# Ports exposed to the public internet: 3080 (HTTP) and/or 3443 (HTTPS)
#
VOLUME="$HOME/sourcegraph-docker/sourcegraph-frontend-$1-disk"
./ensure-volume.sh $VOLUME 100
docker run --detach \
    --name=sourcegraph-frontend-$1 \
    --network=sourcegraph \
    --restart=always \
    --cpus=4 \
    --memory=8g \
    --health-cmd="wget -q 'http://127.0.0.1:3080/healthz' -O /dev/null || exit 1" \
    --health-interval=5s \
    --health-timeout=10s \
    --health-retries=5 \
    --health-start-period=300s \
    -e DEPLOY_TYPE=pure-docker \
    -e GOMAXPROCS=12 \
    -e JAEGER_AGENT_HOST=jaeger \
    -e PGHOST=pgsql \
    -e CODEINTEL_PGHOST=codeintel-db \
    -e CODEINSIGHTS_PGDATASOURCE=postgres://postgres:password@codeinsights-db:5432/postgres \
    -e SRC_GIT_SERVERS="$(addresses "gitserver-" $NUM_GITSERVER ":3178")" \
    -e SRC_SYNTECT_SERVER=http://syntect-server:9238 \
    -e SEARCHER_URL="$(addresses "http://searcher-" $NUM_SEARCHER ":3181")" \
    -e SYMBOLS_URL="$(addresses "http://symbols-" $NUM_SYMBOLS ":3184")" \
    -e INDEXED_SEARCH_SERVERS="$(addresses "zoekt-webserver-" $NUM_INDEXED_SEARCH ":6070")" \
    -e SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090 \
    -e REPO_UPDATER_URL=http://repo-updater:3182 \
    -e GRAFANA_SERVER_URL=http://grafana:3370 \
    -e JAEGER_SERVER_URL=http://jaeger:16686 \
    -e GITHUB_BASE_URL=http://github-proxy:3180 \
    -e PROMETHEUS_URL=http://prometheus:9090 \
    -v $VOLUME:/mnt/cache \
    -p 0.0.0.0:$((3080 + $1)):3080 \
    index.docker.io/sourcegraph/frontend:3.31.2@sha256:f82223d843fc1422b914da3338d70a6b634793d43947ebe7a3b3403ea974d462

# Note: SRC_GIT_SERVERS, SEARCHER_URL, and SYMBOLS_URL are space-separated
# lists which each allow you to specify more container instances for scaling
# purposes. Be sure to also apply such a change here to the frontend-internal
# service.

echo "Deployed sourcegraph-frontend $1 service"
