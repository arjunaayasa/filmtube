'use client';

import { useEffect, useRef, useState } from 'react';
import { useUIStore } from '@/lib/store';

interface VideoPlayerProps {
  src: string;
  poster?: string;
  title?: string;
}

export function VideoPlayer({ src, poster, title }: VideoPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [isMuted, setIsMuted] = useState(false);
  const [showControls, setShowControls] = useState(false);
  const [quality, setQuality] = useState('720p');
  const { cinemaMode, toggleCinemaMode } = useUIStore();
  const [hls, setHls] = useState<any>(null);

  // Initialize HLS.js
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const isSafari = /^((?!chrome|android).)*safari/i.test(navigator.userAgent);

    if (!isSafari && window.Hls) {
      const hlsInstance = new window.Hls({
        enableWorker: true,
        lowLatencyMode: true,
      });

      hlsInstance.loadSource(src);
      hlsInstance.attachMedia(videoRef.current);

      hlsInstance.on(window.Hls.Events.MANIFEST_PARSED, () => {
        console.log('HLS manifest loaded');
      });

      hlsInstance.on(window.Hls.Events.ERROR, (event: any, data: any) => {
        if (data.fatal) {
          console.error('HLS error:', data);
          hlsInstance.destroy();
        }
      });

      setHls(hlsInstance);

      return () => {
        hlsInstance.destroy();
      };
    } else if (videoRef.current) {
      // Safari native HLS support
      videoRef.current.src = src;
    }
  }, [src]);

  const togglePlay = () => {
    if (!videoRef.current) return;

    if (isPlaying) {
      videoRef.current.pause();
    } else {
      videoRef.current.play();
    }
    setIsPlaying(!isPlaying);
  };

  const handleTimeUpdate = () => {
    if (videoRef.current) {
      setCurrentTime(videoRef.current.currentTime);
    }
  };

  const handleLoadedMetadata = () => {
    if (videoRef.current) {
      setDuration(videoRef.current.duration);
    }
  };

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const time = parseFloat(e.target.value);
    if (videoRef.current) {
      videoRef.current.currentTime = time;
      setCurrentTime(time);
    }
  };

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const vol = parseFloat(e.target.value);
    if (videoRef.current) {
      videoRef.current.volume = vol;
      setVolume(vol);
      setIsMuted(vol === 0);
    }
  };

  const toggleMute = () => {
    if (videoRef.current) {
      videoRef.current.muted = !isMuted;
      setIsMuted(!isMuted);
    }
  };

  const toggleFullscreen = () => {
    if (!containerRef.current) return;

    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      containerRef.current.requestFullscreen();
    }
  };

  const formatTime = (time: number): string => {
    const minutes = Math.floor(time / 60);
    const seconds = Math.floor(time % 60);
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  return (
    <div
      ref={containerRef}
      className={`player-section relative w-full aspect-video bg-black rounded-xl overflow-hidden shadow-2xl shadow-primary/10 group ring-1 ring-white/10 ${cinemaMode ? 'cinema-mode-active' : ''}`}
      onMouseEnter={() => setShowControls(true)}
      onMouseLeave={() => setShowControls(false)}
    >
      <video
        ref={videoRef}
        className="w-full h-full object-contain"
        poster={poster}
        onTimeUpdate={handleTimeUpdate}
        onLoadedMetadata={handleLoadedMetadata}
        onPlay={() => setIsPlaying(true)}
        onPause={() => setIsPlaying(false)}
      />

      {/* Play Button Overlay (when paused) */}
      {!isPlaying && (
        <div className="absolute inset-0 flex items-center justify-center z-10 pointer-events-none">
          <button
            onClick={togglePlay}
            className="w-20 h-20 bg-primary/90 rounded-full flex items-center justify-center shadow-lg shadow-primary/30 backdrop-blur-sm transform hover:scale-110 transition-transform duration-300 pointer-events-auto"
          >
            <span className="material-icons text-white text-5xl ml-1">play_arrow</span>
          </button>
        </div>
      )}

      {/* Controls Overlay */}
      <div
        className={`absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black via-black/80 to-transparent px-6 pb-6 pt-16 transition-opacity duration-300 ${
          showControls || !isPlaying ? 'opacity-100' : 'opacity-0'
        }`}
      >
        {/* Progress Bar */}
        <div className="w-full h-1 bg-white/20 rounded-full cursor-pointer group/progress relative mb-3">
          <div
            className="absolute top-0 left-0 h-full bg-primary rounded-full"
            style={{ width: `${(currentTime / duration) * 100}%` }}
          >
            <div className="absolute right-0 top-1/2 -translate-y-1/2 w-3 h-3 bg-white rounded-full scale-0 group-hover/progress:scale-100 transition-transform shadow" />
          </div>
          <input
            type="range"
            min={0}
            max={duration}
            value={currentTime}
            onChange={handleSeek}
            className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
          />
        </div>

        {/* Control Buttons */}
        <div className="flex items-center justify-between text-white">
          <div className="flex items-center gap-4">
            <button onClick={togglePlay} className="hover:text-primary transition-colors">
              <span className="material-icons">
                {isPlaying ? 'pause' : 'play_arrow'}
              </span>
            </button>

            <button onClick={toggleMute} className="hover:text-primary transition-colors">
              <span className="material-icons-outlined">
                {isMuted || volume === 0 ? 'volume_off' : 'volume_up'}
              </span>
            </button>

            <input
              type="range"
              min={0}
              max={1}
              step={0.1}
              value={isMuted ? 0 : volume}
              onChange={handleVolumeChange}
              className="w-20 h-1 accent-primary"
            />

            <span className="text-xs font-medium text-gray-300">
              {formatTime(currentTime)} / {formatTime(duration)}
            </span>
          </div>

          <div className="flex items-center gap-4">
            {/* Cinema Mode Toggle */}
            <button
              onClick={toggleCinemaMode}
              className="flex items-center gap-1.5 px-3 py-1 rounded hover:bg-white/10 transition-colors text-xs font-medium uppercase tracking-wider"
            >
              <span className="material-icons-outlined text-sm">wb_incandescent</span>
              Cinema Mode
            </button>

            {/* Quality Selector */}
            <select
              value={quality}
              onChange={(e) => setQuality(e.target.value)}
              className="bg-transparent border border-white/20 rounded px-2 py-1 text-xs font-medium uppercase tracking-wider cursor-pointer hover:bg-white/10 transition-colors"
            >
              <option value="360p">360p</option>
              <option value="720p">720p</option>
            </select>

            <button onClick={toggleFullscreen} className="hover:text-primary transition-colors">
              <span className="material-icons-outlined">fullscreen</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
