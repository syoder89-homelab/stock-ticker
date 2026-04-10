# Implementation Plan

## Design Decisions
- No SLA or SLOs for uptime or latency are provided — using best-effort optimizations for cost and uptime.
- Deployment can be replicas 0. No need for an HPA or VPA for such a simple API quota limited service.
- No need for PDB — should be able to simply set RollingUpdate strategy and maxUnavailable -1 to ensure service during updates / pod evictions.

## Local Development
- Provide generic instructions for local deployments with Kind
- Include a test deployment / validation to Kind in the GHA pipeline

## CI/CD
- Utilize a GitHub Actions workflow for the docker build, test and push
- OIDC auth to Google Artifact Registry set up by syoder88-homelab/homelab-infra/terraform/artifact-registry

## Homelab Deployment
- Application in homelab-apps to deploy the image from GAR, use LoadBalancer service for my cluster
- Provide a link to the homelab-apps repo
- Scrape with Prometheus — need to add ServiceMonitor for my cluster

## Future Investigation
- Investigate an HPA and VPA enabled for test/prod
    - I would expect this service to be I/O bound (outbound API calls) rather than CPU bound.
    - For a production service with higher quota this service would more likely scale with an HPA using the number of in-flight outbound API calls and set default min/max based on testing and real-world observations after deployment
- Investigate adding a PDB vs set RollingUpdate strategy and maxUnavailable -1.

# Resources
- sample-reply.json
- https://github.com/syoder89-homelab/stock-ticker/
- https://github.com/syoder89-homelab/homelab-apps/tree/main/applications/stock-ticker/