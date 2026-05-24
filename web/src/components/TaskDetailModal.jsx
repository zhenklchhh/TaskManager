import { useState, useEffect } from 'react';
import {
  X, Clock, Zap, CheckCircle2, XCircle, AlertTriangle, Timer,
  Calendar, Layers, RefreshCw, Ban, Pencil, Trash2,
  RotateCcw, Save, ChevronRight, Activity, FolderKanban, User,
} from 'lucide-react';
import { useApi } from '../context/ApiContext';

const STATUS_CONFIG = {
  pending:   { icon: Clock,         color: 'text-amber-400',   bg: 'bg-amber-500/10',   border: 'border-amber-500/30',   label: 'Pending' },
  running:   { icon: Zap,           color: 'text-blue-400',    bg: 'bg-blue-500/10',    border: 'border-blue-500/30',    label: 'Running' },
  scheduled: { icon: Timer,         color: 'text-purple-400',  bg: 'bg-purple-500/10',  border: 'border-purple-500/30',  label: 'Scheduled' },
  completed: { icon: CheckCircle2,  color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/30', label: 'Completed' },
  failed:    { icon: XCircle,       color: 'text-red-400',     bg: 'bg-red-500/10',     border: 'border-red-500/30',     label: 'Failed' },
  cancelled: { icon: AlertTriangle, color: 'text-slate-400',   bg: 'bg-slate-500/10',   border: 'border-slate-500/30',   label: 'Cancelled' },
};

const CANCELLABLE = new Set(['pending', 'running', 'scheduled']);
const RETRYABLE   = new Set(['failed', 'cancelled']);
const EDITABLE    = new Set(['pending', 'scheduled']);

function fmt(d) {
  if (!d) return 'N/A';
  return new Date(d).toLocaleString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function fmtDuration(ms) {
  if (!ms) return '—';
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`;
}

export default function TaskDetailModal({ task, onClose, onRefresh, groups = [], members = [] }) {
  const { get, post, patch, del } = useApi();
  const [activeTab, setActiveTab]   = useState('details');
  const [executions, setExecutions] = useState([]);
  const [logsLoading, setLogsLoading] = useState(false);
  const [editMode, setEditMode]     = useState(false);
  const [editForm, setEditForm]     = useState({
    title:       task.title,
    priority:    task.priority    ?? 5,
    cron_expr:   task.cron_expr   ?? '',
    max_retries: task.max_retries ?? 3,
    group_id:    task.group_id    ?? '',
  });
  const [actionLoading, setActionLoading] = useState('');
  const [actionError, setActionError]     = useState('');

  useEffect(() => {
    if (activeTab === 'logs') fetchLogs();
  }, [activeTab, task.id]);

  const fetchLogs = async () => {
    setLogsLoading(true);
    try {
      const res = await get(`/tasks/${task.id}/logs`);
      if (res.ok) {
        const data = await res.json();
        setExecutions(data.executions || []);
      }
    } finally {
      setLogsLoading(false);
    }
  };

  const runAction = async (key, fn) => {
    setActionLoading(key);
    setActionError('');
    try {
      await fn();
      onRefresh?.();
      if (key !== 'edit') onClose();
    } catch (err) {
      setActionError(err.message || 'Action failed');
    } finally {
      setActionLoading('');
    }
  };

  const handleCancel = () => runAction('cancel', async () => {
    const res = await post(`/tasks/${task.id}/cancel`, {});
    if (!res.ok) throw new Error(await res.text());
  });

  const handleRetry = () => runAction('retry', async () => {
    const res = await post(`/tasks/${task.id}/retry`, {});
    if (!res.ok) throw new Error(await res.text());
  });

  const handleDelete = () => {
    if (!window.confirm(`Delete task "${task.title}"?`)) return;
    runAction('delete', async () => {
      const res = await del(`/tasks/${task.id}`);
      if (!res.ok) throw new Error(await res.text());
    });
  };

  const numVal = (v) => { const n = Number(v); return v !== '' && !isNaN(n) ? n : undefined; };

  const handleEdit = (e) => {
    e.preventDefault();
    runAction('edit', async () => {
      const body = {
        title:       editForm.title     || undefined,
        priority:    numVal(editForm.priority),
        cron_expr:   editForm.cron_expr || undefined,
        max_retries: numVal(editForm.max_retries),
        group_id:    editForm.group_id !== undefined ? (editForm.group_id || '') : undefined,
      };
      const res = await patch(`/tasks/${task.id}`, body);
      if (!res.ok) throw new Error(await res.text());
      setEditMode(false);
    });
  };

  if (!task) return null;

  const status     = STATUS_CONFIG[task.status] || STATUS_CONFIG.pending;
  const StatusIcon = status.icon;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="w-full max-w-2xl glass-card rounded-2xl overflow-hidden shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-700/50 flex items-center justify-between bg-gradient-to-r from-slate-800/50 to-slate-900/50">
          <div className="flex items-center gap-3">
            <div className={`w-10 h-10 rounded-xl ${status.bg} ${status.border} border flex items-center justify-center`}>
              <StatusIcon className={`w-5 h-5 ${status.color}`} />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">{task.title}</h2>
              <div className="flex items-center gap-2 mt-0.5">
                <span className={`text-xs px-2 py-0.5 rounded-full ${status.bg} ${status.border} border ${status.color} font-medium`}>
                  {status.label}
                </span>
                <span className="text-xs text-slate-400">P{task.priority} · {task.type}</span>
              </div>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-white hover:bg-slate-700/50 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-slate-700/50 bg-slate-800/30">
          {['details', 'logs'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-5 py-2.5 text-sm font-medium transition-colors ${
                activeTab === tab
                  ? 'text-indigo-400 border-b-2 border-indigo-400 bg-indigo-500/5'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              {tab === 'details' ? 'Details' : 'Execution Logs'}
            </button>
          ))}
        </div>

        {/* Body */}
        <div className="max-h-[55vh] overflow-y-auto">

          {/* ── Details tab ── */}
          {activeTab === 'details' && (
            <div className="p-6 space-y-4">
              {editMode ? (
                <form onSubmit={handleEdit} className="space-y-4">
                  <div>
                    <label className="block text-xs text-slate-400 mb-1">Title</label>
                    <input
                      className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500"
                      value={editForm.title}
                      onChange={(e) => setEditForm((f) => ({ ...f, title: e.target.value }))}
                      required
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-xs text-slate-400 mb-1">Priority (1–10)</label>
                      <input
                        type="number" min={1} max={10}
                        className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500"
                        value={editForm.priority}
                        onChange={(e) => setEditForm((f) => ({ ...f, priority: e.target.value }))}
                      />
                    </div>
                    <div>
                      <label className="block text-xs text-slate-400 mb-1">Max Retries</label>
                      <input
                        type="number" min={0} max={10}
                        className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500"
                        value={editForm.max_retries}
                        onChange={(e) => setEditForm((f) => ({ ...f, max_retries: e.target.value }))}
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block text-xs text-slate-400 mb-1">Cron Expression</label>
                    <input
                      className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white font-mono text-sm focus:outline-none focus:border-indigo-500"
                      value={editForm.cron_expr}
                      onChange={(e) => setEditForm((f) => ({ ...f, cron_expr: e.target.value }))}
                    />
                  </div>
                  {groups.length > 0 && (
                    <div>
                      <label className="block text-xs text-slate-400 mb-1">Group</label>
                      <select
                        className="w-full bg-slate-900/50 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-indigo-500"
                        value={editForm.group_id}
                        onChange={(e) => setEditForm((f) => ({ ...f, group_id: e.target.value }))}
                      >
                        <option value="">No Group</option>
                        {groups.map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}
                      </select>
                    </div>
                  )}
                  {actionError && <p className="text-red-400 text-xs">{actionError}</p>}
                  <div className="flex gap-2 justify-end">
                    <button
                      type="button"
                      onClick={() => { setEditMode(false); setActionError(''); }}
                      className="px-4 py-2 text-sm text-slate-400 hover:text-white transition-colors"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={actionLoading === 'edit'}
                      className="flex items-center gap-1.5 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm rounded-lg disabled:opacity-50 transition-colors"
                    >
                      <Save className="w-3.5 h-3.5" />
                      {actionLoading === 'edit' ? 'Saving…' : 'Save'}
                    </button>
                  </div>
                </form>
              ) : (
                <>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="p-4 rounded-xl bg-slate-800/30 border border-slate-700/50">
                      <div className="flex items-center gap-2 mb-1">
                        <Calendar className="w-4 h-4 text-purple-400" />
                        <span className="text-xs text-slate-400 uppercase tracking-wide">Next Run</span>
                      </div>
                      <div className="text-sm text-white font-medium">{fmt(task.next_run_at)}</div>
                    </div>
                    <div className="p-4 rounded-xl bg-slate-800/30 border border-slate-700/50">
                      <div className="flex items-center gap-2 mb-1">
                        <Clock className="w-4 h-4 text-cyan-400" />
                        <span className="text-xs text-slate-400 uppercase tracking-wide">Cron</span>
                      </div>
                      <div className="text-sm text-white font-mono">{task.cron_expr || '—'}</div>
                    </div>
                  </div>

                  <div className="p-4 rounded-xl bg-slate-800/30 border border-slate-700/50">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <Zap className="w-4 h-4 text-amber-400" />
                        <span className="text-xs text-slate-400 uppercase tracking-wide">Retry Progress</span>
                      </div>
                      <span className="text-sm text-white font-medium">{task.retry_count} / {task.max_retries}</span>
                    </div>
                    <div className="w-full h-2 bg-slate-700 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-indigo-500 to-purple-600 transition-all"
                        style={{ width: `${Math.min(((task.retry_count || 0) / (task.max_retries || 1)) * 100, 100)}%` }}
                      />
                    </div>
                  </div>

                  {task.last_error_msg && (
                    <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/30">
                      <div className="flex items-center gap-2 mb-2">
                        <XCircle className="w-4 h-4 text-red-400" />
                        <span className="text-xs text-red-300 uppercase tracking-wide font-medium">Last Error</span>
                      </div>
                      <div className="text-sm text-red-200 font-mono bg-red-500/5 p-3 rounded-lg">{task.last_error_msg}</div>
                    </div>
                  )}

                  <div className="p-4 rounded-xl bg-slate-800/30 border border-slate-700/50">
                    <div className="flex items-center gap-2 mb-3">
                      <Layers className="w-4 h-4 text-emerald-400" />
                      <span className="text-xs text-slate-400 uppercase tracking-wide">Payload</span>
                    </div>
                    <pre className="text-xs text-slate-300 font-mono bg-slate-900/50 p-4 rounded-lg overflow-x-auto max-h-36">
                      {(() => { try { return JSON.stringify(JSON.parse(task.payload), null, 2); } catch { return task.payload; } })()}
                    </pre>
                  </div>

                  <div className="grid grid-cols-2 gap-4 text-xs">
                    <div><span className="text-slate-400">Created:</span><span className="text-slate-300 ml-2">{fmt(task.created_at)}</span></div>
                    <div><span className="text-slate-400">Updated:</span><span className="text-slate-300 ml-2">{fmt(task.updated_at)}</span></div>
                  </div>

                  {(task.group_id || task.assigned_to) && (
                    <div className="flex flex-wrap gap-4 text-xs">
                      {task.group_id && (
                        <div className="flex items-center gap-1.5">
                          <FolderKanban className="w-3.5 h-3.5 text-purple-400" />
                          <span className="text-slate-400">Group:</span>
                          <span className="text-white">{groups.find((g) => g.id === task.group_id)?.name ?? task.group_id}</span>
                        </div>
                      )}
                      {task.assigned_to && (
                        <div className="flex items-center gap-1.5">
                          <User className="w-3.5 h-3.5 text-cyan-400" />
                          <span className="text-slate-400">Assigned:</span>
                          <span className="text-white">{members.find((m) => m.user_id === task.assigned_to)?.name ?? task.assigned_to}</span>
                        </div>
                      )}
                    </div>
                  )}

                  {actionError && <p className="text-red-400 text-xs mt-2">{actionError}</p>}
                </>
              )}
            </div>
          )}

          {/* ── Logs tab ── */}
          {activeTab === 'logs' && (
            <div className="p-6">
              {logsLoading ? (
                <div className="flex items-center justify-center py-12">
                  <RefreshCw className="w-6 h-6 text-indigo-400 animate-spin" />
                </div>
              ) : executions.length === 0 ? (
                <div className="text-center py-12">
                  <Activity className="w-10 h-10 text-slate-600 mx-auto mb-3" />
                  <p className="text-slate-400 text-sm">No executions recorded yet</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {executions.map((ex) => {
                    const cfg  = STATUS_CONFIG[ex.status] || STATUS_CONFIG.pending;
                    const Icon = cfg.icon;
                    return (
                      <div key={ex.id} className="p-3 rounded-xl bg-slate-800/30 border border-slate-700/50">
                        <div className="flex items-center justify-between mb-1">
                          <div className="flex items-center gap-2">
                            <Icon className={`w-4 h-4 ${cfg.color}`} />
                            <span className={`text-xs font-medium ${cfg.color} capitalize`}>{ex.status}</span>
                            {ex.worker_id && (
                              <span className="text-xs text-slate-500">· worker {ex.worker_id.slice(0, 8)}</span>
                            )}
                          </div>
                          <span className="text-xs text-slate-500">{fmtDuration(ex.duration_ms)}</span>
                        </div>
                        <div className="flex gap-4 text-xs text-slate-500 mt-1">
                          <span>Start: {fmt(ex.started_at)}</span>
                          {ex.finished_at && (
                            <span>
                              <ChevronRight className="w-3 h-3 inline -mt-0.5" />
                              End: {fmt(ex.finished_at)}
                            </span>
                          )}
                        </div>
                        {ex.error && (
                          <p className="text-xs text-red-300 font-mono mt-2 bg-red-500/5 p-2 rounded">{ex.error}</p>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-slate-700/50 bg-slate-800/30 flex items-center justify-between">
          <div className="flex gap-2">
            {CANCELLABLE.has(task.status) && !editMode && (
              <button
                onClick={handleCancel}
                disabled={!!actionLoading}
                className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-amber-400 bg-amber-500/10 border border-amber-500/30 rounded-lg hover:bg-amber-500/20 disabled:opacity-50 transition-colors"
              >
                <Ban className="w-3.5 h-3.5" />
                {actionLoading === 'cancel' ? 'Cancelling…' : 'Cancel Task'}
              </button>
            )}
            {RETRYABLE.has(task.status) && !editMode && (
              <button
                onClick={handleRetry}
                disabled={!!actionLoading}
                className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-emerald-400 bg-emerald-500/10 border border-emerald-500/30 rounded-lg hover:bg-emerald-500/20 disabled:opacity-50 transition-colors"
              >
                <RotateCcw className="w-3.5 h-3.5" />
                {actionLoading === 'retry' ? 'Retrying…' : 'Retry'}
              </button>
            )}
            {EDITABLE.has(task.status) && !editMode && (
              <button
                onClick={() => setEditMode(true)}
                className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-indigo-400 bg-indigo-500/10 border border-indigo-500/30 rounded-lg hover:bg-indigo-500/20 transition-colors"
              >
                <Pencil className="w-3.5 h-3.5" /> Edit
              </button>
            )}
          </div>
          <div className="flex gap-2">
            {!editMode && (
              <button
                onClick={handleDelete}
                disabled={!!actionLoading}
                className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-red-400 bg-red-500/10 border border-red-500/30 rounded-lg hover:bg-red-500/20 disabled:opacity-50 transition-colors"
              >
                <Trash2 className="w-3.5 h-3.5" />
                {actionLoading === 'delete' ? 'Deleting…' : 'Delete'}
              </button>
            )}
            <button
              onClick={onClose}
              className="px-4 py-2 bg-slate-700/50 hover:bg-slate-700 text-white text-sm font-medium rounded-lg transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
