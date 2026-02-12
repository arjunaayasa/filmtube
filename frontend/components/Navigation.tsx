'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/lib/store';

export function Navigation() {
  const router = useRouter();
  const { user, isAuthenticated, clearAuth } = useAuthStore();

  const handleLogout = () => {
    clearAuth();
    router.push('/login');
  };

  return (
    <nav className="nav-section sticky top-0 z-40 bg-background-light/90 dark:bg-background-dark/90 backdrop-blur-md border-b border-gray-200 dark:border-white/10 h-16">
      <div className="max-w-[1600px] mx-auto px-4 h-full flex items-center justify-between">
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-2 group">
            <div className="w-8 h-8 rounded bg-primary flex items-center justify-center text-white font-bold text-lg group-hover:bg-primary/90 transition-colors">
              F
            </div>
            <span className="text-xl font-bold tracking-tight text-slate-900 dark:text-white">
              Film<span className="text-primary">Tube</span>
            </span>
          </Link>
          <div className="hidden md:flex items-center gap-6 text-sm font-medium text-slate-500 dark:text-slate-400">
            <Link href="/" className="hover:text-primary transition-colors">
              Discover
            </Link>
            <Link href="/films" className="hover:text-primary transition-colors">
              Films
            </Link>
            {isAuthenticated && (
              <Link href="/upload" className="hover:text-primary transition-colors">
                Upload
              </Link>
            )}
          </div>
        </div>

        <div className="flex items-center gap-4">
          {isAuthenticated ? (
            <>
              <div className="hidden sm:block text-sm text-slate-600 dark:text-slate-300">
                {user?.name}
              </div>
              <div className="w-8 h-8 rounded-full overflow-hidden border border-gray-200 dark:border-white/10 bg-primary/10 flex items-center justify-center">
                <span className="text-sm font-medium text-primary">
                  {user?.name?.charAt(0).toUpperCase()}
                </span>
              </div>
              <button
                onClick={handleLogout}
                className="text-sm text-slate-500 dark:text-slate-400 hover:text-primary transition-colors"
              >
                Logout
              </button>
            </>
          ) : (
            <>
              <Link
                href="/login"
                className="text-sm font-medium text-slate-600 dark:text-slate-300 hover:text-primary transition-colors"
              >
                Login
              </Link>
              <Link
                href="/register"
                className="bg-primary hover:bg-red-700 text-white text-sm font-medium py-2 px-4 rounded transition-colors"
              >
                Sign Up
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
