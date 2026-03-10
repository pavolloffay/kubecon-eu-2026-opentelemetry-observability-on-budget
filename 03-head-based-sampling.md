= Head based sampling

The sampling decision is made at the start of the trace (the "head") and propagated to all downstream services via trace context. The first service in the request chain decides whether to sample, and all downstream services honor that decision.

== How it works

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


## Head sampling in the SDK

## Head sampling in the collector
