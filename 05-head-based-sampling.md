# Head based sampling

The sampling decision is made at the start of the trace (the "head") and propagated to all downstream services via trace context. The first service in the request chain decides whether to sample, and all downstream services honor that decision.

## How it works

1. A request arrives at the entry point service
2. The sampler makes a decision (e.g., 10% probability)
3. The decision is encoded in the `traceparent` header (`sampled` flag)
4. All downstream services read the flag and follow the decision
5. Sampled spans are exported; unsampled spans are dropped

```mermaid
flowchart LR
    subgraph "Sampling Decision at Entry Point"
        R[Request] --> D{Sample?<br/>10% ratio}
        D -->|Yes ✓| S1[Service A<br/>sampled=true]
        D -->|No ✗| U1[Service A<br/>sampled=false]
    end

    subgraph "Decision Propagated Downstream"
        S1 -->|trace context| S2[Service B<br/>sampled=true]
        S2 -->|trace context| S3[Service C<br/>sampled=true]
        S1 --> E1[Export to collector]
        S2 --> E1
        S3 --> E1

        U1 -->|trace context| U2[Service B<br/>sampled=false]
        U2 -->|trace context| U3[Service C<br/>sampled=false]
        U1 --> E2[Dropped]
        U2 --> E2
        U3 --> E2
    end

    style S1 fill:#4CAF50
    style S2 fill:#4CAF50
    style S3 fill:#4CAF50
    style E1 fill:#4CAF50
    style U1 fill:#64B5F6
    style U2 fill:#64B5F6
    style U3 fill:#64B5F6
    style E2 fill:#64B5F6
```

## Head based sampling configuration

### Env variables

```bash
OTEL_TRACES_SAMPLER="parentbased_traceidratio" # or always_on, always_off, traceidratio, parentbased_always_on, parentbased_always_off, , parentbased_jaeger_remote, jaeger_remote
OTEL_TRACES_SAMPLER_ARG="0.2" # or for jaeger_remote|parentbased_jaeger_remote: KV of endpoint, pollingIntervalMs, initialSamplingRate
```

### SDK

```go
provider := trace.NewTracerProvider(
    trace.WithSampler(trace.AlwaysSample()),
)
```

### Declarative configuration

