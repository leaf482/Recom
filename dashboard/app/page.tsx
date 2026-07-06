"use client";

import { useCallback, useEffect, useState } from "react";
import {
  ExperimentMetricsResponse,
  getExperimentMetrics,
  getHealth,
  getUserFeatures,
  getUserRecommendations,
  RecommendationsResponse,
  UserFeatures,
} from "@/lib/api";

const PRESET_USERS = ["user_1", "user_2", "user_3", "user_4", "user_5"];
const AUTO_REFRESH_MS = 5000;

function ScoreTable({
  title,
  scores,
}: {
  title: string;
  scores: Record<string, number>;
}) {
  const entries = Object.entries(scores).sort((a, b) => b[1] - a[1]);

  if (entries.length === 0) {
    return (
      <div>
        <h3>{title}</h3>
        <p className="muted">No data yet.</p>
      </div>
    );
  }

  return (
    <div>
      <h3>{title}</h3>
      <table className="table">
        <thead>
          <tr>
            <th>Key</th>
            <th>Score</th>
          </tr>
        </thead>
        <tbody>
          {entries.map(([key, value]) => (
            <tr key={key}>
              <td>{key}</td>
              <td>{value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default function DashboardPage() {
  const [selectedUser, setSelectedUser] = useState("user_1");
  const [customUser, setCustomUser] = useState("");
  const [healthStatus, setHealthStatus] = useState<string>("checking...");
  const [apiConnected, setApiConnected] = useState<boolean | null>(null);
  const [features, setFeatures] = useState<UserFeatures | null>(null);
  const [recommendations, setRecommendations] = useState<RecommendationsResponse | null>(null);
  const [metrics, setMetrics] = useState<ExperimentMetricsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const activeUser = customUser.trim() || selectedUser;

  const loadDashboard = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const [health, userFeatures, userRecommendations, experimentMetrics] = await Promise.all([
        getHealth(),
        getUserFeatures(activeUser),
        getUserRecommendations(activeUser),
        getExperimentMetrics("default"),
      ]);

      setHealthStatus(health.status);
      setApiConnected(true);
      setFeatures(userFeatures);
      setRecommendations(userRecommendations);
      setMetrics(experimentMetrics);
    } catch (err) {
      setApiConnected(false);
      setHealthStatus("unreachable");
      setError(err instanceof Error ? err.message : "Failed to load dashboard data");
    } finally {
      setLoading(false);
    }
  }, [activeUser]);

  useEffect(() => {
    void loadDashboard();
  }, [loadDashboard]);

  useEffect(() => {
    const timer = window.setInterval(() => {
      void loadDashboard();
    }, AUTO_REFRESH_MS);

    return () => window.clearInterval(timer);
  }, [loadDashboard]);

  return (
    <main className="page">
      <header className="header">
        <h1>EchoRec Dashboard</h1>
        <p>Real-time music recommendation platform — interview-friendly system view</p>
      </header>

      <div className="toolbar">
        <div className="user-buttons">
          {PRESET_USERS.map((userId) => (
            <button
              key={userId}
              className={`btn ${selectedUser === userId && !customUser.trim() ? "active" : ""}`}
              onClick={() => {
                setSelectedUser(userId);
                setCustomUser("");
              }}
            >
              {userId}
            </button>
          ))}
        </div>
        <input
          className="input"
          placeholder="Custom userId"
          value={customUser}
          onChange={(event) => setCustomUser(event.target.value)}
        />
        <button className="btn primary" onClick={() => void loadDashboard()} disabled={loading}>
          {loading ? "Refreshing..." : "Refresh"}
        </button>
        <span className="muted">Auto-refresh every {AUTO_REFRESH_MS / 1000}s</span>
      </div>

      {error ? <div className="error-box">{error}</div> : null}

      <div className="grid">
        <section className="panel">
          <h2>System Status</h2>
          <p>
            API health:{" "}
            <span className={apiConnected ? "status-ok" : "status-error"}>{healthStatus}</span>
          </p>
          <p className="muted">
            Connection: {apiConnected === null ? "checking" : apiConnected ? "connected" : "failed"}
          </p>
          <p className="muted">Active user: {activeUser}</p>
        </section>

        <section className="panel full-width">
          <h2>Architecture</h2>
          <div className="architecture">
            simulator → Redpanda → consumer → Redis → Recommendation API → Dashboard
          </div>
        </section>

        <section className="panel">
          <h2>User Feature Profile</h2>
          {features ? (
            <>
              <ScoreTable title="Genre Scores" scores={features.genreScores} />
              <ScoreTable title="Artist Scores" scores={features.artistScores} />
              <h3>Recent Tracks</h3>
              {features.recentTracks.length > 0 ? (
                <div className="chip-list">
                  {features.recentTracks.map((trackId) => (
                    <span key={trackId} className="chip">
                      {trackId}
                    </span>
                  ))}
                </div>
              ) : (
                <p className="muted">No recent tracks yet.</p>
              )}
              <ScoreTable title="Event Counts" scores={features.eventCounts} />
            </>
          ) : (
            <p className="muted">{loading ? "Loading features..." : "No feature data."}</p>
          )}
        </section>

        <section className="panel">
          <h2>Recommendations</h2>
          {recommendations ? (
            <>
              <div className="strategy">Strategy: {recommendations.strategy}</div>
              {recommendations.recommendations.length > 0 ? (
                <table className="table">
                  <thead>
                    <tr>
                      <th>Track</th>
                      <th>Artist</th>
                      <th>Genre</th>
                      <th>Score</th>
                      <th>Reason</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recommendations.recommendations.map((item) => (
                      <tr key={item.trackId}>
                        <td>
                          <strong>{item.title}</strong>
                          <div className="muted">{item.trackId}</div>
                        </td>
                        <td>{item.artistName}</td>
                        <td>{item.genre}</td>
                        <td className="score">{item.score.toFixed(2)}</td>
                        <td>{item.reason}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <p className="muted">No recommendations returned.</p>
              )}
            </>
          ) : (
            <p className="muted">{loading ? "Loading recommendations..." : "No recommendations."}</p>
          )}
        </section>

        <section className="panel full-width">
          <h2>Experiment Metrics</h2>
          {metrics ? (
            <table className="table">
              <thead>
                <tr>
                  <th>Strategy</th>
                  <th>Requests</th>
                  <th>Impressions</th>
                  <th>Avg Latency (ms)</th>
                  <th>P95 Latency (ms)</th>
                </tr>
              </thead>
              <tbody>
                {metrics.strategies.map((strategy) => (
                  <tr key={strategy.strategy}>
                    <td>{strategy.strategy}</td>
                    <td>{strategy.recommendationRequests}</td>
                    <td>{strategy.impressions}</td>
                    <td>{strategy.averageLatencyMs.toFixed(1)}</td>
                    <td>{strategy.p95LatencyMs.toFixed(1)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <p className="muted">{loading ? "Loading metrics..." : "No metrics yet."}</p>
          )}
        </section>
      </div>
    </main>
  );
}
