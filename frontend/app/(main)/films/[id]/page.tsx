import { Navigation } from '@/components/Navigation';
import { VideoPlayer } from '@/components/VideoPlayer';
import { api, type Film } from '@/lib/api';
import { notFound } from 'next/navigation';

async function getFilm(id: string): Promise<Film> {
  try {
    return await api.getFilm(id);
  } catch (error) {
    return notFound();
  }
}

async function getPlaybackURL(id: string) {
  try {
    return await api.getPlaybackURL(id);
  } catch (error) {
    return null;
  }
}

export default async function FilmPage({ params }: { params: { id: string } }) {
  const film = await getFilm(params.id);
  const playbackInfo = await getPlaybackURL(params.id);

  const isReady = film.status === 'READY';

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark">
      <Navigation />

      <main className="max-w-[1600px] mx-auto px-4 py-6 w-full">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
          {/* Left Column: Player & Info (Span 8-9 cols) */}
          <div className="lg:col-span-8 xl:col-span-9 flex flex-col gap-6">
            {/* Video Player Component */}
            {isReady && playbackInfo?.hls_master_url ? (
              <VideoPlayer
                src={playbackInfo.hls_master_url}
                poster={playbackInfo.thumbnail_url}
                title={film.title}
              />
            ) : (
              <div className="w-full aspect-video bg-black rounded-xl overflow-hidden flex items-center justify-center">
                <div className="text-center text-white">
                  <span className="material-icons text-6xl mb-4">video_library</span>
                  <p className="text-xl font-semibold mb-2">
                    {film.status === 'TRANSCODING' ? 'Processing...' : 'Not Ready'}
                  </p>
                  <p className="text-sm text-gray-400">
                    {film.status === 'TRANSCODING'
                      ? 'This film is being processed. Please check back later.'
                      : 'This film is not yet available for playback.'}
                  </p>
                </div>
              </div>
            )}

            {/* Action Bar */}
            <div className="info-section flex flex-col md:flex-row md:items-start justify-between gap-4 border-b border-gray-200 dark:border-white/5 pb-6">
              <div className="flex-1">
                <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2 leading-tight">
                  {film.title}
                </h1>
                <div className="flex flex-wrap items-center gap-3 text-sm text-slate-500 dark:text-slate-400">
                  <span className="bg-gray-200 dark:bg-white/10 px-2 py-0.5 rounded text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wide">
                    {film.type === 'SHORT_FILM' ? 'Short' : 'Feature'}
                  </span>
                  <span>•</span>
                  <span>{new Date(film.created_at).getFullYear()}</span>
                  <span>•</span>
                  <span>{formatDuration(film.duration)}</span>
                  <span>•</span>
                  <span>{formatNumber(film.view_count)} views</span>
                </div>
              </div>
              <div className="flex items-center gap-3 self-start mt-2 md:mt-0">
                <button className="w-10 h-10 rounded-full border border-gray-200 dark:border-white/10 flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-gray-100 dark:hover:bg-white/5 hover:text-primary dark:hover:text-primary transition-colors">
                  <span className="material-icons-outlined">bookmark_border</span>
                </button>
                <button className="w-10 h-10 rounded-full border border-gray-200 dark:border-white/10 flex items-center justify-center text-slate-500 dark:text-slate-400 hover:bg-gray-100 dark:hover:bg-white/5 hover:text-primary dark:hover:text-primary transition-colors">
                  <span className="material-icons-outlined">share</span>
                </button>
              </div>
            </div>

            {/* Tabs & Content */}
            <div className="info-section">
              {/* Tabs Header */}
              <div className="flex gap-8 border-b border-gray-200 dark:border-white/5 mb-6">
                <button className="pb-3 border-b-2 border-primary text-primary font-medium text-sm tracking-wide">
                  Synopsis
                </button>
                <button className="pb-3 border-b-2 border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 font-medium text-sm tracking-wide transition-colors">
                  Director&apos;s Notes
                </button>
                <button className="pb-3 border-b-2 border-transparent text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200 font-medium text-sm tracking-wide transition-colors">
                  Cast & Crew
                </button>
              </div>

              {/* Content Body */}
              <div className="text-slate-600 dark:text-slate-300 leading-relaxed">
                {film.description ? (
                  <p>{film.description}</p>
                ) : (
                  <p className="text-slate-500 dark:text-slate-500 italic">
                    No description available.
                  </p>
                )}
              </div>
            </div>
          </div>

          {/* Right Column: Sidebar (Span 4-3 cols) */}
          <aside className="sidebar-section lg:col-span-4 xl:col-span-3 space-y-8">
            {/* Creator Profile Card */}
            <div className="bg-white dark:bg-surface-dark rounded-xl p-5 border border-gray-200 dark:border-white/5 shadow-sm">
              <div className="flex items-center gap-4 mb-4">
                <div className="w-14 h-14 rounded-full overflow-hidden ring-2 ring-primary/20 bg-primary/10 flex items-center justify-center text-primary font-bold">
                  {(film.created_by?.name || 'U').charAt(0).toUpperCase()}
                </div>
                <div>
                  <h3 className="font-bold text-slate-900 dark:text-white">
                    {film.created_by?.name || 'Unknown Creator'}
                  </h3>
                  <p className="text-xs text-slate-500 dark:text-slate-400">Filmmaker</p>
                </div>
              </div>
            </div>

            {/* Film Details */}
            <div>
              <h3 className="text-sm font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-4">
                Film Details
              </h3>
              <div className="bg-gray-100 dark:bg-surface-dark rounded-lg p-5 border border-gray-200 dark:border-white/5 space-y-3 text-sm">
                <div className="flex justify-between">
                  <span className="text-slate-500 dark:text-slate-400">Type</span>
                  <span className="font-medium">{film.type === 'SHORT_FILM' ? 'Short Film' : 'Feature Film'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-500 dark:text-slate-400">Duration</span>
                  <span className="font-medium">{formatDuration(film.duration)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-500 dark:text-slate-400">Uploaded</span>
                  <span className="font-medium">{new Date(film.created_at).toLocaleDateString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-500 dark:text-slate-400">Status</span>
                  <span className={`font-medium ${
                    film.status === 'READY' ? 'text-green-500' :
                    film.status === 'TRANSCODING' ? 'text-yellow-500' :
                    'text-gray-500'
                  }`}>
                    {film.status}
                  </span>
                </div>
              </div>
            </div>
          </aside>
        </div>
      </main>
    </div>
  );
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m ${secs}s`;
}

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
  return num.toString();
}
