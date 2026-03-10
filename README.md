# Tutorial: Full-Stack Observability on a Budget: A Guide to Strategic Sampling and Data Optimization - Pavol Loffay, Red Hat

This repository hosts content for tutorial for Kubecon EU 2026 Amsterdam.

Previous tutorials:
* [Tutorial: Exploring the Power of Distributed Tracing with OpenTelemetry on Kubernetes - Pavol Loffay & Benedikt Bongartz, Red Hat; Matej Gera, Coralogix; Anthony Mirabella, AWS; Anusha Reddy Narapureddy, Apple](https://github.com/pavolloffay/kubecon-eu-2024-opentelemetry-kubernetes-tracing-tutorial)
* [Exploring the Power of OpenTelemetry on Kubernetes - Pavol Loffay, Benedikt Bongartz & Yuri Oliveira Sa, Red Hat; Severin Neumann, Cisco; Kristina Pathak, LightStep](https://github.com/pavolloffay/kubecon-eu-2023-opentelemetry-kubernetes-tutorial)
* [Tutorial: Exploring the Power of Metrics Collection with OpenTelemetry on Kubernetes - Pavol Loffay & Benedikt Bongartz, Red Hat; Anthony Mirabella, AWS; Matej Gera, Coralogix; Anusha Reddy Narapureddy, Apple](https://github.com/pavolloffay/kubecon-na-2023-opentelemetry-kubernetes-metrics-tutorial)

__Abstract__: Observability costs can quickly spiral out of control. This tutorial provides a holistic framework for managing these costs without sacrificing insight. We will systematically compare head-based, probabilistic, and tail-based sampling, explaining their trade-offs in cost, computational overhead, and data fidelity. We'll directly address the hidden costs of tail sampling—which can increase compute load—and clarify when to use it. Beyond sampling, you'll learn to profile telemetry to eliminate waste (duplicates, debug logs) and use smart routing to send data to cheaper backends. You will leave equipped to design a cost-effective observability strategy in Kubernetes using OpenTelemetry, choose the right sampling method for your workload, and gain clear visibility into your spending.
__Sched__: - https://kccnceu2026.sched.com/event/2CW3t/tutorial-full-stack-observability-on-a-budget-a-guide-to-strategic-sampling-and-data-optimization-pavol-loffay-red-hat

## Agenda

Internal meeting doc: https://docs.google.com/document/d/1rbc0JqMP7i4koKpxqb9gYovmAlJ_BRN1Ttg3EhY9cbY/edit

Each tutorial step is located in a separate file:

1. [Welcome & Setup](01-welcome-setup.md) (5 min)
1. [Sampling overview](02-sampling-overview.md) (10 min)
1. [Data profiling]() (10 min)
1. [Head based sampling]() (10 minutes)
1. [Tail based sampling]() (20 minutes)
1. [Cleaning logs]() (10 minutes)
1. Wrap up & Questions

