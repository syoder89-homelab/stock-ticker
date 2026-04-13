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
| `OTLP_ENDPOINT` | _(empty)_ | OpenTelemetry collector endpoint |

## Building

```bash
# Build locally
go build -o stock-ticker ./cmd/stock-ticker

# Build container image
docker build -t stock-ticker .
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

## Deploying to Kubernetes

Basic manifests are in the [`deploy/`](deploy/) directory:

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
   # Edit deploy/secret.yaml and replace "demo" with your real Alpha Vantage API key,
   # or create the secret imperatively:
   kubectl create namespace stock-ticker
   kubectl -n stock-ticker create secret generic stock-ticker-api-key \
     --from-literal=API_KEY=your-key-here
   ```

2. **Apply all manifests:**

   ```bash
   kubectl apply -f deploy/
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

### Customization

- **Change the stock symbol or history window:** Edit `deploy/configmap.yaml` and re-apply.
- **Use a different image tag:** Update the `image:` field in `deploy/deployment.yaml`.
- **Configure ingress:** Edit the `host` in `deploy/ingress.yaml` to match your domain. Add `ingressClassName` or annotations as needed for your ingress controller.
