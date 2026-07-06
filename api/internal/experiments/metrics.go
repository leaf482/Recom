package experiments

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	metricsFieldRequests    = "recommendation_requests"
	metricsFieldImpressions = "impressions"
	maxLatencySamples       = 100
)

type StrategyMetrics struct {
	Strategy               string  `json:"strategy"`
	RecommendationRequests int64   `json:"recommendationRequests"`
	Impressions            int64   `json:"impressions"`
	AverageLatencyMs       float64 `json:"averageLatencyMs"`
	P95LatencyMs           float64 `json:"p95LatencyMs"`
}

type MetricsRecorder struct {
	redis *redis.Client
}

func NewMetricsRecorder(client *redis.Client) *MetricsRecorder {
	return &MetricsRecorder{redis: client}
}

func (m *MetricsRecorder) RecordRecommendation(
	ctx context.Context,
	experimentID, strategy string,
	impressionCount int,
	latencyMs float64,
) error {
	metricsKey := metricsKey(experimentID, strategy)
	latenciesKey := latenciesKey(experimentID, strategy)

	pipe := m.redis.Pipeline()
	pipe.HIncrBy(ctx, metricsKey, metricsFieldRequests, 1)
	pipe.HIncrBy(ctx, metricsKey, metricsFieldImpressions, int64(impressionCount))
	pipe.LPush(ctx, latenciesKey, formatLatency(latencyMs))
	pipe.LTrim(ctx, latenciesKey, 0, maxLatencySamples-1)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("record experiment metrics: %w", err)
	}

	return nil
}

func (m *MetricsRecorder) GetMetrics(ctx context.Context, experimentID string) ([]StrategyMetrics, error) {
	results := make([]StrategyMetrics, 0, len(Strategies))

	for _, strategy := range Strategies {
		metricsKey := metricsKey(experimentID, strategy)
		latenciesKey := latenciesKey(experimentID, strategy)

		rawMetrics, err := m.redis.HGetAll(ctx, metricsKey).Result()
		if err != nil {
			return nil, fmt.Errorf("load metrics for %s: %w", strategy, err)
		}

		rawLatencies, err := m.redis.LRange(ctx, latenciesKey, 0, -1).Result()
		if err != nil {
			return nil, fmt.Errorf("load latencies for %s: %w", strategy, err)
		}

		avgLatency, p95Latency := summarizeLatencies(rawLatencies)
		results = append(results, StrategyMetrics{
			Strategy:               strategy,
			RecommendationRequests: parseInt64(rawMetrics[metricsFieldRequests]),
			Impressions:            parseInt64(rawMetrics[metricsFieldImpressions]),
			AverageLatencyMs:       roundMetric(avgLatency),
			P95LatencyMs:           roundMetric(p95Latency),
		})
	}

	return results, nil
}

func metricsKey(experimentID, strategy string) string {
	return fmt.Sprintf("experiment:%s:strategy:%s:metrics", experimentID, strategy)
}

func latenciesKey(experimentID, strategy string) string {
	return fmt.Sprintf("experiment:%s:strategy:%s:latencies", experimentID, strategy)
}

func formatLatency(latencyMs float64) string {
	return strconv.FormatFloat(latencyMs, 'f', 2, 64)
}

func parseInt64(raw string) int64 {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func summarizeLatencies(rawValues []string) (float64, float64) {
	if len(rawValues) == 0 {
		return 0, 0
	}

	values := make([]float64, 0, len(rawValues))
	var total float64
	for _, raw := range rawValues {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}
		values = append(values, value)
		total += value
	}

	if len(values) == 0 {
		return 0, 0
	}

	sort.Float64s(values)
	p95Index := int(math.Ceil(float64(len(values))*0.95)) - 1
	if p95Index < 0 {
		p95Index = 0
	}
	if p95Index >= len(values) {
		p95Index = len(values) - 1
	}

	return total / float64(len(values)), values[p95Index]
}

func roundMetric(value float64) float64 {
	return math.Round(value*10) / 10
}
