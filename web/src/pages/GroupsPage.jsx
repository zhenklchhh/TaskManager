import { useState, useEffect } from 'react';
import { useApi } from '../context/ApiContext';
import { FolderKanban, Plus, Pencil, Trash2, X } from 'lucide-react';

export default function GroupsPage() {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [editId, setEditId] = useState(null);
  const [form, setForm] = useState({ name: '', description: '', color: '#6366f1' });
  const { get, post, put, del } = useApi();

  useEffect(() => {
    loadGroups();
  }, []);

  const loadGroups = async () => {
    try {
      const res = await get('/groups');
      if (res.ok) setGroups(await res.json());
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (e) => {
    e.preventDefault();
    const res = await post('/groups', form);
    if (res.ok) {
      const group = await res.json();
      setGroups((prev) => [...prev, group]);
      setForm({ name: '', description: '', color: '#6366f1' });
      setShowCreate(false);
    }
  };

  const handleUpdate = async (e) => {
    e.preventDefault();
    const res = await put(`/groups/${editId}`, form);
    if (res.ok) {
      const updated = await res.json();
      setGroups((prev) => prev.map((g) => (g.id === editId ? updated : g)));
      setEditId(null);
      setForm({ name: '', description: '', color: '#6366f1' });
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('Delete this group?')) return;
    const res = await del(`/groups/${id}`);
    if (res.ok) {
      setGroups((prev) => prev.filter((g) => g.id !== id));
    }
  };

  const startEdit = (group) => {
    setEditId(group.id);
    setForm({ name: group.name, description: group.description, color: group.color });
    setShowCreate(false);
  };

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center h-full">
        <div className="text-slate-400">Loading...</div>
      </div>
    );
  }

  return (
    <div className="p-8 max-w-6xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold gradient-text flex items-center gap-3">
            <FolderKanban className="w-8 h-8 text-purple-400" />
            Project Groups
          </h1>
          <p className="text-slate-400 mt-1">Organize your tasks into colorful project groups · {groups.length} {groups.length === 1 ? 'group' : 'groups'}</p>
        </div>
        <button
          onClick={() => { setShowCreate(true); setEditId(null); setForm({ name: '', description: '', color: '#6366f1' }); }}
          className="flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white text-sm font-semibold rounded-lg shadow-lg shadow-indigo-500/30"
        >
          <Plus className="w-4 h-4" />
          New Group
        </button>
      </div>

      {/* Create/Edit form */}
      {(showCreate || editId) && (
        <div className="glass-card rounded-2xl p-6 mb-6 fade-in">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-white">
              {editId ? 'Edit Group' : 'Create Group'}
            </h3>
            <button
              onClick={() => { setShowCreate(false); setEditId(null); }}
              className="text-slate-400 hover:text-white"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
          <form onSubmit={editId ? handleUpdate : handleCreate} className="space-y-3">
            <div className="flex gap-3">
              <input
                type="text"
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="flex-1 px-3 py-2 bg-slate-700/50 border border-slate-600 rounded-lg text-white text-sm placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="Group name"
                required
              />
              <input
                type="color"
                value={form.color}
                onChange={(e) => setForm({ ...form, color: e.target.value })}
                className="w-10 h-10 rounded-lg border border-slate-600 cursor-pointer"
              />
            </div>
            <input
              type="text"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600 rounded-lg text-white text-sm placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              placeholder="Description (optional)"
            />
            <button
              type="submit"
              className="px-5 py-2.5 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white text-sm font-semibold rounded-lg shadow-md shadow-indigo-500/30"
            >
              {editId ? 'Update Group' : 'Create Group'}
            </button>
          </form>
        </div>
      )}

      {/* Groups grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
        {groups.map((group) => (
          <div
            key={group.id}
            className="glass-card glass-card-hover rounded-2xl p-6 relative overflow-hidden group"
          >
            <div className="absolute top-0 left-0 w-full h-1" style={{ background: `linear-gradient(90deg, ${group.color}, ${group.color}80)` }}></div>
            <div className="flex items-start justify-between mb-3">
              <div className="flex items-center gap-3">
                <div
                  className="w-10 h-10 rounded-xl flex items-center justify-center shadow-lg"
                  style={{ backgroundColor: group.color + '30', boxShadow: `0 0 20px ${group.color}30` }}
                >
                  <FolderKanban className="w-5 h-5" style={{ color: group.color }} />
                </div>
                <h3 className="text-base font-semibold text-white">{group.name}</h3>
              </div>
              <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                  onClick={() => startEdit(group)}
                  className="p-1.5 text-slate-400 hover:text-indigo-400 hover:bg-indigo-500/10 rounded-lg"
                >
                  <Pencil className="w-3.5 h-3.5" />
                </button>
                <button
                  onClick={() => handleDelete(group.id)}
                  className="p-1.5 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
            {group.description && (
              <p className="text-sm text-slate-400 mb-4">{group.description}</p>
            )}
            <div className="flex items-center justify-between pt-3 border-t border-slate-700/30">
              <span className="text-xs text-slate-500 uppercase tracking-wide">Tasks</span>
              <span className="text-lg font-bold text-white">{group.task_count}</span>
            </div>
          </div>
        ))}
      </div>

      {groups.length === 0 && !showCreate && (
        <div className="glass-card rounded-2xl py-16 text-center fade-in">
          <FolderKanban className="w-16 h-16 mx-auto mb-4 text-slate-600" />
          <p className="text-lg text-slate-300 mb-2">No groups yet</p>
          <p className="text-sm text-slate-500 mb-6">Create your first group to organize your tasks</p>
          <button
            onClick={() => setShowCreate(true)}
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white text-sm font-semibold rounded-lg shadow-lg shadow-indigo-500/30"
          >
            <Plus className="w-4 h-4" />
            Create First Group
          </button>
        </div>
      )}
    </div>
  );
}
