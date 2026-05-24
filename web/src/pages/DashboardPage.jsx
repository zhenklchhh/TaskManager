import { useState, useEffect, useCallback } from 'react';
import { useApi } from '../context/ApiContext';
import { useAuth } from '../context/AuthContext';
import {
  PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Legend
} from 'recharts';
import {
  Activity, CheckCircle2, Clock, AlertTriangle, XCircle, ListTodo,
  Timer, Zap, BarChart3, RefreshCw, Plus, Trash2, ChevronDown, X,
} from 'lucide-react';
import TaskDetailModal from '../components/TaskDetailModal';

const STATUS_COLORS = {
  pending: '#f59e0b',
  running: '#3b82f6',
  completed: '#10b981',
  failed: '#ef4444',
  cancelled: '#6b7280',
  scheduled: '#8b5cf6',
};

const STATUS_ICONS = {
  pending: Clock,
  running: Zap,
  completed: CheckCircle2,
  failed: XCircle,
  cancelled: AlertTriangle,
  scheduled: Timer,
};

const PAYLOAD_TEMPLATES = {
  send_email:    JSON.stringify({ to: 'user@example.com', from: 'system@taskmanager.local', body: 'Hello!' }, null, 2),
  send_webhook:  JSON.stringify({ url: 'https://httpbin.org/post', method: 'POST', body: { msg: 'Hello!' } }, null, 2),
};

