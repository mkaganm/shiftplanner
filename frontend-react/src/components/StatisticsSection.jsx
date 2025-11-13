// Statistics Section Component
import { useApp } from '../context/AppContext';
import './StatisticsSection.css';

export function StatisticsSection() {
  const { stats, loading } = useApp();

  return (
    <section className="section">
      <h2>Statistics</h2>
      <div className="stats-list">
        {loading ? (
          <div className="empty-state">Loading...</div>
        ) : stats.length === 0 ? (
          <div className="empty-state">No statistics yet</div>
        ) : (
          stats.map((stat) => (
            <div key={stat.member_id} className="stat-card">
              <h3>{stat.member_name}</h3>
              <div className="stat-item">
                <span>Total Shift Days:</span>
                <strong>{stat.total_days}</strong>
              </div>
              <div className="stat-item">
                <span>Long Shift Count:</span>
                <strong>{stat.long_shift_count}</strong>
              </div>
            </div>
          ))
        )}
      </div>
    </section>
  );
}

