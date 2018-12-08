package main

import (
	"context"
	"google.golang.org/genproto/googleapis/api/label"
	"time"

	"cloud.google.com/go/monitoring/apiv3"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/tools/benchmark/parse"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

var (
	allocatedBytesPerOp = &metricpb.MetricDescriptor{
		Name: "Allocated Bytes Per Operation",
		Type: "custom.googleapis.com/benchmark/allocated_bytes_per_op",
		Labels: []*label.LabelDescriptor{
			{
				Key:         "branch",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Git branch",
			},
			{
				Key:         "githash",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Git hash",
			},
			{
				Key:         "version",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Release Version",
			},
			{
				Key:         "name",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Benchmark name",
			},
		},
		MetricKind:  metricpb.MetricDescriptor_GAUGE,
		ValueType:   metricpb.MetricDescriptor_INT64,
		Unit:        "By",
		Description: "Allocated Bytes Per Operation",
		DisplayName: "Benchmark Allocated Bytes Per Operation",
	}
	allocationsPerOp = &metricpb.MetricDescriptor{
		Name: "Allocations Per Operation",
		Type: "custom.googleapis.com/benchmark/allocations_per_op",
		Labels: []*label.LabelDescriptor{
			{
				Key:         "branch",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Git branch",
			},
			{
				Key:         "githash",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Git hash",
			},
			{
				Key:         "version",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Release Version",
			},
			{
				Key:         "name",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Benchmark name",
			},
		},
		MetricKind:  metricpb.MetricDescriptor_GAUGE,
		ValueType:   metricpb.MetricDescriptor_INT64,
		Description: "Allocations Per Operation",
		DisplayName: "Benchmark Allocations Per Operation",
	}
	nanosecondsPerOp = &metricpb.MetricDescriptor{
		Name: "Nanoseconds Per Operation",
		Type: "custom.googleapis.com/benchmark/ns_per_op",
		Labels: []*label.LabelDescriptor{
			{
				Key:         "branch",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Git branch",
			},
			{
				Key:         "githash",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Git hash",
			},
			{
				Key:         "version",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Release Version",
			},
			{
				Key:         "name",
				ValueType:   label.LabelDescriptor_STRING,
				Description: "Benchmark name",
			},
		},
		MetricKind:  metricpb.MetricDescriptor_GAUGE,
		ValueType:   metricpb.MetricDescriptor_DOUBLE,
		Unit:        "ns",
		Description: "Nanoseconds Per Operation",
		DisplayName: "Benchmark Nanoseconds Per Operation",
	}
)

func NewStackDriverClient(ctx context.Context, projectID string) (Reporter, error) {
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "creating metric client")
	}
	return &stackdriver{
		client:    client,
		projectID: projectID,
	}, nil
}

type stackdriver struct {
	client    *monitoring.MetricClient
	projectID string
}

func (s *stackdriver) Upload(ctx context.Context, set parse.Set, cfg *Config) error {
	err := s.createMetricDescriptors(ctx)
	if err != nil {
		return errors.Wrap(err, "creating metric descriptors")
	}

	var timeseries []*monitoringpb.TimeSeries
	now := time.Now()

	for _, benchmarks := range set {
		for _, benchmark := range benchmarks {
			timeseries = append(timeseries, s.benchmarkTimeseries(ctx, benchmark, cfg, now)...)
		}
	}

	if err := s.client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
		Name:       monitoring.MetricProjectPath(s.projectID),
		TimeSeries: timeseries,
	}); err != nil {
		return errors.Wrap(err, "Failed to write time series data")
	}

	return nil
}

func (s *stackdriver) createMetricDescriptors(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	metricDescriptors := []*metricpb.MetricDescriptor{allocatedBytesPerOp, allocationsPerOp, nanosecondsPerOp}
	for _, descriptor := range metricDescriptors {
		errChan := make(chan error)
		go func() {
			_, err := s.client.GetMetricDescriptor(ctx, &monitoringpb.GetMetricDescriptorRequest{
				Name: monitoring.MetricProjectPath(s.projectID) + descriptor.Type,
			})
			errChan <- err
		}()

		var err error
		select {
		case err = <-errChan:
		case <-time.After(30 * time.Second):
			return errors.New("timed out getting metric descriptor")
		case <-ctx.Done():
			return ctx.Err()
		}
		if err == nil {
			continue
		}

		go func() {
			// TODO: check error type
			_, err = s.client.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
				Name:             monitoring.MetricProjectPath(s.projectID),
				MetricDescriptor: descriptor,
			})
			errChan <- err
		}()

		select {
		case err = <-errChan:
		case <-time.After(30 * time.Second):
			return errors.New("timed out creating metric descriptor")
		case <-ctx.Done():
			return ctx.Err()
		}
		if err != nil {
			return errors.Wrap(err, "creating metric descriptor")
		}
	}
	return nil
}

func (s *stackdriver) benchmarkTimeseries(ctx context.Context, benchmark *parse.Benchmark, cfg *Config, t time.Time) []*monitoringpb.TimeSeries {
	labels := map[string]string{
		"branch":  cfg.Branch,
		"githash": cfg.Githash,
		"version": cfg.Version,
		"name":    benchmark.Name,
	}

	resource := &monitoredrespb.MonitoredResource{
		Type: "global",
		Labels: map[string]string{
			"project_id": s.projectID,
		},
	}

	return []*monitoringpb.TimeSeries{
		{
			Metric: &metricpb.Metric{
				Type:   nanosecondsPerOp.Type,
				Labels: labels,
			},
			Resource: resource,
			Points: []*monitoringpb.Point{
				{
					Interval: &monitoringpb.TimeInterval{
						EndTime: &googlepb.Timestamp{
							Seconds: t.Unix(),
						},
					},
					Value: &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_DoubleValue{
							DoubleValue: benchmark.NsPerOp,
						},
					},
				},
			},
		},
		{
			Metric: &metricpb.Metric{
				Type:   allocatedBytesPerOp.Type,
				Labels: labels,
			},
			Resource: resource,
			Points: []*monitoringpb.Point{
				{
					Interval: &monitoringpb.TimeInterval{
						EndTime: &googlepb.Timestamp{
							Seconds: t.Unix(),
						},
					},
					Value: &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_Int64Value{
							Int64Value: int64(benchmark.AllocedBytesPerOp),
						},
					},
				},
			},
		},
		{
			Metric: &metricpb.Metric{
				Type:   allocationsPerOp.Type,
				Labels: labels,
			},
			Resource: resource,
			Points: []*monitoringpb.Point{
				{
					Interval: &monitoringpb.TimeInterval{
						EndTime: &googlepb.Timestamp{
							Seconds: t.Unix(),
						},
					},
					Value: &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_Int64Value{
							Int64Value: int64(benchmark.AllocsPerOp),
						},
					},
				},
			},
		},
	}
}

func (s *stackdriver) Close() error {
	return s.client.Close()
}
