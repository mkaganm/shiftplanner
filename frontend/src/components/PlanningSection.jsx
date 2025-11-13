// Planning Section Component
import { useState, useEffect } from 'react';
import { useApp } from '../context/AppContext';
import './PlanningSection.css';

export function PlanningSection() {
  const { generateShifts, clearAllShifts, loading } = useApp();
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    const today = new Date().toISOString().split('T')[0];
    setStartDate(today);
    const nextMonth = new Date();
    nextMonth.setMonth(nextMonth.getMonth() + 1);
    setEndDate(nextMonth.toISOString().split('T')[0]);
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!startDate || !endDate) {
      setError('Please select start and end dates');
      return;
    }

    if (new Date(startDate) > new Date(endDate)) {
      setError('Start date cannot be after end date');
      return;
    }

    try {
      const shifts = await generateShifts(startDate, endDate);
      setSuccess(`${shifts.length} shift plans created successfully`);
      setTimeout(() => setSuccess(''), 5000);
    } catch (err) {
      setError(err.message || 'Error creating plan');
    }
  };

  const handleClearAll = async () => {
    if (!window.confirm('Are you sure you want to delete all shifts? This action cannot be undone.')) {
      return;
    }

    setError('');
    setSuccess('');

    try {
      await clearAllShifts();
      setSuccess('All shifts cleared successfully');
      setTimeout(() => setSuccess(''), 5000);
    } catch (err) {
      setError(err.message || 'Error clearing shifts');
    }
  };

  return (
    <section className="section">
      <h2>Create Shift Plan</h2>
      <form onSubmit={handleSubmit} className="plan-form">
        <label>
          Start Date:
          <input
            type="date"
            value={startDate}
            onChange={(e) => setStartDate(e.target.value)}
          />
        </label>
        <label>
          End Date:
          <input
            type="date"
            value={endDate}
            onChange={(e) => setEndDate(e.target.value)}
          />
        </label>
        <button type="submit" disabled={loading}>
          Create Plan
        </button>
      </form>
      <div className="clear-section">
        <button 
          type="button" 
          onClick={handleClearAll} 
          disabled={loading}
          className="clear-button"
        >
          Clear All Shifts
        </button>
      </div>
      {error && <div className="error-message">{error}</div>}
      {success && <div className="message success">{success}</div>}
    </section>
  );
}

