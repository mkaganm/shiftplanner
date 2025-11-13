// Application State Context
import { createContext, useContext, useState, useEffect } from 'react';
import { membersAPI, shiftsAPI, holidaysAPI, statsAPI } from '../services/api';

const AppContext = createContext(null);

export function AppProvider({ children }) {
  const [members, setMembers] = useState([]);
  const [shifts, setShifts] = useState([]);
  const [holidays, setHolidays] = useState({});
  const [stats, setStats] = useState([]);
  const [currentMonth, setCurrentMonth] = useState(new Date().getMonth());
  const [currentYear, setCurrentYear] = useState(new Date().getFullYear());
  const [loading, setLoading] = useState(false);

  // Load holidays on mount
  useEffect(() => {
    loadHolidays();
  }, []);

  const loadHolidays = async () => {
    try {
      const data = await holidaysAPI.getAll();
      setHolidays(data || {});
    } catch (error) {
      console.error('Error loading holidays:', error);
    }
  };

  const loadMembers = async () => {
    try {
      setLoading(true);
      const data = await membersAPI.getAll();
      setMembers(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error loading members:', error);
      setMembers([]);
    } finally {
      setLoading(false);
    }
  };

  const loadShifts = async (startDate, endDate) => {
    try {
      setLoading(true);
      const data = await shiftsAPI.get(startDate, endDate);
      console.log('Loaded shifts:', data);
      setShifts(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error loading shifts:', error);
      setShifts([]);
    } finally {
      setLoading(false);
    }
  };

  const loadStats = async () => {
    try {
      setLoading(true);
      const data = await statsAPI.getAll();
      setStats(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error loading stats:', error);
      setStats([]);
    } finally {
      setLoading(false);
    }
  };

  const createMember = async (name) => {
    try {
      await membersAPI.create(name);
      await loadMembers();
      await loadStats();
    } catch (error) {
      throw error;
    }
  };

  const deleteMember = async (id) => {
    try {
      await membersAPI.delete(id);
      await loadMembers();
      await loadStats();
      await loadShiftsForCurrentMonth();
    } catch (error) {
      throw error;
    }
  };

  const generateShifts = async (startDate, endDate) => {
    try {
      const data = await shiftsAPI.generate(startDate, endDate);
      await loadShiftsForCurrentMonth();
      await loadStats();
      return data;
    } catch (error) {
      throw error;
    }
  };

  const loadShiftsForCurrentMonth = async () => {
    const start = new Date(currentYear, currentMonth, 1);
    const end = new Date(currentYear, currentMonth + 1, 0);
    const startDate = formatDate(start);
    const endDate = formatDate(end);
    await loadShifts(startDate, endDate);
  };

  const formatDate = (date) => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  const previousMonth = () => {
    let newMonth = currentMonth - 1;
    let newYear = currentYear;
    if (newMonth < 0) {
      newMonth = 11;
      newYear--;
    }
    setCurrentMonth(newMonth);
    setCurrentYear(newYear);
  };

  const nextMonth = () => {
    let newMonth = currentMonth + 1;
    let newYear = currentYear;
    if (newMonth > 11) {
      newMonth = 0;
      newYear++;
    }
    setCurrentMonth(newMonth);
    setCurrentYear(newYear);
  };

  // Load shifts when month changes
  useEffect(() => {
    loadShiftsForCurrentMonth();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentMonth, currentYear]);

  const value = {
    members,
    shifts,
    holidays,
    stats,
    currentMonth,
    currentYear,
    loading,
    loadMembers,
    loadShifts,
    loadStats,
    createMember,
    deleteMember,
    generateShifts,
    previousMonth,
    nextMonth,
    formatDate
  };

  return (
    <AppContext.Provider value={value}>
      {children}
    </AppContext.Provider>
  );
}

export function useApp() {
  const context = useContext(AppContext);
  if (!context) {
    throw new Error('useApp must be used within AppProvider');
  }
  return context;
}

