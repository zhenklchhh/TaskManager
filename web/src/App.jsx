import { useState, useEffect, useCallback } from 'react'
import { BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { Activity, CheckCircle, Clock, XCircle, Plus, RefreshCw, Server, Database, Zap, ListChecks, TrendingUp, LayoutDashboard } from 'lucide-react'

const API_BASE = ''

const STATUS_COLORS = {
  pending: '#f59e0b',
  scheduled: '#3b82f6',
  running: '#8b5cf6',
  completed: '#10b981',
  failed: '#ef4444',
}

const STATUS_ICONS = {
  pending: Clock,
  scheduled: ListChecks,
  running: Activity,
  completed: CheckCircle,
  failed: XCircle,
}

const PRIORITY_COLORS = ['#6366f1','#8b5cf6','#a78bfa','#c4b5fd','#818cf8','#6366f1','#4f46e5','#4338ca','#3730a3','#312e81']

function formatDate(d) {
  if (!d) return '\u2014'
  const date = new Date(d)
  if (isNaN(date.getTime())) return '\u2014'
  return date.toLocaleString('ru-RU', { day:'2-digit', month:'2-digit', year:'numeric', hour:'2-digit', minute:'2-digit', second:'2-digit' })
}

function StatCard({ title, value, icon: Icon, color, subtitle }) {
  return (
    <div className="bg-slate-800/50 border border-slate-700/50 rounded-xl p-5 flex items-start gap-4 hover:bg-slate-800/80 transition-all">
      <div className={`p-3 rounded-lg ${color}`}>
        <Icon size={22} className="text-white" />
      </div>
      <div className="text-left">
        <p className="text-slate-400 text-sm font-medium">{title}</p>
        <p className="text-2xl font-bold text-white mt-0.5">{value}</p>
        {subtitle && <p className="text-xs text-slate-500 mt-1">{subtitle}</p>}
      </div>
    </div>
  )
}

function HealthBadge({ service, status }) {
  const ok = status === 'healthy'
  return (
    <div className={`flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-medium ${ok ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'}`}>
      <span className={`w-2 h-2 rounded-full ${ok ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></span>
      {service}
    </div>
  )
}

function CreateTaskModal({ open, onClose, onCreated }) {
  const [form, setForm] = useState({
    title: '', type: 'send_email', payload: '', cron_expr: '*/2 * * * *', priority: 5
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const payloadTemplates = {
    send_email: JSON.stringify({ to: "test@example.com", from: "system@taskmanager.local", body: "Hello from TaskManager!" }, null, 2),
    send_webhook: JSON.stringify({ url: "https://httpbin.org/post", method: "POST", body: { message: "Hello from TaskManager!" } }, null, 2),
  }

  useEffect(() => {
    if (open) {
      setForm(f => ({ ...f, payload: payloadTemplates[f.type] || '' }))
      setError('')
    }
  }, [open])

  const handleTypeChange = (type) => {
    setForm(f => ({ ...f, type, payload: payloadTemplates[type] || '' }))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await fetch(`${API_BASE}/api/v1/tasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, priority: parseInt(form.priority) }),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      onCreated()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={onClose}>
      <div className="bg-slate-800 border border-slate-700 rounded-2xl p-6 w-full max-w-lg mx-4 shadow-2xl" onClick={e => e.stopPropagation()}>
        <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
          <Plus size={20} className="text-indigo-400" /> Create Task
        </h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-slate-400 mb-1">Title</label>
            <input className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-indigo-500" value={form.title} onChange={e => setForm(f => ({...f, title: e.target.value}))} required placeholder="My Task" />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Type</label>
              <select className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-indigo-500" value={form.type} onChange={e => handleTypeChange(e.target.value)}>
                <option value="send_email">Email</option>
                <option value="send_webhook">Webhook</option>
              </select>
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Priority (1-10)</label>
              <input type="number" min={1} max={10} className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-indigo-500" value={form.priority} onChange={e => setForm(f => ({...f, priority: e.target.value}))} />
            </div>
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">Cron Expression</label>
            <input className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white font-mono text-sm focus:outline-none focus:border-indigo-500" value={form.cron_expr} onChange={e => setForm(f => ({...f, cron_expr: e.target.value}))} required />
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">Payload (JSON)</label>
            <textarea rows={5} className="w-full bg-slate-900 border border-slate-600 rounded-lg px-3 py-2 text-white font-mono text-sm focus:outline-none focus:border-indigo-500" value={form.payload} onChange={e => setForm(f => ({...f, payload: e.target.value}))} required />
          </div>
          {error && <p className="text-red-400 text-sm">{error}</p>}
          <div className="flex gap-3 justify-end">
            <button type="button" onClick={onClose} className="px-4 py-2 rounded-lg text-slate-400 hover:text-white transition">Cancel</button>
            <button type="submit" disabled={loading} className="px-5 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-medium transition disabled:opacity-50">
              {loading ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function TaskTable({ tasks, total }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-slate-700">
            <th className="text-left py-3 px-4 text-slate-400 font-medium">Title</th>
            <th className="text-left py-3 px-4 text-slate-400 font-medium">Type</th>
            <th className="text-left py-3 px-4 text-slate-400 font-medium">Status</th>
            <th className="text-left py-3 px-4 text-slate-400 font-medium">Priority</th>
            <th className="text-left py-3 px-4 text-slate-400 font-medium">Retries</th>
            <th className="text-left py-3 px-4 text-slate-400 font-medium">Next Run</th>
            <th className="text-left py-3 px-4 text-slate-400 font-medium">Created</th>
          </tr>
        </thead>
        <tbody>
          {tasks.map((t, i) => {
            const StatusIcon = STATUS_ICONS[t.status] || Clock
            return (
              <tr key={t.id || i} className="border-b border-slate-700/50 hover:bg-slate-800/50 transition">
                <td className="py-3 px-4 text-white font-medium">{t.title}</td>
                <td className="py-3 px-4">
                  <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-slate-700 text-slate-300">
                    {t.type}
                  </span>
                </td>
                <td className="py-3 px-4">
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold" style={{ backgroundColor: `${STATUS_COLORS[t.status]}20`, color: STATUS_COLORS[t.status] }}>
                    <StatusIcon size={12} /> {t.status}
                  </span>
                </td>
                <td className="py-3 px-4">
                  <span className="inline-flex items-center justify-center w-7 h-7 rounded-full text-xs font-bold text-white" style={{ backgroundColor: PRIORITY_COLORS[(t.priority - 1) % 10] }}>
                    {t.priority}
                  </span>
                </td>
                <td className="py-3 px-4 text-slate-400">{t.retry_count}/{t.max_retries}</td>
                <td className="py-3 px-4 text-slate-400 text-xs">{formatDate(t.next_run_at)}</td>
                <td className="py-3 px-4 text-slate-400 text-xs">{formatDate(t.created_at)}</td>
              </tr>
            )
          })}
          {tasks.length === 0 && (
            <tr><td colSpan={7} className="py-8 text-center text-slate-500">No tasks found</td></tr>
          )}
        </tbody>
      </table>
      {total > 0 && <div className="text-right py-2 px-4 text-xs text-slate-500">Total tasks: {total}</div>}
    </div>
  )
}

function App() {
  const [stats, setStats] = useState(null)
  const [tasks, setTasks] = useState([])
  const [totalTasks, setTotalTasks] = useState(0)
  const [health, setHealth] = useState(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [lastUpdate, setLastUpdate] = useState(null)
  const [statusFilter, setStatusFilter] = useState('')
  const [taskLog, setTaskLog] = useState([])

  const fetchAll = useCallback(async () => {
    try {
      const [statsRes, tasksRes, healthRes] = await Promise.all([
        fetch(`${API_BASE}/api/v1/dashboard/stats`),
        fetch(`${API_BASE}/api/v1/dashboard/tasks?limit=50${statusFilter ? `&status=${statusFilter}` : ''}`),
        fetch(`${API_BASE}/health`),
      ])
      if (statsRes.ok) setStats(await statsRes.json())
      if (tasksRes.ok) {
        const data = await tasksRes.json()
        setTasks(data.tasks || [])
        setTotalTasks(data.total || 0)
      }
      if (healthRes.ok) setHealth(await healthRes.json())
      setLastUpdate(new Date())
    } catch (err) {
      console.error('Fetch error:', err)
    }
  }, [statusFilter])

  useEffect(() => {
    fetchAll()
  }, [fetchAll])

  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(fetchAll, 3000)
    return () => clearInterval(interval)
  }, [autoRefresh, fetchAll])

  useEffect(() => {
    if (!stats) return
    const now = new Date().toLocaleTimeString('ru-RU')
    setTaskLog(prev => {
      const entry = { time: now, completed: stats.completed_tasks, failed: stats.failed_tasks, running: stats.running_tasks, pending: stats.pending_tasks }
      const next = [...prev, entry]
      return next.length > 20 ? next.slice(-20) : next
    })
  }, [stats])

  const statusPieData = stats ? [
    { name: 'Pending', value: stats.pending_tasks, color: STATUS_COLORS.pending },
    { name: 'Scheduled', value: stats.scheduled_tasks, color: STATUS_COLORS.scheduled },
    { name: 'Running', value: stats.running_tasks, color: STATUS_COLORS.running },
    { name: 'Completed', value: stats.completed_tasks, color: STATUS_COLORS.completed },
    { name: 'Failed', value: stats.failed_tasks, color: STATUS_COLORS.failed },
  ].filter(d => d.value > 0) : []

  const typeBarData = stats ? Object.entries(stats.tasks_by_type || {}).map(([name, value]) => ({ name, value })) : []
  const priorityBarData = stats ? Object.entries(stats.tasks_by_priority || {}).map(([p, v]) => ({ priority: `P${p}`, value: v })).sort((a,b) => a.priority.localeCompare(b.priority)) : []

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-900 to-indigo-950 text-white">
      <header className="border-b border-slate-800 bg-slate-900/80 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-indigo-600 rounded-lg">
              <LayoutDashboard size={22} />
            </div>
            <div>
              <h1 className="text-lg font-bold tracking-tight">TaskManager</h1>
              <p className="text-xs text-slate-400">Task Scheduler Dashboard</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            {health && health.services && Object.entries(health.services).map(([svc, st]) => (
              <HealthBadge key={svc} service={svc} status={st} />
            ))}
            <div className="flex items-center gap-2 text-xs text-slate-500">
              <button onClick={() => setAutoRefresh(v => !v)} className={`px-3 py-1.5 rounded-lg border text-xs font-medium transition ${autoRefresh ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400' : 'bg-slate-800 border-slate-700 text-slate-400'}`}>
                {autoRefresh ? 'Live' : 'Paused'}
              </button>
              <button onClick={fetchAll} className="p-1.5 rounded-lg hover:bg-slate-800 transition text-slate-400 hover:text-white">
                <RefreshCw size={16} />
              </button>
            </div>
            {lastUpdate && <span className="text-xs text-slate-600">{lastUpdate.toLocaleTimeString('ru-RU')}</span>}
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-6 space-y-6">
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          <StatCard title="Total" value={stats?.total_tasks ?? '\u2014'} icon={Server} color="bg-slate-600" subtitle="tasks in system" />
          <StatCard title="Pending" value={stats?.pending_tasks ?? '\u2014'} icon={Clock} color="bg-amber-600" />
          <StatCard title="Scheduled" value={stats?.scheduled_tasks ?? '\u2014'} icon={ListChecks} color="bg-blue-600" />
          <StatCard title="Running" value={stats?.running_tasks ?? '\u2014'} icon={Activity} color="bg-purple-600" />
          <StatCard title="Completed" value={stats?.completed_tasks ?? '\u2014'} icon={CheckCircle} color="bg-emerald-600" />
          <StatCard title="Failed" value={stats?.failed_tasks ?? '\u2014'} icon={XCircle} color="bg-red-600" />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-slate-800/50 border border-slate-700/50 rounded-xl p-5">
            <h3 className="text-sm font-semibold text-slate-300 mb-4 flex items-center gap-2"><TrendingUp size={16} className="text-indigo-400" /> Status Distribution</h3>
            <ResponsiveContainer width="100%" height={220}>
              <PieChart>
                <Pie data={statusPieData} cx="50%" cy="50%" innerRadius={50} outerRadius={80} dataKey="value" label={({name, value}) => `${name}: ${value}`}>
                  {statusPieData.map((entry, i) => <Cell key={i} fill={entry.color} />)}
                </Pie>
                <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: 8, color: '#fff' }} />
              </PieChart>
            </ResponsiveContainer>
          </div>

          <div className="bg-slate-800/50 border border-slate-700/50 rounded-xl p-5">
            <h3 className="text-sm font-semibold text-slate-300 mb-4 flex items-center gap-2"><Database size={16} className="text-indigo-400" /> By Task Type</h3>
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={typeBarData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 12 }} />
                <YAxis tick={{ fill: '#94a3b8', fontSize: 12 }} />
                <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: 8, color: '#fff' }} />
                <Bar dataKey="value" fill="#6366f1" radius={[4,4,0,0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>

          <div className="bg-slate-800/50 border border-slate-700/50 rounded-xl p-5">
            <h3 className="text-sm font-semibold text-slate-300 mb-4 flex items-center gap-2"><Zap size={16} className="text-indigo-400" /> By Priority</h3>
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={priorityBarData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="priority" tick={{ fill: '#94a3b8', fontSize: 12 }} />
                <YAxis tick={{ fill: '#94a3b8', fontSize: 12 }} />
                <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: 8, color: '#fff' }} />
                <Bar dataKey="value" fill="#8b5cf6" radius={[4,4,0,0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {taskLog.length > 1 && (
          <div className="bg-slate-800/50 border border-slate-700/50 rounded-xl p-5">
            <h3 className="text-sm font-semibold text-slate-300 mb-4 flex items-center gap-2"><Activity size={16} className="text-indigo-400" /> Real-time Activity</h3>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={taskLog}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="time" tick={{ fill: '#94a3b8', fontSize: 10 }} />
                <YAxis tick={{ fill: '#94a3b8', fontSize: 12 }} />
                <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: 8, color: '#fff' }} />
                <Legend />
                <Bar dataKey="completed" name="Completed" stackId="a" fill="#10b981" />
                <Bar dataKey="running" name="Running" stackId="a" fill="#8b5cf6" />
                <Bar dataKey="pending" name="Pending" stackId="a" fill="#f59e0b" />
                <Bar dataKey="failed" name="Failed" stackId="a" fill="#ef4444" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}

        <div className="bg-slate-800/50 border border-slate-700/50 rounded-xl">
          <div className="flex items-center justify-between p-5 border-b border-slate-700/50">
            <h3 className="text-sm font-semibold text-slate-300 flex items-center gap-2"><ListChecks size={16} className="text-indigo-400" /> Tasks</h3>
            <div className="flex items-center gap-3">
              <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className="bg-slate-900 border border-slate-600 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none">
                <option value="">All statuses</option>
                <option value="pending">Pending</option>
                <option value="scheduled">Scheduled</option>
                <option value="running">Running</option>
                <option value="completed">Completed</option>
                <option value="failed">Failed</option>
              </select>
              <button onClick={() => setModalOpen(true)} className="flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium transition">
                <Plus size={16} /> Create Task
              </button>
            </div>
          </div>
          <TaskTable tasks={tasks} total={totalTasks} />
        </div>
      </main>

      <CreateTaskModal open={modalOpen} onClose={() => setModalOpen(false)} onCreated={fetchAll} />
    </div>
  )
}

export default App