function CreateTaskModal({ onClose, onCreated, groups, members }) {
  const { post } = useApi();
  const [form, setForm] = useState({ title: '', type: 'send_email', payload: PAYLOAD_TEMPLATES.send_email, cron_expr: '*/5 * * * *', priority: 5, group_id: '', assigned_to: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const setField = (k, v) => setForm((f) => ({ ...f, [k]: v }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const res = await post('/tasks', {
        ...form,
        priority: Number(form.priority),
        group_id: form.group_id || undefined,
        assigned_to: form.assigned_to || undefined,
      });
      if (!res.ok) throw new Error(await res.text());
      onCreated();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="w-full max-w-lg glass-card rounded-2xl p-6 shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-bold text-white flex items-center gap-2"><Plus className="w-5 h-5 text-indigo-400" /> New Task</h2>
          <button onClick={onClose} className="p-1.5 text-slate-400 hover:text-white rounded-lg transition-colors"><X className="w-5 h-5" /></button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs text-slate-400 mb-1">Title</label>
            <input required value={form.title} onChange={(e) => setField('title', e.target.value)} placeholder="My Task"
              className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs text-slate-400 mb-1">Type</label>
              <select value={form.type} onChange={(e) => { setField('type', e.target.value); setField('payload', PAYLOAD_TEMPLATES[e.target.value] || '{}'); }}
                className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500">
                <option value="send_email">Email</option>
                <option value="send_webhook">Webhook</option>
              </select>
            </div>
            <div>
              <label className="block text-xs text-slate-400 mb-1">Priority (1–10)</label>
              <input type="number" min={1} max={10} value={form.priority} onChange={(e) => setField('priority', e.target.value)}
                className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500" />
            </div>
          </div>
          {(groups.length > 0 || members.length > 0) && (
            <div className="grid grid-cols-2 gap-4">
              {groups.length > 0 && (
                <div>
                  <label className="block text-xs text-slate-400 mb-1">Group</label>
                  <select value={form.group_id} onChange={(e) => setField('group_id', e.target.value)}
                    className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500">
                    <option value="">No Group</option>
                    {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
                  </select>
                </div>
              )}
              {members.length > 0 && (
                <div>
                  <label className="block text-xs text-slate-400 mb-1">Assign To</label>
                  <select value={form.assigned_to} onChange={(e) => setField('assigned_to', e.target.value)}
                    className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500">
                    <option value="">Unassigned</option>
                    {members.map((m) => <option key={m.user_id} value={m.user_id}>{m.name}</option>)}
                  </select>
                </div>
              )}
            </div>
          )}
          <div>
            <label className="block text-xs text-slate-400 mb-1">Cron Expression</label>
            <input required value={form.cron_expr} onChange={(e) => setField('cron_expr', e.target.value)}
              className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white font-mono text-sm focus:outline-none focus:border-indigo-500" />
          </div>
          <div>
            <label className="block text-xs text-slate-400 mb-1">Payload (JSON)</label>
            <textarea rows={4} required value={form.payload} onChange={(e) => setField('payload', e.target.value)}
              className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white font-mono text-sm focus:outline-none focus:border-indigo-500" />
          </div>
          {error && <p className="text-red-400 text-xs">{error}</p>}
          <div className="flex gap-3 justify-end pt-1">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-slate-400 hover:text-white transition-colors">Cancel</button>
            <button type="submit" disabled={loading}
              className="px-5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium rounded-lg disabled:opacity-50 transition-colors">
              {loading ? 'Creating…' : 'Create Task'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const [stats, setStats] = useState(null);
  const [tasks, setTasks] = useState([]);
  const [groups, setGroups] = useState([]);
  const [selectedGroup, setSelectedGroup] = useState('all');
  const [viewMode, setViewMode] = useState('company'); // 'company' | 'personal'
  const [selectedTaskId, setSelectedTaskId] = useState(null);
  const [initialLoading, setInitialLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [selectedIds, setSelectedIds] = useState(new Set());
  const [createOpen, setCreateOpen] = useState(false);
  const [batchPriorityOpen, setBatchPriorityOpen] = useState(false);
  const [batchPriority, setBatchPriority] = useState(5);
  const [batchLoading, setBatchLoading] = useState('');
  const [members, setMembers] = useState([]);
  const { get, post, patch, put } = useApi();
  const { user } = useAuth();

  const handleBatchCancel = useCallback(async () => {
    if (!selectedIds.size) return;
    setBatchLoading('cancel');
    try {
      const res = await post('/tasks/batch/cancel', { ids: [...selectedIds] });
      if (res.ok) { setSelectedIds(new Set()); loadData(false); }
    } finally { setBatchLoading(''); }
  }, [selectedIds]);

  const handleBatchPriority = useCallback(async () => {
    if (!selectedIds.size) return;
    setBatchLoading('priority');
    try {
      const res = await put('/tasks/batch/priority', { ids: [...selectedIds], priority: Number(batchPriority) });
      if (res.ok) { setSelectedIds(new Set()); setBatchPriorityOpen(false); loadData(false); }
    } finally { setBatchLoading(''); }
  }, [selectedIds, batchPriority]);

  const toggleSelect = (id) => setSelectedIds((prev) => {
    const next = new Set(prev);
    next.has(id) ? next.delete(id) : next.add(id);
    return next;
  });

  const toggleSelectAll = () => {
    if (selectedIds.size === filteredTasks.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(filteredTasks.map((t) => t.id)));
    }
  };

  const loadData = async (isInitial = false) => {
    if (isInitial) setInitialLoading(true);
    else setRefreshing(true);
    try {
      const [statsRes, tasksRes, groupsRes] = await Promise.all([
        get('/dashboard/stats'),
        get('/dashboard/tasks'),
        get('/groups'),
      ]);
      if (statsRes.ok) setStats(await statsRes.json());
      if (tasksRes.ok) {
        const data = await tasksRes.json();
        setTasks(data.tasks || []);
      }
      if (groupsRes.ok) {
        const data = await groupsRes.json();
        setGroups(Array.isArray(data) ? data : []);
      }
    } finally {
      setInitialLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    get('/companies/members').then((res) => {
      if (res.ok) res.json().then((data) => setMembers(Array.isArray(data) ? data : []));
    });
  }, []);

  useEffect(() => {
    loadData(true);
  }, [selectedGroup, viewMode]);

  useEffect(() => {
    const interval = setInterval(() => loadData(false), 5000);
    return () => clearInterval(interval);
  }, [selectedGroup, viewMode]);

  const filteredTasks = tasks.filter((task) => {
    if (selectedGroup !== 'all' && task.group_id !== selectedGroup) return false;
    if (viewMode === 'personal' && task.assigned_to !== user?.id) return false;
    return true;
  });

  const statusData = stats
    ? Object.entries(STATUS_COLORS).map(([status, color]) => ({
        name: status,
        value: stats[`${status}_tasks`] || 0,
        color,
      })).filter(d => d.value > 0)
    : [];

  const statCards = stats ? [
    { label: 'Total Tasks', value: stats.total_tasks || 0, icon: ListTodo, gradient: 'from-indigo-500 to-purple-600', iconBg: 'bg-indigo-500/20', iconColor: 'text-indigo-300' },
    { label: 'Pending', value: stats.pending_tasks || 0, icon: Clock, gradient: 'from-amber-500 to-orange-600', iconBg: 'bg-amber-500/20', iconColor: 'text-amber-300' },
    { label: 'Running', value: stats.running_tasks || 0, icon: Zap, gradient: 'from-blue-500 to-cyan-600', iconBg: 'bg-blue-500/20', iconColor: 'text-blue-300' },
    { label: 'Completed', value: stats.completed_tasks || 0, icon: CheckCircle2, gradient: 'from-emerald-500 to-teal-600', iconBg: 'bg-emerald-500/20', iconColor: 'text-emerald-300' },
    { label: 'Failed', value: stats.failed_tasks || 0, icon: XCircle, gradient: 'from-red-500 to-pink-600', iconBg: 'bg-red-500/20', iconColor: 'text-red-300' },
  ] : [];

  const selectedTask = selectedTaskId ? tasks.find((t) => t.id === selectedTaskId) ?? null : null;

  if (initialLoading && !stats) {
    return (
      <div className="p-8 flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-indigo-400 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold gradient-text flex items-center gap-3">
            <BarChart3 className="w-8 h-8 text-indigo-400" />
            Dashboard
          </h1>
          <p className="text-slate-400 mt-1">Welcome back, {user?.name?.split(' ')[0] || 'there'}! Here's your task overview.</p>
        </div>
        <div className="flex items-center gap-3">
          {/* View mode toggle */}
          <div className="flex glass-card rounded-lg p-1">
            <button
              onClick={() => setViewMode('company')}
              className={`px-4 py-1.5 text-xs font-medium rounded-md transition-all ${
                viewMode === 'company'
                  ? 'bg-gradient-to-r from-indigo-500 to-purple-600 text-white shadow-md'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              Company
            </button>
            <button
              onClick={() => setViewMode('personal')}
              className={`px-4 py-1.5 text-xs font-medium rounded-md transition-all ${
                viewMode === 'personal'
                  ? 'bg-gradient-to-r from-indigo-500 to-purple-600 text-white shadow-md'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              Personal
            </button>
          </div>

          {/* Group filter */}
          <select
            value={selectedGroup}
            onChange={(e) => setSelectedGroup(e.target.value)}
            className="px-3 py-2 glass-card rounded-lg text-sm text-white"
          >
            <option value="all">All Groups</option>
            {groups.map((g) => (
              <option key={g.id} value={g.id}>{g.name}</option>
            ))}
          </select>

          <button
            onClick={() => setCreateOpen(true)}
            className="flex items-center gap-1.5 px-3 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium rounded-lg transition-colors"
          >
            <Plus className="w-4 h-4" /> New Task
          </button>
          <button
            onClick={() => loadData(false)}
            className={`p-2 glass-card rounded-lg transition-colors ${
              refreshing ? 'text-indigo-400' : 'text-slate-400 hover:text-white'
            }`}
            title="Refresh"
          >
            <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
        {statCards.map(({ label, value, icon: Icon, gradient, iconBg, iconColor }) => (
          <div
            key={label}
            className="glass-card glass-card-hover rounded-2xl p-5 relative overflow-hidden group"
          >
            <div className={`absolute inset-0 bg-gradient-to-br ${gradient} opacity-0 group-hover:opacity-10 transition-opacity`}></div>
            <div className="relative">
              <div className={`inline-flex items-center justify-center w-10 h-10 rounded-xl ${iconBg} mb-3`}>
                <Icon className={`w-5 h-5 ${iconColor}`} />
              </div>
              <div className="text-xs text-slate-400 font-medium uppercase tracking-wide mb-1">{label}</div>
              <div className="text-3xl font-bold text-white">{value}</div>
            </div>
          </div>
        ))}
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {/* Status pie chart */}
        <div className="glass-card rounded-2xl p-6">
          <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2"><Activity className="w-4 h-4 text-indigo-400" /> Task Status Distribution</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={statusData}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={90}
                paddingAngle={2}
                dataKey="value"
              >
                {statusData.map((entry, i) => (
                  <Cell key={i} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px' }}
                labelStyle={{ color: '#fff' }}
              />
              <Legend
                formatter={(value) => <span className="text-slate-300 text-xs capitalize">{value}</span>}
              />
            </PieChart>
          </ResponsiveContainer>
        </div>

        {/* Groups bar chart */}
        <div className="glass-card rounded-2xl p-6">
          <h3 className="text-base font-semibold text-white mb-4 flex items-center gap-2"><BarChart3 className="w-4 h-4 text-purple-400" /> Tasks by Group</h3>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={groups.map(g => ({ name: g.name, tasks: g.task_count, fill: g.color }))}>
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 12 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 12 }} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px' }}
                labelStyle={{ color: '#fff' }}
              />
              <Bar dataKey="tasks" radius={[4, 4, 0, 0]}>
                {groups.map((g, i) => (
                  <Cell key={i} fill={g.color} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Batch action bar */}
      {selectedIds.size > 0 && (
        <div className="glass-card rounded-xl px-5 py-3 flex items-center gap-3">
          <span className="text-sm text-slate-300 font-medium">{selectedIds.size} selected</span>
          <div className="flex gap-2 ml-2">
            <button
              onClick={handleBatchCancel}
              disabled={!!batchLoading}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-amber-400 bg-amber-500/10 border border-amber-500/30 rounded-lg hover:bg-amber-500/20 disabled:opacity-50 transition-colors"
            >
              <Trash2 className="w-3.5 h-3.5" />
              {batchLoading === 'cancel' ? 'Cancelling…' : 'Cancel Selected'}
            </button>
            <div className="relative">
              <button
                onClick={() => setBatchPriorityOpen((v) => !v)}
                className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-indigo-400 bg-indigo-500/10 border border-indigo-500/30 rounded-lg hover:bg-indigo-500/20 transition-colors"
              >
                Set Priority <ChevronDown className="w-3 h-3" />
              </button>
              {batchPriorityOpen && (
                <div className="absolute top-full mt-1 left-0 z-20 bg-slate-800 border border-slate-700 rounded-xl p-3 shadow-xl min-w-40">
                  <label className="block text-xs text-slate-400 mb-1">Priority (1–10)</label>
                  <input
                    type="number" min={1} max={10}
                    value={batchPriority}
                    onChange={(e) => setBatchPriority(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-1.5 text-white text-sm mb-2 focus:outline-none focus:border-indigo-500"
                  />
                  <button
                    onClick={handleBatchPriority}
                    disabled={!!batchLoading}
                    className="w-full py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium rounded-lg disabled:opacity-50 transition-colors"
                  >
                    {batchLoading === 'priority' ? 'Applying…' : 'Apply'}
                  </button>
                </div>
              )}
            </div>
          </div>
          <button onClick={() => setSelectedIds(new Set())} className="ml-auto text-slate-400 hover:text-white">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* Task list */}
      <div className="glass-card rounded-2xl overflow-hidden">
        <div className="px-6 py-4 border-b border-slate-700/50 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              checked={filteredTasks.length > 0 && selectedIds.size === filteredTasks.length}
              onChange={toggleSelectAll}
              className="w-4 h-4 rounded accent-indigo-500"
            />
            <h3 className="text-base font-semibold text-white flex items-center gap-2">
              <Activity className="w-4 h-4 text-indigo-400" />
              Recent Tasks
            </h3>
          </div>
          <span className="text-xs text-slate-400 px-2 py-1 bg-slate-700/30 rounded-full">{filteredTasks.length} tasks</span>
        </div>
        <div className="divide-y divide-slate-700/30">
          {filteredTasks.slice(0, 20).map((task) => {
            const Icon = STATUS_ICONS[task.status] || Clock;
            return (
              <div
                key={task.id}
                className={`px-6 py-4 flex items-center justify-between hover:bg-slate-700/20 transition-colors cursor-pointer ${
                  selectedIds.has(task.id) ? 'bg-indigo-500/5' : ''
                }`}
                onClick={() => setSelectedTaskId(task.id)}
              >
                <div className="flex items-center gap-3">
                  <input
                    type="checkbox"
                    checked={selectedIds.has(task.id)}
                    onChange={(e) => { e.stopPropagation(); toggleSelect(task.id); }}
                    onClick={(e) => e.stopPropagation()}
                    className="w-4 h-4 rounded accent-indigo-500"
                  />
                  <div className="w-9 h-9 rounded-lg flex items-center justify-center" style={{ backgroundColor: (STATUS_COLORS[task.status] || '#6b7280') + '20' }}>
                    <Icon className="w-4 h-4" style={{ color: STATUS_COLORS[task.status] || '#6b7280' }} />
                  </div>
                  <div>
                    <div className="text-sm text-white font-medium">{task.title}</div>
                    <div className="text-xs text-slate-400">{task.type}</div>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <span
                    className="text-xs px-3 py-1 rounded-full capitalize font-medium"
                    style={{
                      backgroundColor: (STATUS_COLORS[task.status] || '#6b7280') + '20',
                      color: STATUS_COLORS[task.status] || '#6b7280',
                    }}
                  >
                    {task.status}
                  </span>
                  {task.priority !== undefined && (
                    <span className="text-xs text-slate-500 px-2 py-1 bg-slate-700/30 rounded-full">P{task.priority}</span>
                  )}
                </div>
              </div>
            );
          })}
          {filteredTasks.length === 0 && (
            <div className="px-6 py-16 text-center">
              <ListTodo className="w-12 h-12 text-slate-600 mx-auto mb-3" />
              <div className="text-slate-400">No tasks found</div>
              <div className="text-xs text-slate-500 mt-1">Try changing filters or creating new tasks</div>
            </div>
          )}
        </div>
      </div>

      {selectedTask && (
        <TaskDetailModal
          task={selectedTask}
          onClose={() => setSelectedTaskId(null)}
          onRefresh={() => loadData(false)}
          groups={groups}
          members={members}
        />
      )}

      {createOpen && (
        <CreateTaskModal
          onClose={() => setCreateOpen(false)}
          onCreated={() => { setCreateOpen(false); loadData(false); }}
          groups={groups}
          members={members}
        />
      )}
    </div>
  );
}
