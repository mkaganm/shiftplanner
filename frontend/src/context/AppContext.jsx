// Application State Context
import { createContext, useContext, useState, useEffect } from 'react';
import { membersAPI, shiftsAPI, holidaysAPI, statsAPI, leaveDaysAPI } from '../services/api';

const AppContext = createContext(null);

export function AppProvider({ children }) {
  const [members, setMembers] = useState([]);
  const [shifts, setShifts] = useState([]);
  const [holidays, setHolidays] = useState({});
  const [stats, setStats] = useState([]);
  const [leaveDays, setLeaveDays] = useState([]);
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
      // Reload leave days to remove deleted member's leave days
      const start = new Date(currentYear, currentMonth, 1);
      const end = new Date(currentYear, currentMonth + 1, 0);
      const startDate = formatDate(start);
      const endDate = formatDate(end);
      await loadLeaveDays(null, startDate, endDate);
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

  const clearAllShifts = async () => {
    try {
      await shiftsAPI.clearAll();
      await loadShiftsForCurrentMonth();
      await loadStats();
    } catch (error) {
      throw error;
    }
  };

  const updateShiftForDate = async (date, memberId) => {
    try {
      await shiftsAPI.updateForDate(date, memberId);
      await loadShiftsForCurrentMonth();
      await loadStats();
    } catch (error) {
      throw error;
    }
  };

  const importShifts = async (file) => {
    try {
      const result = await shiftsAPI.import(file);
      await loadMembers();
      await loadShiftsForCurrentMonth();
      await loadStats();
      return result;
    } catch (error) {
      throw error;
    }
  };

  const loadLeaveDays = async (memberId, startDate, endDate) => {
    try {
      setLoading(true);
      const data = await leaveDaysAPI.getAll(memberId, startDate, endDate);
      setLeaveDays(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Error loading leave days:', error);
      setLeaveDays([]);
    } finally {
      setLoading(false);
    }
  };

  const createLeaveDay = async (memberId, startDate, endDate) => {
    try {
      await leaveDaysAPI.create(memberId, startDate, endDate);
      // Reload all leave days
      await loadLeaveDays();
    } catch (error) {
      throw error;
    }
  };

  const deleteLeaveDay = async (id) => {
    try {
      await leaveDaysAPI.delete(id);
      // Reload all leave days
      await loadLeaveDays();
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

  // Load shifts and leave days when month changes
  useEffect(() => {
    loadShiftsForCurrentMonth();
    // Load leave days for current month
    const start = new Date(currentYear, currentMonth, 1);
    const end = new Date(currentYear, currentMonth + 1, 0);
    const startDate = formatDate(start);
    const endDate = formatDate(end);
    loadLeaveDays(null, startDate, endDate);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentMonth, currentYear]);

  const value = {
    members,
    shifts,
    holidays,
    stats,
    leaveDays,
    currentMonth,
    currentYear,
    loading,
    loadMembers,
    loadShifts,
    loadStats,
    loadLeaveDays,
    createMember,
    deleteMember,
    generateShifts,
    clearAllShifts,
    createLeaveDay,
    deleteLeaveDay,
    updateShiftForDate,
    importShifts,
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

