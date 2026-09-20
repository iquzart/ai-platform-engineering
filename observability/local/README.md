# Local observability stack

This Compose stack provides Grafana OSS, OpenTelemetry Collector Contrib,
Prometheus, Tempo, and Loki for local development. Applications send traces,
metrics, and logs **only** to the Collector's OTLP endpoints:

| Protocol | Endpoint |
| --- | --- |
| OTLP/gRPC | `localhost:4317` |
| OTLP/HTTP | `http://localhost:4318` |

The Collector exports traces to Tempo, logs to Loki through Loki's OTLP
ingestion endpoint, and exposes received metrics for Prometheus to scrape.
Tempo and Loki have no host-published ports. Grafana (`localhost:3000`) and
Prometheus (`localhost:9090`) are loopback-only developer endpoints.

## Image versions and update policy

The stack uses exact upstream stable-release tags, not floating tags such as
`latest`, so a checked-out `.env` produces the same component versions. The
current pins are OpenTelemetry Collector Contrib `0.161.0`, Prometheus
`v3.14.0`, Tempo `3.0.3`, Loki `3.7.8`, and Grafana OSS `13.2.2`. They are
defined in `.env.example` and may be overridden in a local `.env` only for
tested rollback or upgrade work. Refresh all pins together after checking the
upstream releases and validating this Compose stack; do not use image digests
for this current-version policy.

Versions were selected from the projects' stable GitHub releases on 2026-09-20.
The Collector's `otlp/tempo` exporter uses Tempo's OTLP/gRPC receiver at
`tempo:4317`; its `otlphttp/loki` exporter uses Loki's OTLP HTTP endpoint at
`http://loki:3100/otlp/v1/logs`, which remains enabled by this Loki 3.x
configuration.

## Start

```sh
cd observability/local
cp .env.example .env
# Replace GRAFANA_ADMIN_PASSWORD in .env with a unique local password.
docker compose up -d
docker compose ps
```

Open Grafana at `http://localhost:3000` and sign in with the credentials in
`.env`. Prometheus is available at `http://localhost:9090`.

Grafana provisions Prometheus, Tempo, and Loki datasources. Prometheus
exemplars link to Tempo; Tempo trace views link to Loki logs.

## Configuration and retention

All service configuration mounts are read-only. Named volumes persist
Prometheus, Tempo, Loki, and Grafana data. Prometheus, Tempo, and Loki retain
data for seven days (168 hours), subject to local disk capacity. The Collector
health endpoint is internal at port 13133.

## Validate and stop

```sh
GRAFANA_ADMIN_PASSWORD=local-validation-only docker compose config
docker compose ps
docker compose down
```

`docker compose down` stops the stack but retains data. To also remove all
local observability data, run `docker compose down -v`. Do not use that command
if dashboards or telemetry history must be retained.

## Application configuration

For a host-process application, configure either `OTEL_EXPORTER_OTLP_ENDPOINT`
as `http://localhost:4318` (HTTP) or `localhost:4317` (gRPC). A containerized
application that joins the `local-observability` network should use
`otel-collector:4318` or `otel-collector:4317`. Do not point applications at
Prometheus, Tempo, or Loki directly.
