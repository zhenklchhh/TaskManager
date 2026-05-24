import { useState, useEffect } from 'react';
import { useApi } from '../context/ApiContext';
import { useAuth } from '../context/AuthContext';
import { Settings, Building2, ExternalLink, User, Activity } from 'lucide-react';

export default function SettingsPage() {
  const [company, setCompany] = useState(null);
  const [loading, setLoading] = useState(true);
  const { get } = useApi();
  const { user } = useAuth();

  useEffect(() => {
    loadCompany();
  }, []);

  const loadCompany = async () => {
    try {
      const res = await get('/companies/current');
      if (res.ok) setCompany(await res.json());
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center h-full">
        <div className="text-slate-400">Loading...</div>
      </div>
    );
  }

  return (
    <div className="p-8 max-w-3xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold gradient-text flex items-center gap-3">
          <Settings className="w-8 h-8 text-indigo-400" />
          Settings
        </h1>
        <p className="text-slate-400 mt-1">Manage your account and company preferences</p>
      </div>

      {/* Company info */}
      <div className="glass-card rounded-2xl p-6 mb-6">
        <div className="flex items-center gap-3 mb-5">
          <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-md shadow-indigo-500/20">
            <Building2 className="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white">Company</h2>
            <p className="text-xs text-slate-400">Your organization details</p>
          </div>
        </div>
        {company ? (
          <div className="space-y-4 ml-1">
            <div className="flex items-center justify-between py-3 border-b border-slate-700/30">
              <label className="text-sm text-slate-400">Name</label>
              <div className="text-sm font-semibold text-white">{company.name}</div>
            </div>
            <div className="flex items-center justify-between py-3 border-b border-slate-700/30">
              <label className="text-sm text-slate-400">Company ID</label>
              <div className="text-xs text-slate-300 font-mono">{company.id}</div>
            </div>
            <div className="flex items-center justify-between py-3">
              <label className="text-sm text-slate-400">Created</label>
              <div className="text-sm text-slate-300">
                {new Date(company.created_at).toLocaleDateString()}
              </div>
            </div>
          </div>
        ) : (
          <p className="text-sm text-slate-400">No company info available</p>
        )}
      </div>

      {/* User info */}
      <div className="glass-card rounded-2xl p-6 mb-6">
        <div className="flex items-center gap-3 mb-5">
          <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-pink-500 to-purple-600 flex items-center justify-center shadow-md shadow-pink-500/20">
            <User className="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white">Profile</h2>
            <p className="text-xs text-slate-400">Your personal account info</p>
          </div>
        </div>
        <div className="space-y-4 ml-1">
          <div className="flex items-center justify-between py-3 border-b border-slate-700/30">
            <label className="text-sm text-slate-400">Name</label>
            <div className="text-sm font-semibold text-white">{user?.name}</div>
          </div>
          <div className="flex items-center justify-between py-3">
            <label className="text-sm text-slate-400">Email</label>
            <div className="text-sm text-slate-300">{user?.email}</div>
          </div>
        </div>
      </div>

      {/* Grafana link */}
      <div className="glass-card rounded-2xl p-6">
        <div className="flex items-center gap-3 mb-5">
          <div className="w-11 h-11 rounded-xl bg-gradient-to-br from-orange-500 to-red-600 flex items-center justify-center shadow-md shadow-orange-500/20">
            <Activity className="w-5 h-5 text-white" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white">Monitoring</h2>
            <p className="text-xs text-slate-400">Real-time metrics powered by Grafana</p>
          </div>
        </div>
        <a
          href="http://localhost:3001"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-r from-orange-500 to-red-600 hover:from-orange-600 hover:to-red-700 text-white rounded-lg text-sm font-semibold shadow-md shadow-orange-500/30"
        >
          <ExternalLink className="w-4 h-4" />
          Open Grafana Dashboard
        </a>
      </div>
    </div>
  );
}
