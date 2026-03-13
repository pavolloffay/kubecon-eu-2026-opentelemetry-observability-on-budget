# Cleaning up logs

We have already looked at identifying and filtering large logs. In this chapter we will look at log deduplication.

## Logs deduplication

Duplicate logs inflate storage costs and make analysis harder. Common causes:
- Retry logic logging the same error multiple times
- Application bugs producing identical log lines in loops

### Using the logdedup processor

The [logdedup processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logdedupprocessor) aggregates identical logs within a time window and exports a single log with a count.

```yaml
processors:
  logdedup:
    interval: 10s              # aggregation window
    log_count_attribute: log_count  # attribute name for duplicate count
    timezone: UTC
    conditions:
      - 'body == body'         # dedupe logs with same body (OTTL condition)
```

### Configuration example

```yaml
processors:
  logdedup:
    interval: 30s
    log_count_attribute: duplicate_count
    exclude:
      # Don't dedupe logs with trace context (they're unique per request)
      - 'trace_id != nil'

service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [logdedup, batch]
      exporters: [otlp]
```

### How it works

1. Logs are grouped by their body content (and optionally attributes)
2. During the interval, duplicates are counted but not exported
3. At the end of each interval, one log per group is exported with `duplicate_count` attribute
4. Logs with unique content pass through immediately

### Trade-offs

- Latency: Logs are delayed by the aggregation interval
- Memory: Processor buffers unique logs during the interval
- Loss of timing precision: Individual timestamps are lost for duplicates

