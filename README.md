# stock-ticker

Stock Ticker web service and deployment demonstration.

A Go service that fetches stock price data from the [Alpha Vantage API](https://www.alphavantage.co/), calculates a daily average closing price over configurable days, and serves the results via a JSON HTTP API.

## API

| Endpoint | Description |
|---|---|
| `GET /api/v1/ticker` | Returns stock data and daily average |
| `GET /healthz` | Liveness probe |
| `GET /readyz` | Readiness probe |
| `GET /startupz` | Startup probe |
| `GET /metrics` | Prometheus metrics endpoint |

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|---|---|---|
| `API_KEY` | `demo` | Alpha Vantage API key |
| `SYMBOL` | `MSFT` | Stock symbol to track |
| `NDAYS` | `7` | Number of days of history |
| `SERVICE_ADDR` | `:8080` | Server bind address |
| `QUOTE_SERVICE_URL` | `https://www.alphavantage.co/query` | Alpha Vantage endpoint |
| `FUNCTION` | `TIME_SERIES_DAILY` | API function |
| `LOG_LEVEL` | `DEBUG` | Log level (DEBUG, INFO, WARN, ERROR) |
| `DISABLE_METRICS` | _(empty)_ | Set to `true` or `1` to disable Prometheus metrics |
| `OTLP_ENDPOINT` | _(empty)_ | OpenTelemetry collector endpoint |

## Building

```bash
# Build locally
go build -ldflags "-X main.version=dev -X main.commit=$(git rev-parse --short HEAD)" \
    -o stock-ticker ./cmd/stock-ticker

# Build container image
docker build \
    --build-arg VERSION=dev \
    --build-arg COMMIT=$(git rev-parse --short HEAD) \
    -t stock-ticker .
```

## Running Locally

```bash
export API_KEY="your-alpha-vantage-key"
go run ./cmd/stock-ticker
# → http://localhost:8080/api/v1/ticker
```

## Testing

```bash
# Run unit tests
go test ./...

# Run the live Alpha Vantage integration test explicitly
API_KEY="your-alpha-vantage-key" go test -tags=integration ./...
```

The integration test is excluded from default test runs so normal development and CI do not consume the Alpha Vantage daily quota.

## CI/CD

A GitHub Actions workflow ([`.github/workflows/built-test-push.yml`](.github/workflows/built-test-push.yml)) runs on every push and PR to `main`:

1. **Build** — Builds the Docker image (includes `go test` inside the Dockerfile) and injects version/commit via build args. Runs on all pushes and PRs.
2. **Push** — On merge to `main` only, authenticates to GCP via Workload Identity Federation and pushes the image to Google Artifact Registry. Images are tagged `YYYYMMDD-<short sha>`.

## Metrics

The service exposes Prometheus metrics at `GET /metrics`.

Custom metrics include:

- `stock_ticker_ticker_requests_total`
- `stock_ticker_ticker_request_duration_seconds`
- `stock_ticker_ticker_errors_total`
- `stock_ticker_upstream_requests_total`
- `stock_ticker_upstream_request_duration_seconds`
- `stock_ticker_upstream_errors_total`

## Deploying to Kubernetes

Basic manifests are in the [`deploy/k8s/`](deploy/k8s/) directory:

| File | Resource | Purpose |
|---|---|---|
| `namespace.yaml` | Namespace | Creates the `stock-ticker` namespace |
| `configmap.yaml` | ConfigMap | Non-sensitive configuration (`SYMBOL`, `NDAYS`) |
| `secret.yaml` | Secret | Alpha Vantage API key |
| `deployment.yaml` | Deployment | Runs the application container |
| `service.yaml` | Service | Exposes the app within the cluster (ClusterIP) |
| `ingress.yaml` | Ingress | Routes external traffic to the service |

### Quick Start

1. **Set your API key** in the secret manifest before applying:

   ```bash
   # Create the api key secret imperatively with your real Alpha Vantage API key:
   kubectl create namespace stock-ticker
   kubectl -n stock-ticker create secret generic stock-ticker-api-key \
     --from-literal=API_KEY=your-key-here
   ```

2. **Apply all manifests:**

   ```bash
   kubectl apply -f deploy/k8s/
   ```

3. **Verify the deployment:**

   ```bash
   kubectl -n stock-ticker get pods
   kubectl -n stock-ticker logs deploy/stock-ticker
   ```

4. **Access the service** (port-forward for local testing):

   ```bash
   kubectl -n stock-ticker port-forward svc/stock-ticker 8080:8080
   curl http://localhost:8080/api/v1/ticker
   ```

### Helm

A Helm chart used for the production deployment is in [`deploy/helm/`](deploy/helm/). This is a copy of the chart deployed via [homelab-apps](https://github.com/syoder89/homelab-apps).

### Customization

- **Change the stock symbol or history window:** Edit `deploy/k8s/configmap.yaml` and re-apply.
- **Use a different image tag:** Update the `image:` field in `deploy/k8s/deployment.yaml`.
- **Configure ingress:** Edit the `host` in `deploy/k8s/ingress.yaml` to match your domain. Add `ingressClassName` or annotations as needed for your ingress controller.

### Prometheus Metrics

The deployment includes standard Prometheus scrape annotations on the pod template:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

This works out of the box with Prometheus instances configured to use annotation-based service discovery.

If you are using the **Prometheus Operator**, create a `ServiceMonitor` instead:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: stock-ticker
  namespace: stock-ticker
  labels:
    app: stock-ticker
spec:
  selector:
    matchLabels:
      app: stock-ticker
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

Apply it alongside the other manifests. Ensure your Prometheus Operator is configured to select `ServiceMonitor` resources from the `stock-ticker` namespace (via `serviceMonitorNamespaceSelector` or a cluster-wide configuration).
