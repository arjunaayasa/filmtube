import Link from 'next/link';
import { type Film } from '@/lib/api';

interface FilmCardProps {
  film: Film;
}

export function FilmCard({ film }: FilmCardProps) {
  const duration = formatDuration(film.duration);
  const timeAgo = formatTimeAgo(film.published_at || film.created_at);

  return (
    <Link
      href={`/films/${film.id}`}
      className="group cursor-pointer flex flex-col gap-3"
    >
      <div className="relative aspect-video rounded-xl overflow-hidden bg-gray-800">
        {film.thumbnail_url ? (
          <img
            src={film.thumbnail_url}
            alt={film.title}
            className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-surface-dark to-background-dark">
            <span className="material-icons text-4xl text-gray-700">movie</span>
          </div>
        )}
        <div className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-medium px-1.5 py-0.5 rounded">
          {duration}
        </div>
        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors" />
      </div>

      <div className="flex gap-3 items-start">
        <div className="w-9 h-9 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold text-xs mt-0.5">
          {(film.created_by?.name || 'U').charAt(0).toUpperCase()}
        </div>
        <div className="flex flex-col">
          <h4 className="text-white font-semibold text-base leading-tight line-clamp-2 group-hover:text-primary transition-colors">
            {film.title}
          </h4>
          <div className="text-gray-400 text-sm mt-1 flex flex-col">
            <span className="hover:text-white transition-colors">
              {film.created_by?.name || 'Unknown Creator'}
            </span>
            <span className="text-xs mt-0.5">
              {formatNumber(film.view_count)} views • {timeAgo}
            </span>
          </div>
        </div>
      </div>
    </Link>
  );
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }
  return `${minutes}:${secs.toString().padStart(2, '0')}`;
}

function formatTimeAgo(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)} minutes ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} hours ago`;
  if (seconds < 604800) return `${Math.floor(seconds / 86400)} days ago`;
  if (seconds < 2592000) return `${Math.floor(seconds / 604800)} weeks ago`;
  if (seconds < 31536000) return `${Math.floor(seconds / 2592000)} months ago`;
  return `${Math.floor(seconds / 31536000)} years ago`;
}

function formatNumber(num: number): string {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
  return num.toString();
}
