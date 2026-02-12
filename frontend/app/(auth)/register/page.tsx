'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { api } from '@/lib/api';
import { useAuthStore } from '@/lib/store';

export default function RegisterPage() {
  const router = useRouter();
  const { setAuth } = useAuthStore();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [role, setRole] = useState<'USER' | 'CREATOR'>('USER');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setLoading(true);

    try {
      const response = await api.register({
        email,
        password,
        name,
        role: role === 'CREATOR' ? 'CREATOR' : undefined,
      });
      setAuth(response.user, response.token);
      router.push('/');
    } catch (err: any) {
      setError(err.message || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="flex justify-center mb-8">
          <Link href="/" className="flex items-center gap-2 group">
            <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center text-white font-bold text-xl group-hover:bg-primary/90 transition-colors">
              F
            </div>
            <span className="text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
              Film<span className="text-primary">Tube</span>
            </span>
          </Link>
        </div>

        {/* Register Form */}
        <div className="bg-white dark:bg-surface-dark rounded-xl p-8 border border-gray-200 dark:border-white/5 shadow-sm">
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
            Create Account
          </h1>
          <p className="text-slate-500 dark:text-slate-400 mb-6">
            Join FilmTube to share your films with the world
          </p>

          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-500 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                Name
              </label>
              <input
                id="name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                className="w-full bg-gray-100 dark:bg-white/5 border border-gray-300 dark:border-white/10 rounded-lg py-2.5 px-4 text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary dark:text-white transition-colors"
                placeholder="Your name"
              />
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                className="w-full bg-gray-100 dark:bg-white/5 border border-gray-300 dark:border-white/10 rounded-lg py-2.5 px-4 text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary dark:text-white transition-colors"
                placeholder="your@email.com"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={8}
                className="w-full bg-gray-100 dark:bg-white/5 border border-gray-300 dark:border-white/10 rounded-lg py-2.5 px-4 text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary dark:text-white transition-colors"
                placeholder="Min. 8 characters"
              />
            </div>

            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                Confirm Password
              </label>
              <input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                minLength={8}
                className="w-full bg-gray-100 dark:bg-white/5 border border-gray-300 dark:border-white/10 rounded-lg py-2.5 px-4 text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary dark:text-white transition-colors"
                placeholder="Confirm your password"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                Account Type
              </label>
              <div className="flex gap-4">
                <label className="flex-1 flex items-center gap-2 p-3 bg-gray-100 dark:bg-white/5 rounded-lg cursor-pointer border-2 border-transparent has-[:checked]:border-primary">
                  <input
                    type="radio"
                    name="role"
                    value="USER"
                    checked={role === 'USER'}
                    onChange={() => setRole('USER')}
                    className="accent-primary"
                  />
                  <span className="text-sm text-slate-700 dark:text-slate-300">Viewer</span>
                </label>
                <label className="flex-1 flex items-center gap-2 p-3 bg-gray-100 dark:bg-white/5 rounded-lg cursor-pointer border-2 border-transparent has-[:checked]:border-primary">
                  <input
                    type="radio"
                    name="role"
                    value="CREATOR"
                    checked={role === 'CREATOR'}
                    onChange={() => setRole('CREATOR')}
                    className="accent-primary"
                  />
                  <span className="text-sm text-slate-700 dark:text-slate-300">Creator</span>
                </label>
              </div>
              <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">
                Creators can upload films to the platform
              </p>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-primary hover:bg-red-700 text-white font-semibold py-2.5 px-4 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {loading ? (
                <>
                  <span className="material-icons text-sm animate-spin">refresh</span>
                  Creating account...
                </>
              ) : (
                'Create Account'
              )}
            </button>
          </form>

          <div className="mt-6 text-center text-sm text-slate-500 dark:text-slate-400">
            Already have an account?{' '}
            <Link href="/login" className="text-primary hover:text-red-400 font-medium transition-colors">
              Sign in
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
