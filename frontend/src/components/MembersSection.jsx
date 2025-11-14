// Members Section Component
import { useState, useEffect } from 'react';
import { useApp } from '../context/AppContext';
import './MembersSection.css';

export function MembersSection() {
  const { members, createMember, deleteMember, leaveDays, loadLeaveDays, createLeaveDay, deleteLeaveDay, loading, formatDate } = useApp();
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [leaveStartDate, setLeaveStartDate] = useState('');
  const [leaveEndDate, setLeaveEndDate] = useState('');
  const [leaveError, setLeaveError] = useState('');
  const [showLeaveModal, setShowLeaveModal] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (!name.trim()) {
      setError('Please enter member name');
      return;
    }

    try {
      await createMember(name.trim());
      setName('');
    } catch (err) {
      setError(err.message || 'Error adding member');
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this member?')) {
      return;
    }

    try {
      await deleteMember(id);
    } catch (err) {
      alert(err.message || 'Error deleting member');
    }
  };

  useEffect(() => {
    loadLeaveDays();
  }, []);


  const openLeaveModal = (memberId) => {
    setShowLeaveModal(memberId);
    loadLeaveDays(memberId);
    setLeaveStartDate('');
    setLeaveEndDate('');
    setLeaveError('');
  };

  const closeLeaveModal = () => {
    setShowLeaveModal(null);
    setLeaveStartDate('');
    setLeaveEndDate('');
    setLeaveError('');
  };

  const handleAddLeaveDay = async (e, memberId) => {
    e.preventDefault();
    setLeaveError('');

    if (!leaveStartDate.trim()) {
      setLeaveError('Please select start date');
      return;
    }

    if (!leaveEndDate.trim()) {
      setLeaveError('Please select end date');
      return;
    }

    const startDate = new Date(leaveStartDate);
    const endDate = new Date(leaveEndDate);

    if (startDate > endDate) {
      setLeaveError('Start date must be before or equal to end date');
      return;
    }

    try {
      await createLeaveDay(memberId, leaveStartDate, leaveEndDate);
      setLeaveStartDate('');
      setLeaveEndDate('');
      await loadLeaveDays(memberId);
    } catch (err) {
      setLeaveError(err.message || 'Error adding leave days');
    }
  };

  const handleDeleteLeaveDayWithRefresh = async (leaveDayId, memberId) => {
    if (!confirm('Are you sure you want to delete this leave day?')) {
      return;
    }

    try {
      await deleteLeaveDay(leaveDayId);
      await loadLeaveDays(memberId);
    } catch (err) {
      alert(err.message || 'Error deleting leave day');
    }
  };

  const getMemberLeaveDays = (memberId) => {
    return leaveDays.filter(ld => ld.member_id === memberId);
  };

  return (
    <section className="section">
      <h2>Team Members</h2>
      <form onSubmit={handleSubmit} className="member-form">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Enter member name"
        />
        <button type="submit" disabled={loading}>
          Add Member
        </button>
      </form>
      {error && <div className="error-message">{error}</div>}
      <div className="members-list">
        {members.length === 0 ? (
          <div className="empty-state">No members added yet</div>
        ) : (
          members.map((member) => {
            const memberLeaveDays = getMemberLeaveDays(member.id);
            return (
              <div key={member.id} className="member-item">
                <div className="member-header">
                  <span>{member.name}</span>
                  <div className="member-actions">
                    <button
                      className="leave-days-btn"
                      onClick={() => openLeaveModal(member.id)}
                      disabled={loading}
                      title="Manage leave days"
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                        <line x1="16" y1="2" x2="16" y2="6"></line>
                        <line x1="8" y1="2" x2="8" y2="6"></line>
                        <line x1="3" y1="10" x2="21" y2="10"></line>
                      </svg>
                      <span className="leave-count">{memberLeaveDays.length}</span>
                    </button>
                    <button
                      className="delete-btn"
                      onClick={() => handleDelete(member.id)}
                      disabled={loading}
                      title="Delete member"
                    >
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <polyline points="3 6 5 6 21 6"></polyline>
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Leave Days Modal */}
      {showLeaveModal && (
        <div className="modal-overlay" onClick={closeLeaveModal}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div>
                <h3>Leave Management</h3>
                <p style={{ margin: '4px 0 0 0', fontSize: '14px', color: '#7f8c8d', fontWeight: 'normal' }}>
                  {members.find(m => m.id === showLeaveModal)?.name}
                </p>
              </div>
              <button className="modal-close" onClick={closeLeaveModal} title="Close">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </button>
            </div>
            
            <div className="modal-body">
              <form onSubmit={(e) => handleAddLeaveDay(e, showLeaveModal)} className="leave-day-form-modal">
                <div className="form-row">
                  <div className="form-group">
                    <label htmlFor="leave-start-date">Start Date</label>
                    <input
                      id="leave-start-date"
                      type="date"
                      value={leaveStartDate}
                      onChange={(e) => setLeaveStartDate(e.target.value)}
                      placeholder="Select start date"
                    />
                  </div>
                  <div className="form-group">
                    <label htmlFor="leave-end-date">End Date</label>
                    <input
                      id="leave-end-date"
                      type="date"
                      value={leaveEndDate}
                      onChange={(e) => setLeaveEndDate(e.target.value)}
                      placeholder="Select end date"
                      min={leaveStartDate}
                    />
                  </div>
                </div>
                <button type="submit" className="add-leave-btn" disabled={loading}>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                    <line x1="12" y1="5" x2="12" y2="19"></line>
                    <line x1="5" y1="12" x2="19" y2="12"></line>
                  </svg>
                  Add Leave Days
                </button>
              </form>
              {leaveError && <div className="error-message" style={{ marginTop: '16px', padding: '12px', background: '#fee', border: '1px solid #fcc', borderRadius: '6px', color: '#c33' }}>{leaveError}</div>}
              
              <div className="leave-days-list-modal">
                <h4>Registered Leave Days ({getMemberLeaveDays(showLeaveModal).length})</h4>
                {getMemberLeaveDays(showLeaveModal).length === 0 ? (
                  <div className="empty-state-modal">
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                      <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                      <line x1="16" y1="2" x2="16" y2="6"></line>
                      <line x1="8" y1="2" x2="8" y2="6"></line>
                      <line x1="3" y1="10" x2="21" y2="10"></line>
                    </svg>
                    <p>No leave days added yet</p>
                  </div>
                ) : (
                  <div className="leave-days-grid">
                    {getMemberLeaveDays(showLeaveModal).map((leaveDay) => (
                      <div key={leaveDay.id} className="leave-day-card">
                        <div className="leave-day-date">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                            <line x1="16" y1="2" x2="16" y2="6"></line>
                            <line x1="8" y1="2" x2="8" y2="6"></line>
                            <line x1="3" y1="10" x2="21" y2="10"></line>
                          </svg>
                          <span>{formatDate(new Date(leaveDay.leave_date))}</span>
                        </div>
                        <button
                          className="delete-leave-btn"
                          onClick={() => handleDeleteLeaveDayWithRefresh(leaveDay.id, showLeaveModal)}
                          disabled={loading}
                          title="Delete leave day"
                        >
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <polyline points="3 6 5 6 21 6"></polyline>
                            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                          </svg>
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}

