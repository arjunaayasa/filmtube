import { Navigation } from '@/components/Navigation';
import { FilmCard } from '@/components/FilmCard';
import { api, type Film } from '@/lib/api';

async function getFilms() {
  try {
    const response = await api.getFilms(1, 20, 'READY');
    return response.films;
  } catch (error) {
    console.error('Failed to fetch films:', error);
    return [];
  }
}

export default async function HomePage() {
  const films = await getFilms();

  return (
    <div className="flex min-h-screen">
      {/* Main Content */}
      <main className="flex-1 overflow-y-auto bg-background-light dark:bg-background-dark">
        <Navigation />

        {/* Category Filter */}
        <div className="sticky top-14 z-20 bg-background-light/95 dark:bg-background-dark/95 backdrop-blur-sm w-full py-3 px-4 border-b border-gray-200 dark:border-white/5">
          <div className="max-w-[1600px] mx-auto">
            <div className="flex gap-3 overflow-x-auto hide-scrollbar">
              <button className="px-3 py-1.5 rounded-lg bg-gray-900 dark:bg-white text-white dark:text-black text-sm font-medium whitespace-nowrap transition-colors">
                All
              </button>
              <button className="px-3 py-1.5 rounded-lg bg-gray-200 dark:bg-surface-dark hover:bg-gray-300 dark:hover:bg-surface-darker text-gray-800 dark:text-gray-100 text-sm font-medium whitespace-nowrap transition-colors">
                Short Films
              </button>
              <button className="px-3 py-1.5 rounded-lg bg-gray-200 dark:bg-surface-dark hover:bg-gray-300 dark:hover:bg-surface-darker text-gray-800 dark:text-gray-100 text-sm font-medium whitespace-nowrap transition-colors">
                Documentaries
              </button>
              <button className="px-3 py-1.5 rounded-lg bg-gray-200 dark:bg-surface-dark hover:bg-gray-300 dark:hover:bg-surface-darker text-gray-800 dark:text-gray-100 text-sm font-medium whitespace-nowrap transition-colors">
                Animation
              </button>
              <button className="px-3 py-1.5 rounded-lg bg-gray-200 dark:bg-surface-dark hover:bg-gray-300 dark:hover:bg-surface-darker text-gray-800 dark:text-gray-100 text-sm font-medium whitespace-nowrap transition-colors">
                Sci-Fi
              </button>
              <button className="px-3 py-1.5 rounded-lg bg-gray-200 dark:bg-surface-dark hover:bg-gray-300 dark:hover:bg-surface-darker text-gray-800 dark:text-gray-100 text-sm font-medium whitespace-nowrap transition-colors">
                Experimental
              </button>
            </div>
          </div>
        </div>

        {/* Featured Section */}
        <section className="max-w-[2000px] mx-auto p-4 md:p-6 pb-2">
          <div className="relative w-full rounded-2xl overflow-hidden bg-surface-dark border border-white/5 shadow-2xl flex flex-col md:flex-row h-auto md:h-[300px] group cursor-pointer hover:border-white/20 transition-all">
            <div className="relative w-full md:w-[55%] h-48 md:h-full overflow-hidden bg-gray-900">
              <div className="absolute inset-0 bg-gradient-to-r from-primary/20 to-transparent" />
            </div>
            <div className="flex-1 p-6 md:p-8 flex flex-col justify-center bg-gradient-to-b from-surface-dark to-background-dark relative z-10">
              <div className="flex items-center gap-2 mb-3">
                <span className="w-2 h-2 rounded-full bg-primary animate-pulse" />
                <span className="text-primary text-xs font-bold uppercase tracking-wider">
                  Featured Premiere
                </span>
              </div>
              <h2 className="text-2xl md:text-3xl font-bold text-white mb-2 leading-tight">
                Discover Amazing Films
              </h2>
              <p className="text-gray-300 text-sm mb-6 line-clamp-2 md:line-clamp-3">
                Explore a curated collection of short films, feature films, and documentaries from emerging independent creators around the world.
              </p>
              <div className="flex gap-3 mt-auto">
                <button className="flex-1 bg-white text-black font-semibold py-2 px-4 rounded hover:bg-gray-200 transition-colors flex items-center justify-center gap-2">
                  <span className="material-icons text-lg">play_arrow</span>
                  Start Watching
                </button>
              </div>
            </div>
          </div>
        </section>

        {/* Films Grid */}
        <section className="max-w-[2000px] mx-auto p-4 md:p-6 pt-2">
          <h3 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
            <span className="material-icons text-primary">movie_filter</span>
            Recommended for You
          </h3>

          {films.length === 0 ? (
            <div className="text-center py-20">
              <p className="text-gray-400 mb-4">No films available yet.</p>
              <p className="text-sm text-gray-500">Be the first to upload a film!</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-x-4 gap-y-8">
              {films.map((film) => (
                <FilmCard key={film.id} film={film} />
              ))}
            </div>
          )}
        </section>

        {/* Footer */}
        <footer className="border-t border-gray-200 dark:border-white/5 mt-auto py-8">
          <div className="max-w-[1600px] mx-auto px-4 flex flex-col md:flex-row justify-between items-center gap-4 text-sm text-slate-500 dark:text-slate-400">
            <p>&copy; 2024 FilmTube. Empowering creators everywhere.</p>
            <div className="flex gap-6">
              <a className="hover:text-primary transition-colors" href="#">Privacy</a>
              <a className="hover:text-primary transition-colors" href="#">Terms</a>
              <a className="hover:text-primary transition-colors" href="#">Help</a>
            </div>
          </div>
        </footer>
      </main>
    </div>
  );
}
