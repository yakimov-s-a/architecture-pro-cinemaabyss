# Проектная работа

## Задание 1

[![Диаграмма контейнеров](diagrams/container/CinemaAbyss_Container.png)](diagrams/container/CinemaAbyss_Container.puml)

## Задание 2

### 1. Proxy

[Proxy](src/microservices/proxy)

### 2. Kafka

[Events](src/microservices/events)

**Тесты:**

![Результаты тестов](screenshots/test-results.png)

**Kafka topics:**

![Kafka topics](screenshots/kafka-topics.png)

## Задание 3

### 1. CI/CD

* [Docker Build and Push](.github/workflows/docker-build-push.yml)
* [API Tests](.github/workflows/api-tests.yml)

### 2. Proxy в Kubernetes

**Movies service logs:**

![Movies service logs](screenshots/movies-service-logs-kubernetes.png)

**Events service logs:**

![Events service logs](screenshots/events-service-logs.png)

## Задание 4

**Helm deployment:**

![Helm deployment](screenshots/helm-deployment.png)

**Movies service logs:**

![Movies service logs](screenshots/movies-service-logs-helm.png)

# Задание 5

[Circuit breaker configuration](src/kubernetes/circuit-breaker-config.yaml)

**Circuit breaker stats:**

![Circuit breaker stats](screenshots/circuit-breaker-stats.png)