The [declarative configuration](https://opentelemetry.io/docs/languages/sdk-configuration/declarative-configuration/) (experimental) allows configuring samplers via YAML files:

```yaml
tracer_provider:
  sampler:
    parentbased:
      root:
        traceidratio:
          sampling_ratio: 0.25
      remote_parent_sampled:
        always_on:
      remote_parent_not_sampled:
        always_off:
```

Currently Java is the primary SDK with full declarative configuration support. Support in SDKs is documented in [language-support](https://github.com/open-telemetry/opentelemetry-configuration/blob/main/language-support-status.md).

### Instrumentation CR

```yaml
apiVersion: opentelemetry.io/v1alpha1
kind: Instrumentation
spec:
  sampler:
    type: parentbased_traceidratio
    argument: "1"
```

## Exercise: Change sampling rate to 50%

We are using the `Instrumentation` CR to manage the configuration for the SDKs in the cluster.
Therefore we need to configure the sampling rate in the `Instrumentation` CR: `spec.sampler.argument: 0.5`.

* Change the sampler argument to `0.5` in the [app/01-instrumentation.yaml](./app/01-instrumentation.yaml)

```bash
kubectl apply -f app/01-instrumentation.yaml
make restart
```

![Drop in span count in receiver](./images/p8s-head-sampling-drop-receiver-accepted-spans.png)
![Drop in span count](./images/p8s-head-sampling-drop-span-count.png)


- [All signals per second (traces, metrics, logs)](http://localhost:9090/query?g0.expr=label_replace%28sum%28rate%28otelcol_receiver_accepted_spans_total%5B5m%5D%29%29%2C+%22signal%22%2C+%22traces%22%2C+%22%22%2C+%22%22%29%0Aor%0Alabel_replace%28sum%28rate%28otelcol_receiver_accepted_metric_points_total%5B5m%5D%29%29%2C+%22signal%22%2C+%22metrics%22%2C+%22%22%2C+%22%22%29%0Aor%0Alabel_replace%28sum%28rate%28otelcol_receiver_accepted_log_records_total%5B5m%5D%29%29%2C+%22signal%22%2C+%22logs%22%2C+%22%22%2C+%22%22%29&g0.show_tree=0&g0.tab=graph&g0.range_input=1h&g0.res_type=auto&g0.res_density=medium&g0.display_mode=lines&g0.show_exemplars=0&g1.expr=label_replace%28sum%28rate%28otelcol_receiver_refused_spans_total%5B5m%5D%29%29%2C+%22signal%22%2C+%22traces%22%2C+%22%22%2C+%22%22%29%0Aor%0Alabel_replace%28sum%28rate%28otelcol_receiver_refused_metric_points_total%5B5m%5D%29%29%2C+%22signal%22%2C+%22metrics%22%2C+%22%22%2C+%22%22%29%0Aor%0Alabel_replace%28sum%28rate%28otelcol_receiver_refused_log_records_total%5B5m%5D%29%29%2C+%22signal%22%2C+%22logs%22%2C+%22%22%2C+%22%22%29&g1.show_tree=0&g1.tab=graph&g1.range_input=1h&g1.res_type=auto&g1.res_density=medium&g1.display_mode=lines&g1.show_exemplars=0)
- [Span, metrics and logs per second by service](http://localhost:9090/query?g0.expr=sum+by+%28service_name%29+%28rate%28telemetry_spans_count_total%5B5m%5D%29%29&g0.show_tree=0&g0.tab=graph&g0.range_input=1h&g0.res_type=auto&g0.res_density=medium&g0.display_mode=lines&g0.show_exemplars=0&g1.expr=sum+by+%28service_name%29+%28rate%28telemetry_metrics_count_total%5B5m%5D%29%29&g1.show_tree=0&g1.tab=graph&g1.range_input=1h&g1.res_type=auto&g1.res_density=medium&g1.display_mode=lines&g1.show_exemplars=0&g2.expr=sum+by+%28service_name%29+%28rate%28telemetry_logs_count_total%5B5m%5D%29%29+&g2.show_tree=0&g2.tab=graph&g2.range_input=1h&g2.res_type=auto&g2.res_density=medium&g2.display_mode=lines&g2.show_exemplars=0)

## Jaeger remote sampling

Jaeger remote sampling allows SDKs to dynamically fetch sampling strategies from the OpenTelemetry Collector, enabling centralized per-service sampling configuration without redeploying applications.

How it works:
1. The collector serves sampling strategies via the [`jaegerremotesampling`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/jaegerremotesampling) extension
2. SDKs periodically poll the collector for their sampling configuration
3. When the configuration changes, SDKs pick up new rates without restart

Benefits:
- Centralized control - manage sampling rates for all services from one place
- Per-service and endpoint rates - critical services get higher sampling, noisy services get lower
- Dynamic updates - change rates without redeploying applications

### (Optional) Exercise: Enable Jaeger remote sampling
Change the [app/01-instrumentation.yaml](./app/01-instrumentation.yaml) and [app/03-collector-data-profiling.yaml](./app/03-collector-data-profiling.yaml).

How is it supported in the SDKs?
- backend1 (Python) - not supported
- backend2 (Java) - fully supported out of the box with the javaagent
- backend3 (Go) - [supported](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/samplers/jaegerremote), but requires manual SDK configuration (not via env var with operator injection)
- frontend (Node.js) - not supported

## Head sampling in the collector

Head sampling can also be done in the collector using the [probabilistic sampler processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/probabilisticsamplerprocessor). Unlike SDK sampling, the collector samples after spans are already created and exported by the application. It supports both **traces** and **logs**.

### Sampling modes

#### Hash Seed

Uses the FNV hash function on the Trace ID (or a specified attribute for logs) and compares against the sampling percentage. Uses 14 bits of randomness.

```yaml
processors:
  probabilistic_sampler:
    mode: hash_seed
    sampling_percentage: 15
    hash_seed: 42  # must be the same across all collectors in the same tier
```

- Traces: hashes the Trace ID
- Logs: can hash any attribute (useful when logs don't have a Trace ID)
- Best for: simple percentage-based sampling, especially for logs

#### Proportional

* Reduces items by a fixed ratio regardless of prior sampling decisions.

```yaml
processors:
  probabilistic_sampler:
    mode: proportional
    sampling_percentage: 25
```

#### Equalizing

* Items already sampled at a lower rate pass through; items sampled at a higher rate are further reduced.

```yaml
processors:
  probabilistic_sampler:
    mode: equalizing
    sampling_percentage: 10
```

- Best for: ensuring a uniform sampling rate across services with different SDK configurations

### Comparison
Practical difference with 10%:

Upstream SDK sampled at 50% → sends trace with th:8000...

* hash_seed:    Ignores th:, hashes TraceID → keeps 10% → effective 5%                                                                                                                                                                                                                                                                        
* proportional: Ignores th:, uses TraceID randomness → keeps 10% → effective 5%                                                                                                                                                                                                                                                               
* equalizing:   Reads th:, sees 50% > 10% target → reduces to 10% → effective 10%

Key takeaway: hash_seed and proportional behave similarly (both ignore upstream and write `th:` to tracestate), but hash_seed uses FNV hash with 14-bit precision while proportional uses raw TraceID bits with 56-bit precision.


### Exercise: Decrease traces and logs export rate by 50%

* Enable `probabilistic_sampler` in the collector [app/03-collector-data-profiling.yaml](./app/03-collector-data-profiling.yaml) for traces and logs.

```yaml
probabilistic_sampler:
  mode: equalizing
  sampling_percentage: 50
```

```bash
kubectl apply -f app/03-collector-data-profiling.yaml
```

![Spans per service](./images/p8s-collector-probabilistic-sampler.png)

- [Span, metrics and logs per second by service](http://localhost:9090/query?g0.expr=sum+by+%28service_name%29+%28rate%28telemetry_spans_count_total%5B5m%5D%29%29&g0.show_tree=0&g0.tab=graph&g0.range_input=1h&g0.res_type=auto&g0.res_density=medium&g0.display_mode=lines&g0.show_exemplars=0&g1.expr=sum+by+%28service_name%29+%28rate%28telemetry_metrics_count_total%5B5m%5D%29%29&g1.show_tree=0&g1.tab=graph&g1.range_input=1h&g1.res_type=auto&g1.res_density=medium&g1.display_mode=lines&g1.show_exemplars=0&g2.expr=sum+by+%28service_name%29+%28rate%28telemetry_logs_count_total%5B5m%5D%29%29+&g2.show_tree=0&g2.tab=graph&g2.range_input=1h&g2.res_type=auto&g2.res_density=medium&g2.display_mode=lines&g2.show_exemplars=0)

### When to use it

- You can't modify the application or SDK configuration (third-party services, legacy apps)
- You want centralized sampling control without touching each service's config
- As a safety net to enforce maximum ingestion rate even if SDKs are misconfigured

---

[Next steps](./06-tail-based-sampling.md)
