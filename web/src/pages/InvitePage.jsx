import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useApi } from '../context/ApiContext';
import { UserPlus2 } from 'lucide-react';

export default function InvitePage() {
  const { token } = useParams();
  const { user, loading: authLoading, refreshToken } = useAuth();
  const { post } = useApi();
  const navigate = useNavigate();
  const [error, setError] = useState('');
  const [joining, setJoining] = useState(false);

  useEffect(() => {
    if (!authLoading && !user) {
      // Save invite token and redirect to register
      localStorage.setItem('pending_invite', token);
      navigate('/register');
    }
  }, [authLoading, user]);

  const handleJoin = async () => {
    setJoining(true);
    setError('');
    try {
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
      setJoining(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-900">
        <div className="w-8 h-8 border-2 border-indigo-400 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      <div className="w-full max-w-md p-8 bg-slate-800/50 backdrop-blur-sm rounded-2xl border border-slate-700 shadow-2xl text-center">
        <UserPlus2 className="w-12 h-12 text-emerald-400 mx-auto mb-4" />
        <h1 className="text-xl font-bold text-white mb-2">Join Company</h1>
        <p className="text-slate-400 text-sm mb-6">
          You've been invited to join a team on TaskManager.
        </p>

        {error && (
          <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <button
          onClick={handleJoin}
          disabled={joining}
          className="w-full py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
        >
          {joining ? 'Joining...' : 'Accept Invite'}
        </button>
      </div>
    </div>
  );
}
