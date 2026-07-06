export type HealthResponse = {
  status: string;
};

export type UserFeatures = {
  userId: string;
  genreScores: Record<string, number>;
  artistScores: Record<string, number>;
  recentTracks: string[];
  eventCounts: Record<string, number>;
};

export type Recommendation = {
  trackId: string;
  title: string;
  artistId: string;
  artistName: string;
  genre: string;
  score: number;
  reason: string;
};

export type RecommendationsResponse = {
  userId: string;
  strategy: string;
  recommendations: Recommendation[];
};

export type StrategyMetrics = {
  strategy: string;
  recommendationRequests: number;
  impressions: number;
  averageLatencyMs: number;
  p95LatencyMs: number;
};

export type ExperimentMetricsResponse = {
  experimentId: string;
  strategies: StrategyMetrics[];
};

const API_BASE = "/backend";

async function fetchJSON<T>(path: string): Promise<T> {
  const response = await fetch(path, { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status} ${response.statusText}`);
  }
  return response.json() as Promise<T>;
}

export function getHealth() {
  return fetchJSON<HealthResponse>(`${API_BASE}/health`);
}

export function getUserFeatures(userId: string) {
  return fetchJSON<UserFeatures>(`${API_BASE}/users/${encodeURIComponent(userId)}/features`);
}

export function getUserRecommendations(userId: string) {
  return fetchJSON<RecommendationsResponse>(
    `${API_BASE}/users/${encodeURIComponent(userId)}/recommendations`
  );
}

export function getExperimentMetrics(experimentId = "default") {
  return fetchJSON<ExperimentMetricsResponse>(
    `${API_BASE}/experiments/${encodeURIComponent(experimentId)}/metrics`
  );
}
