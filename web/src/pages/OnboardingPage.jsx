import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useApi } from '../context/ApiContext';
import { Building2, UserPlus2, Link as LinkIcon } from 'lucide-react';

export default function OnboardingPage() {
  const [mode, setMode] = useState(null); // 'create' | 'join'
  const [companyName, setCompanyName] = useState('');
  const [inviteToken, setInviteToken] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { refreshToken } = useAuth();
  const { post } = useApi();
  const navigate = useNavigate();

  const handleCreate = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await post('/companies', { name: companyName });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text);
      }
      await refreshToken();
      navigate('/');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleJoin = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      // Extract token from link if full URL pasted
      let token = inviteToken;
      if (token.includes('/invite/')) {
        token = token.split('/invite/').pop();
      }
      const res = await post('/companies/join', { token });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text);
      }
      await refreshToken();
      navigate('/');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  if (!mode) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-mesh p-4">
        <div className="w-full max-w-lg p-8 glass-card rounded-2xl fade-in">
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl mb-4 pulse-glow">
              <Building2 className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-3xl font-bold gradient-text">Welcome to TaskManager</h1>
            <p className="text-slate-400 mt-2">Create a new company or join an existing one</p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <button
              onClick={() => setMode('create')}
              className="flex flex-col items-center gap-3 p-6 bg-slate-700/30 hover:bg-slate-700/50 border border-slate-600 hover:border-indigo-500 rounded-xl transition-all"
            >
              <Building2 className="w-10 h-10 text-indigo-400" />
              <span className="text-white font-medium">Create Company</span>
              <span className="text-slate-400 text-sm text-center">Start a new team workspace</span>
            </button>
            <button
              onClick={() => setMode('join')}
              className="flex flex-col items-center gap-3 p-6 bg-slate-700/30 hover:bg-slate-700/50 border border-slate-600 hover:border-emerald-500 rounded-xl transition-all"
            >
              <UserPlus2 className="w-10 h-10 text-emerald-400" />
              <span className="text-white font-medium">Join Company</span>
              <span className="text-slate-400 text-sm text-center">Use an invite link or token</span>
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-mesh p-4">
      <div className="w-full max-w-md p-8 glass-card rounded-2xl fade-in">
        <button
          onClick={() => { setMode(null); setError(''); }}
          className="text-slate-400 hover:text-white text-sm mb-4 transition-colors"
        >
          ← Back
        </button>

        {error && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        {mode === 'create' && (
          <>
            <div className="text-center mb-6">
              <Building2 className="w-10 h-10 text-indigo-400 mx-auto mb-3" />
              <h2 className="text-xl font-bold text-white">Create Company</h2>
              <p className="text-slate-400 text-sm mt-1">Give your team a name</p>
            </div>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1">Company Name</label>
                <input
                  type="text"
                  value={companyName}
                  onChange={(e) => setCompanyName(e.target.value)}
                  className="w-full px-4 py-2.5 bg-slate-700/50 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                  placeholder="Acme Corp"
                  required
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white font-semibold rounded-lg disabled:opacity-50 flex items-center justify-center gap-2 shadow-lg shadow-indigo-500/30"
              >
                {loading ? 'Creating...' : 'Create Company'}
              </button>
            </form>
          </>
        )}

        {mode === 'join' && (
          <>
            <div className="text-center mb-6">
              <LinkIcon className="w-10 h-10 text-emerald-400 mx-auto mb-3" />
              <h2 className="text-xl font-bold text-white">Join Company</h2>
              <p className="text-slate-400 text-sm mt-1">Paste your invite link or token</p>
            </div>
            <form onSubmit={handleJoin} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-1">Invite Token or Link</label>
                <input
                  type="text"
                  value={inviteToken}
                  onChange={(e) => setInviteToken(e.target.value)}
                  className="w-full px-4 py-2.5 bg-slate-700/50 border border-slate-600 rounded-lg text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                  placeholder="Paste invite link or token..."
                  required
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-600 hover:to-teal-700 text-white font-semibold rounded-lg disabled:opacity-50 flex items-center justify-center gap-2 shadow-lg shadow-emerald-500/30"
              >
                {loading ? 'Joining...' : 'Join Company'}
              </button>
            </form>
          </>
        )}
      </div>
    </div>
  );
}
