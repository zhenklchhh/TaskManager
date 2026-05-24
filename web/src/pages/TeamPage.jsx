import { useState, useEffect } from 'react';
import { useApi } from '../context/ApiContext';
import { useAuth } from '../context/AuthContext';
import { Users, UserMinus, Link as LinkIcon, Mail, Copy, Check } from 'lucide-react';

export default function TeamPage() {
  const [members, setMembers] = useState([]);
  const [invites, setInvites] = useState([]);
  const [inviteEmail, setInviteEmail] = useState('');
  const [copied, setCopied] = useState(null);
  const [loading, setLoading] = useState(true);
  const { get, post } = useApi();
  const { user } = useAuth();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [membersRes, invitesRes] = await Promise.all([
        get('/companies/members'),
        get('/companies/invites'),
      ]);
      if (membersRes.ok) setMembers(await membersRes.json());
      if (invitesRes.ok) setInvites(await invitesRes.json());
    } finally {
      setLoading(false);
    }
  };

  const createInviteLink = async () => {
    const res = await post('/companies/invites/link', {});
    if (res.ok) {
      const invite = await res.json();
      setInvites((prev) => [invite, ...prev]);
    }
  };

  const createEmailInvite = async (e) => {
    e.preventDefault();
    if (!inviteEmail) return;
    const res = await post('/companies/invites/email', { email: inviteEmail });
    if (res.ok) {
      const invite = await res.json();
      setInvites((prev) => [invite, ...prev]);
      setInviteEmail('');
    }
  };

  const removeMember = async (userId) => {
    const res = await post('/companies/members/remove', { user_id: userId });
    if (res.ok) {
      setMembers((prev) => prev.filter((m) => m.user_id !== userId));
    }
  };

  const copyLink = (token) => {
    const link = `${window.location.origin}/invite/${token}`;
    navigator.clipboard.writeText(link);
    setCopied(token);
    setTimeout(() => setCopied(null), 2000);
  };

  const isOwner = members.some((m) => m.user_id === user?.id && m.role === 'owner');

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center h-full">
        <div className="text-slate-400">Loading...</div>
      </div>
    );
  }

  return (
    <div className="p-8 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold gradient-text flex items-center gap-3">
            <Users className="w-8 h-8 text-cyan-400" />
            Team Members
          </h1>
          <p className="text-slate-400 mt-1">Manage your team and invite new members · {members.length} {members.length === 1 ? 'member' : 'members'}</p>
        </div>
      </div>

      {/* Members list */}
      <div className="glass-card rounded-2xl overflow-hidden mb-8">
        {members.map((member) => (
          <div
            key={member.id}
            className="flex items-center justify-between px-6 py-4 border-b border-slate-700/30 last:border-0 hover:bg-slate-700/10 transition-colors"
          >
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white text-sm font-bold shadow-md">
                {member.name?.[0]?.toUpperCase() || '?'}
              </div>
              <div>
                <div className="text-sm font-semibold text-white">{member.name}</div>
                <div className="text-xs text-slate-400">{member.email}</div>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <span className={`text-xs px-3 py-1 rounded-full font-medium ${
                member.role === 'owner'
                  ? 'bg-gradient-to-r from-amber-500/20 to-orange-500/20 text-amber-300 border border-amber-500/30'
                  : 'bg-slate-700/50 text-slate-300 border border-slate-600/30'
              }`}>
                {member.role}
              </span>
              {isOwner && member.user_id !== user?.id && member.role !== 'owner' && (
                <button
                  onClick={() => removeMember(member.user_id)}
                  className="p-1.5 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
                  title="Remove member"
                >
                  <UserMinus className="w-4 h-4" />
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Invite section */}
      <h2 className="text-xl font-bold text-white mb-4">Invite People</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5 mb-6">
        <div className="glass-card glass-card-hover rounded-2xl p-6">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-xl bg-indigo-500/20 flex items-center justify-center">
              <LinkIcon className="w-5 h-5 text-indigo-300" />
            </div>
            <div>
              <div className="text-base font-semibold text-white">Invite Link</div>
              <div className="text-xs text-slate-400">Shareable link, 7 days valid</div>
            </div>
          </div>
          <button
            onClick={createInviteLink}
            className="w-full py-2.5 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white text-sm font-semibold rounded-lg shadow-md shadow-indigo-500/30"
          >
            Generate Link
          </button>
        </div>

        <div className="glass-card glass-card-hover rounded-2xl p-6">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-xl bg-emerald-500/20 flex items-center justify-center">
              <Mail className="w-5 h-5 text-emerald-300" />
            </div>
            <div>
              <div className="text-base font-semibold text-white">Email Invite</div>
              <div className="text-xs text-slate-400">Send invite directly to email</div>
            </div>
          </div>
          <form onSubmit={createEmailInvite} className="space-y-3">
            <input
              type="email"
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
              className="w-full px-3 py-2 bg-slate-700/50 border border-slate-600 rounded-lg text-white text-sm placeholder-slate-400"
              placeholder="colleague@example.com"
            />
            <button
              type="submit"
              className="w-full py-2.5 bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-600 hover:to-teal-700 text-white text-sm font-semibold rounded-lg shadow-md shadow-emerald-500/30"
            >
              Send Invite
            </button>
          </form>
        </div>
      </div>

      {/* Existing invites */}
      {invites.length > 0 && (
        <div className="glass-card rounded-2xl overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-700/30">
            <h3 className="text-base font-semibold text-white">Pending Invites</h3>
          </div>
          {invites.filter(i => !i.used).map((invite) => (
            <div key={invite.id} className="flex items-center justify-between px-5 py-3 border-b border-slate-700/50 last:border-0">
              <div className="text-sm text-slate-300">
                {invite.email || 'Link invite'}
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-slate-500">
                  Expires {new Date(invite.expires_at).toLocaleDateString()}
                </span>
                <button
                  onClick={() => copyLink(invite.token)}
                  className="p-1.5 text-slate-400 hover:text-indigo-400 hover:bg-indigo-500/10 rounded-lg transition-colors"
                  title="Copy invite link"
                >
                  {copied === invite.token ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
