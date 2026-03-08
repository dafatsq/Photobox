import React, { useState } from 'react';
import { useAdminStore, FrameOption } from '../store/adminStore';
import './AdminPanel.css';

const AdminPanel: React.FC = () => {
    const {
        bypassPayment,
        setBypassPayment,
        frames,
        addFrame,
        removeFrame,
        closeAdmin,
    } = useAdminStore();

    const [newId, setNewId] = useState('');
    const [newLabel, setNewLabel] = useState('');
    const [newColor, setNewColor] = useState('#ffffff');

    const handleAddFrame = (e: React.FormEvent) => {
        e.preventDefault();
        if (!newId.trim() || !newLabel.trim()) return;
        // Prevent duplicate IDs
        if (frames.some((f) => f.id === newId.trim())) {
            alert('Frame ID already exists!');
            return;
        }
        addFrame({ id: newId.trim(), label: newLabel.trim(), color: newColor });
        setNewId('');
        setNewLabel('');
        setNewColor('#ffffff');
    };

    return (
        <div className="admin-overlay" onClick={closeAdmin}>
            <div className="admin-panel" onClick={(e) => e.stopPropagation()}>
                <div className="admin-header">
                    <h2>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="20" height="20" style={{ display: 'inline', verticalAlign: 'middle', marginRight: '0.5rem' }}>
                            <circle cx="12" cy="12" r="3" />
                            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
                        </svg>
                        Admin Panel
                    </h2>
                    <button className="admin-close" onClick={closeAdmin}>✕</button>
                </div>

                {/* Section 1: Settings */}
                <section className="admin-section">
                    <h3>Settings</h3>
                    <label className="admin-toggle-row">
                        <span>Bypass Payment</span>
                        <input
                            type="checkbox"
                            checked={bypassPayment}
                            onChange={(e) => setBypassPayment(e.target.checked)}
                        />
                        <span className={`toggle-switch ${bypassPayment ? 'on' : ''}`} />
                    </label>
                </section>

                {/* Section 2: Frame Manager */}
                <section className="admin-section">
                    <h3>Frame Manager</h3>
                    <table className="admin-table">
                        <thead>
                            <tr>
                                <th>ID</th>
                                <th>Label</th>
                                <th>Color</th>
                                <th>Preview</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            {frames.map((f) => (
                                <tr key={f.id}>
                                    <td><code>{f.id}</code></td>
                                    <td>{f.label}</td>
                                    <td><code>{f.color}</code></td>
                                    <td>
                                        <div
                                            className="color-swatch"
                                            style={{ backgroundColor: f.color }}
                                        />
                                    </td>
                                    <td>
                                        <button
                                            className="admin-delete-btn"
                                            onClick={() => removeFrame(f.id)}
                                        >
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="16" height="16">
                                                <polyline points="3 6 5 6 21 6" />
                                                <path d="M19 6l-1 14H6L5 6" />
                                                <path d="M10 11v6M14 11v6" />
                                                <path d="M9 6V4h6v2" />
                                            </svg>
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>

                    <form className="admin-add-form" onSubmit={handleAddFrame}>
                        <input
                            type="text"
                            placeholder="ID (e.g. pastel_pink)"
                            value={newId}
                            onChange={(e) => setNewId(e.target.value)}
                            required
                        />
                        <input
                            type="text"
                            placeholder="Label (e.g. Pastel Pink)"
                            value={newLabel}
                            onChange={(e) => setNewLabel(e.target.value)}
                            required
                        />
                        <input
                            type="color"
                            value={newColor}
                            onChange={(e) => setNewColor(e.target.value)}
                            title="Frame color"
                        />
                        <button type="submit" className="admin-add-btn">+ Add</button>
                    </form>
                </section>
            </div>
        </div>
    );
};

export default AdminPanel;
