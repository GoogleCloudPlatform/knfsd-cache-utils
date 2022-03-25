# gen-overrides

Generate a list of overrides to enable/disable individual metrics.

This is intended to be executed directly from the `knfsd-metrics-agent` directory.

```sh
go run ./cmd/gen-overrides/
```

This will output YAML comments that can be pasted into the `common.yaml` file.

If a receiver only contains a single metric *do not* include the metric in the overrides. It wouldn't make any sense as its better to completely remove the receiver from the pipeline. Otherwise you're still paying the CPU cost to scrape the metric, but never exporting it.
