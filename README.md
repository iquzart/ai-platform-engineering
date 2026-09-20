
<div align="center">

# AI Platform Engineering

**Open-source learning, experimentation, and engineering for building AI platforms.**

<p>
  <img src="https://img.shields.io/badge/AI-Platform-blue" alt="AI Platform">
  <img src="https://img.shields.io/badge/Open%20Source-First-green" alt="Open Source First">
  <img src="https://img.shields.io/badge/Local-First-orange" alt="Local First">
  <img src="https://img.shields.io/badge/Kubernetes-Ready-326CE5" alt="Kubernetes">
</p>

</div>

An open-source learning and experimentation platform for building the infrastructure, runtime, and engineering foundations required to build, deploy, operate, and scale AI applications and agents.

This repository is a **monorepo of experiments, reference implementations, architectures, and reusable components** for building an AI platform.

The goal is not to build a single AI application. The goal is to understand and build the platform underneath AI applications.

## Areas

```text
ai-platform-engineering/
│
├── agents/              # Agents, tools, MCP, memory and runtimes
├── ai-gateway/          # Model routing, access, rate limits and cost controls
├── rag/                 # Retrieval Augmented Generation
├── semantic-cache/      # Semantic caching with Redis, pgvector and embeddings
├── memory/              # AI memory and state
│
├── models/              # Models, inference and model providers
├── inference/           # vLLM, GPU inference and model serving
├── gpu/                 # CUDA, GPU infrastructure and Kubernetes
│
├── workflows/           # Durable workflows and orchestration
├── security/            # Guardrails, permissions and AI security
├── evaluation/          # LLM, RAG, agent and prompt evaluation
├── observability/       # LLM, agent and AI platform observability
├── data/                # Data ingestion and processing
│
├── infrastructure/      # Docker, Kubernetes, Helm and Terraform
├── examples/             # Complete reference implementations
├── experiments/         # Small isolated learning experiments
└── docs/                 # Architecture, concepts and decisions
```

## Platform

The different components are explored as building blocks of an AI platform:

```text
                         AI APPLICATIONS
                               │
                               ▼
                         ┌───────────┐
                         │ AI Gateway│
                         └─────┬─────┘
                               │
             ┌─────────────────┼─────────────────┐
             ▼                 ▼                 ▼
          Models             Agents             RAG
             │                 │                 │
             │            ┌────┼────┐            │
             │            ▼    ▼    ▼            │
             │          Tools MCP Memory          │
             │                 │                 │
             └─────────────────┼─────────────────┘
                               │
                    ┌──────────┼──────────┐
                    ▼          ▼          ▼
                 Cache      Workflows   Security
                    │          │          │
                    └──────────┼──────────┘
                               │
                    ┌──────────┼──────────┐
                    ▼          ▼          ▼
               Evaluation Observability Infrastructure
                               │
                               ▼
                    ┌──────────────────────┐
                    │   INFERENCE PLATFORM │
                    ├──────────────────────┤
                    │ vLLM                 │
                    │ CUDA                 │
                    │ GPU Kubernetes       │
                    │ Model Serving        │
                    └──────────────────────┘
```

## Goals

* Learn how modern AI platforms work.
* Experiment with open-source technologies.
* Build small, reproducible implementations.
* Understand the infrastructure behind AI applications and agents.
* Explore GPU-enabled AI inference and model serving.
* Turn useful experiments into reusable platform components.
* Explore how these components work together in real environments.

## Engineering Principles

* **Open source first**
* **Local first** where practical
* **Reproducible** experiments
* **Observable** by default
* **Secure by default**
* **Modular** components
* **Human controlled** for consequential actions

## Experiments

Experiments are small, focused implementations used to answer specific technical questions.

```text
experiments/
├── 001-rag/
├── 002-semantic-cache/
├── 003-ai-gateway/
├── 004-agent-tools/
├── 005-agent-memory/
├── 006-vllm/
├── 007-gpu-kubernetes/
└── ...
```

Useful experiments may later become reusable components under the main platform directories.

## Inference & GPU

A core part of the platform is understanding how AI models are actually served.

Areas of exploration include:

```text
inference/
├── vllm/
├── model-serving/
├── batching/
├── quantization/
└── benchmarks/

gpu/
├── cuda/
├── kubernetes/
├── gpu-operator/
├── device-plugin/
├── scheduling/
├── node-pools/
└── monitoring/
```

Example target architecture:

```text
                       AI Gateway
                            │
                            ▼
                     Model Router
                            │
                            ▼
                    Kubernetes Service
                            │
                            ▼
                         vLLM
                            │
                            ▼
                    NVIDIA GPU Node
                            │
                         CUDA
                            │
                            ▼
                          GPU
```

This area will explore:

* vLLM
* CUDA
* NVIDIA GPU Operator
* Kubernetes GPU scheduling
* GPU-enabled node pools
* Model serving
* Quantization
* Continuous batching
* Inference performance
* GPU utilization
* Model scaling
* Inference observability

## Technology

The platform will explore technologies such as:

* Ollama
* vLLM
* Hosted LLM providers
* LiteLLM
* Envoy
* PostgreSQL / pgvector
* Redis
* Temporal
* OpenTelemetry
* Prometheus
* Loki
* Tempo
* Grafana
* CUDA
* Kubernetes
* NVIDIA GPU Operator
* Docker
* Helm
* Terraform
* MCP and agent frameworks

The technology choices are intentionally not fixed and will evolve as the experiments progress.

## Status

This is an evolving learning and engineering repository.

Some components are experiments, some are reference implementations, and some may eventually become reusable AI platform components.
