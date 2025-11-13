// Dashboard Page
import { useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { useApp } from '../context/AppContext';
import { MembersSection } from '../components/MembersSection';
import { StatisticsSection } from '../components/StatisticsSection';
import { PlanningSection } from '../components/PlanningSection';
import { CalendarSection } from '../components/CalendarSection';
import './Dashboard.css';

export function Dashboard() {
  const { logout } = useAuth();
  const { loadMembers, loadStats } = useApp();

  useEffect(() => {
    loadMembers();
    loadStats();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="container">
      <header>
        <h1>Shift Planning System</h1>
        <button onClick={logout} className="logout-btn">
          Logout
        </button>
      </header>

      <div className="main-content">
        <MembersSection />
        <StatisticsSection />
        <PlanningSection />
        <CalendarSection />
      </div>
    </div>
  );
}

