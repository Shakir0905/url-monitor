import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { auth } from '../api';

export default function Register() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const { data } = await auth.register(email, password);
      localStorage.setItem('token', data.token);
      localStorage.setItem('user_id', data.user_id);
      navigate('/dashboard');
    } catch (err) {
      setError(err.response?.data?.error || 'Registration failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-4 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-purple-950/30 via-transparent to-indigo-950/20"></div>
      <div className="absolute top-1/4 -right-20 w-96 h-96 bg-purple-600/20 rounded-full blur-3xl"></div>
      <div className="absolute bottom-1/4 -left-20 w-96 h-96 bg-pink-600/20 rounded-full blur-3xl"></div>

      <div className="max-w-md w-full glass rounded-2xl p-8 fade-in relative z-10">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl gradient-bg glow mb-4">
            <svg className="w-7 h-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold gradient-text mb-2">Create account</h1>
          <p className="text-sm text-zinc-400">Start monitoring in 30 seconds</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-2 uppercase tracking-wider">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              className="w-full px-4 py-3 bg-zinc-900/50 border border-zinc-800 rounded-xl text-white placeholder-zinc-600 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-zinc-400 mb-2 uppercase tracking-wider">Password (8+ chars)</label>
            <input
              type="password"
              required
              minLength={8}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full px-4 py-3 bg-zinc-900/50 border border-zinc-800 rounded-xl text-white placeholder-zinc-600 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition"
            />
          </div>
          {error && (
            <div className="px-4 py-3 bg-red-950/50 border border-red-900/50 rounded-xl text-sm text-red-400">
              {error}
            </div>
          )}
          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 gradient-bg text-white font-semibold rounded-xl hover:opacity-90 disabled:opacity-50 transition-all glow"
          >
            {loading ? 'Creating...' : 'Create account'}
          </button>
        </form>

        <div className="mt-6 text-center text-sm text-zinc-500">
          Have account? <Link to="/login" className="text-indigo-400 hover:text-indigo-300 font-medium">Sign in</Link>
        </div>
      </div>
    </div>
  );
}
