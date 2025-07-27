# IoT Smart Irrigation System Architecture

## System Overview

This is a comprehensive multi-language IoT messaging system designed for smart irrigation management. The system follows clean architecture principles, domain-driven design patterns, and microservices architecture.

## Folder Structure Design

```
iot-server-smart-irrigation/
├── services/                    # Core microservices
│   ├── go/                     # Go-based services
│   │   ├── mqtt-gateway/       # MQTT message gateway service
│   │   ├── device-manager/     # Device management service
│   │   ├── data-processor/     # High-performance data processing
│   │   └── notification-service/ # Real-time notifications
│   ├── python/                 # Python-based services
│   │   ├── api-gateway/        # Main REST API gateway
│   │   ├── ml-engine/          # Machine learning processing
│   │   ├── llm-orchestrator/   # LLM integration service
│   │   └── analytics-service/  # Data analytics and reporting
│   └── frontend/               # Frontend applications
│       ├── web-dashboard/      # React/Vue.js admin dashboard
│       ├── mobile-app/         # React Native mobile app
│       └── monitoring-ui/      # System monitoring interface
├── shared/                     # Shared components and configurations
│   ├── schemas/                # Data schemas and models
│   ├── protocols/              # Communication protocols
│   ├── configs/                # Shared configurations
│   └── libraries/              # Shared code libraries
├── infrastructure/             # Infrastructure and deployment
│   ├── docker/                 # Docker configurations
│   ├── k8s/                   # Kubernetes manifests
│   ├── mqtt/                  # MQTT broker configuration
│   ├── databases/             # Database schemas and migrations
│   └── monitoring/            # Monitoring and observability
├── scripts/                   # Automation and utility scripts
├── docs/                     # System documentation
├── tests/                    # Integration and E2E tests
└── .ci/                     # CI/CD pipeline configurations
```

## Service Interactions

### MQTT Message Flow
1. IoT devices → MQTT Broker → Go MQTT Gateway
2. Go services process high-frequency sensor data
3. Python services handle complex analytics and ML
4. Frontend applications consume processed data via APIs

### Data Flow Architecture
- **Ingestion**: MQTT Gateway (Go) → Message Queue
- **Processing**: Data Processor (Go) + ML Engine (Python)
- **Storage**: Time-series DB + Relational DB
- **Consumption**: API Gateway (Python) → Frontend Apps

### Security Layers
- Device authentication via certificates
- MQTT TLS encryption
- Service-to-service mTLS
- API authentication and authorization
- Network segmentation via Docker/K8s

## Development Workflow
1. Local development using Docker Compose
2. Shared schemas ensure consistent data models
3. Independent service deployment and testing
4. Centralized configuration management
5. Automated CI/CD with environment promotion