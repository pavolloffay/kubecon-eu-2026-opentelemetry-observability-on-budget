# Cleaning up logs

We have already looked at identifying and filtering large logs. In this chapter we will look at log deduplication.

## Logs deduplication

Duplicate logs inflate storage costs and make analysis harder. Common causes:
- Retry logic logging the same error multiple times
- Application bugs producing identical log lines in loops

### Using the logdedup processor

The [logdedup processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logdedupprocessor) aggregates identical logs within a time window and exports a single log with a count.

1. Logs are grouped by their body content (and optionally attributes)
1. During the interval, duplicates are counted but not exported
1. At the end of each interval, one log per group is exported with `duplicate_count` attribute
1. Logs with unique content pass through immediately

```yaml
processors:
  logdedup:
    interval: 10s              # aggregation window
    log_count_attribute: log_count  # attribute name for duplicate count
    timezone: UTC
    conditions:
      - 'body == body'         # dedupe logs with same body (OTTL condition)
```

### Monitoring deduplication

The logdedup processor exposes metrics to track its effectiveness:

```promql
# Logs aggregated (input to deduplication)
otelcol_processor_logdedup_aggregated_logs_total

# Logs exported (output after deduplication)
otelcol_processor_logdedup_exported_logs_total

# Deduplication ratio
1 - (rate(otelcol_processor_logdedup_exported_logs_total[5m]) / rate(otelcol_processor_logdedup_aggregated_logs_total[5m]))
```

- [Logdedup processor metrics](http://localhost:9090/query?g0.expr=rate%28otelcol_processor_logdedup_aggregated_logs_total%5B5m%5D%29&g0.show_tree=0&g0.tab=graph&g0.range_input=1h&g1.expr=rate%28otelcol_processor_logdedup_exported_logs_total%5B5m%5D%29&g1.show_tree=0&g1.tab=graph&g1.range_input=1h&g2.expr=1+-+%28rate%28otelcol_processor_logdedup_exported_logs_total%5B5m%5D%29+%2F+rate%28otelcol_processor_logdedup_aggregated_logs_total%5B5m%5D%29%29&g2.show_tree=0&g2.tab=graph&g2.range_input=1h)

### Trade-offs

- Latency: Logs are delayed by the aggregation interval
- Memory: Processor buffers unique logs during the interval
- Loss of timing precision: Individual timestamps are lost for duplicates

