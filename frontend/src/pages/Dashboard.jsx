import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { urls as urlsApi, analytics } from '../api';

export default function Dashboard() {
  const navigate = useNavigate();
  const [dashboard, setDashboard] = useState(null);
  const [urls, setUrls] = useState([]);
  const [newURL, setNewURL] = useState('');
  const [checkInterval, setCheckInterval] = useState(60);
  const [loading, setLoading] = useState(true);
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    loadData();
    const timer = setInterval(loadData, 10000);
    return () => clearInterval(timer);
  }, []);

  async function loadData() {
    try {
      const [d, u] = await Promise.all([analytics.dashboard(), urlsApi.list()]);
      setDashboard(d.data);
      setUrls(u.data.urls || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }

  async function addURL(e) {
    e.preventDefault();
    if (!newURL) return;
    setAdding(true);
    try {
      await urlsApi.create(newURL, checkInterval);
      setNewURL('');
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || 'Failed to add URL');
    } finally {
      setAdding(false);
    }
  }

  async function removeURL(id) {
    if (!confirm('Delete this URL?')) return;
    await urlsApi.remove(id);
    loadData();
  }

  function logout() {
    localStorage.clear();
    navigate('/login');
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-zinc-500">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen relative">
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-indigo-600/10 rounded-full blur-3xl pointer-events-none"></div>

      <header className="border-b border-zinc-900 backdrop-blur-xl bg-black/30 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl gradient-bg flex items-center justify-center">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <h1 className="text-lg font-bold">URL Monitor</h1>
          </div>
          <button
            onClick={logout}
            className="px-4 py-2 text-sm text-zinc-400 hover:text-white hover:bg-zinc-900 rounded-lg transition"
          >
            Logout
          </button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-10 relative z-10 fade-in">
        <div className="mb-8">
          <h2 className="text-3xl font-bold mb-2">Dashboard</h2>
          <p className="text-zinc-500">Real-time monitoring · updates every 10s</p>
        </div>

        {dashboard && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            <StatCard label="Total URLs" value={dashboard.total_urls || 0} icon="📊" />
            <StatCard label="Active" value={dashboard.active_urls || 0} icon="⚡" />
            <StatCard label="Up" value={dashboard.urls_currently_up || 0} icon="🟢" accent="text-emerald-400" />
            <StatCard label="Down" value={dashboard.urls_currently_down || 0} icon="🔴" accent="text-red-400" />
          </div>
        )}

        <div className="glass rounded-2xl p-6 mb-6">
          <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-indigo-500 pulse-dot"></span>
            Add new URL
          </h3>
          <form onSubmit={addURL} className="flex flex-col md:flex-row gap-3">
            <input
              type="url"
              required
              placeholder="https://example.com"
              value={newURL}
              onChange={(e) => setNewURL(e.target.value)}
              className="flex-1 px-4 py-3 bg-zinc-900/50 border border-zinc-800 rounded-xl text-white placeholder-zinc-600 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition"
            />
            <div className="flex items-center gap-2">
              <input
                type="number"
                min="10"
                max="3600"
                value={checkInterval}
                onChange={(e) => setCheckInterval(+e.target.value)}
                className="w-24 px-4 py-3 bg-zinc-900/50 border border-zinc-800 rounded-xl text-white focus:outline-none focus:border-indigo-500 transition"
              />
              <span className="text-sm text-zinc-500">sec</span>
            </div>
            <button
              type="submit"
              disabled={adding}
              className="px-6 py-3 gradient-bg text-white font-semibold rounded-xl hover:opacity-90 disabled:opacity-50 transition glow whitespace-nowrap"
            >
              {adding ? 'Adding...' : 'Add URL'}
            </button>
          </form>
        </div>

        <div className="glass rounded-2xl overflow-hidden">
          <div className="p-6 border-b border-zinc-800/50 flex justify-between items-center">
            <h3 className="text-lg font-semibold">Monitored URLs</h3>
            <span className="text-sm text-zinc-500">{urls.length} total</span>
          </div>
          <div className="divide-y divide-zinc-800/50">
            {urls.length === 0 ? (
              <div className="p-12 text-center">
                <div className="text-4xl mb-3">🚀</div>
                <div className="text-zinc-400 font-medium mb-1">No URLs yet</div>
                <div className="text-sm text-zinc-600">Add your first URL above to start monitoring</div>
              </div>
            ) : (
              urls.map((u) => (
                <div key={u.id} className="p-5 flex items-center justify-between hover:bg-zinc-900/30 transition group">
                  <div className="flex items-center gap-4 min-w-0 flex-1">
                    <div className={`w-2.5 h-2.5 rounded-full flex-shrink-0 ${u.is_active ? 'bg-emerald-500 pulse-dot' : 'bg-zinc-600'}`}></div>
                    <div className="min-w-0 flex-1">
                      <div className="font-medium truncate">{u.url}</div>
                      <div className="text-xs text-zinc-500 mt-1 flex items-center gap-2">
                        <span>Every {u.check_interval_seconds}s</span>
                        <span className="text-zinc-700">·</span>
                        <span className={u.is_active ? 'text-emerald-500' : 'text-zinc-500'}>
                          {u.is_active ? 'Active' : 'Paused'}
                        </span>
                      </div>
                    </div>
                  </div>
                  <button
                    onClick={() => removeURL(u.id)}
                    className="px-3 py-1.5 text-xs text-zinc-500 hover:text-red-400 hover:bg-red-950/30 rounded-lg transition opacity-0 group-hover:opacity-100"
                  >
                    Delete
                  </button>
                </div>
              ))
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

function StatCard({ label, value, icon, accent = '' }) {
  return (
    <div className="glass rounded-2xl p-5 hover:border-zinc-700 transition group">
      <div className="flex items-center justify-between mb-3">
        <span className="text-xs text-zinc-500 uppercase tracking-wider font-medium">{label}</span>
        <span className="text-lg opacity-50 group-hover:opacity-100 transition">{icon}</span>
      </div>
      <div className={`text-3xl font-bold ${accent || 'text-white'}`}>{value}</div>
    </div>
  );
}
