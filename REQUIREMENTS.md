# Base Requirements

## Overview
Write a web service that looks up a fixed number of closing prices of a specific stock. Deploy the web service to Kubernetes.

## Web Service
- Written in golang
- In response to a GET request, the service should return up to the last NDAYS days of data along with the average closing price over those days. The structure of the response is up to you.
- The stock SYMBOL (the symbol to look up) and NDAYS (the number of days) are environment variables provided to your program.
- Use the free quote service from www.alphavantage.co
    - API documentation: https://www.alphavantage.co/documentation/
    - Sample query: https://www.alphavantage.co/query?apikey=demo&function=TIME_SERIES_DAILY_ADJUSTED&symbol=MSFT
- API key is to be provided via secret.
- The API has a quota per key - 25 requests per day.
- Create a docker image that runs your web service.

## Kubernetes Deployment
- Publish your docker image, your code, and provide instructions on how to build the image and run it.
    - Provide generic instructions for local builds using a provided Dockerfile
- Create a Kubernetes manifest that includes a deployment, service and exposes it as an ingress
- Use a configmap for passing in all environment variables. For the exercise use SYMBOL=MSFT and NDAYS=7
- Use a secret to pass in the API key APIKEY=<redacted>
- The sample provided should run on a vanilla Kubernetes environment (minikube, for example).
- Simple single manifest bundled with the application source repo.

---

# My Additional Requirements 

## Web Service
- Include unit tests.
- Create integration tests but due to the severe API quota limit do not enable them by default.
- Include k8s probe endpoints
- Include Prometheus metrics endpoints
    - Create custom metrics for counters, errors, etc and include latency for the external API call
- Must utilize a modern CI/CD pipeline for building, testing and pushing artifacts.

## Kubernetes Deployment
- Must provide a comprehensive, resilient and feature-rich deployment plan. Provide advanced options which are commonly used in large-scale deployments such as pod afinity / anti-afinity for both node and AZ exclusion, HPA, etc.
- Must utilize an external secrets pattern.
