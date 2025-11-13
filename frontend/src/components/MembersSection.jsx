// Members Section Component
import { useState } from 'react';
import { useApp } from '../context/AppContext';
import './MembersSection.css';

export function MembersSection() {
  const { members, createMember, deleteMember, loading } = useApp();
  const [name, setName] = useState('');
  const [error, setError] = useState('');

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
          members.map((member) => (
            <div key={member.id} className="member-item">
              <span>{member.name}</span>
              <button
                className="delete-btn"
                onClick={() => handleDelete(member.id)}
                disabled={loading}
              >
                Delete
              </button>
            </div>
          ))
        )}
      </div>
    </section>
  );
}

