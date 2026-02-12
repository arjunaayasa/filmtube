'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import { useAuthStore } from '@/lib/store';

export default function UploadPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [step, setStep] = useState<'details' | 'upload' | 'processing' | 'done'>('details');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [type, setType] = useState<'SHORT_FILM' | 'FEATURE_FILM'>('SHORT_FILM');
  const [filmId, setFilmId] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState('');

  const canUpload = user?.role === 'CREATOR' || user?.role === 'ADMIN';

  const handleCreateFilm = async () => {
    if (!title.trim()) {
      setError('Please enter a title');
      return;
    }

    try {
      const film = await api.createFilm({
        title: title.trim(),
        description: description.trim(),
        type,
      });
      setFilmId(film.id);
      setStep('upload');
    } catch (err: any) {
      setError(err.message || 'Failed to create film');
    }
  };

  const handleUpload = async () => {
    if (!filmId) return;

    try {
      // Get upload URL
      const uploadInfo = await api.getUploadURL(filmId);
      setStep('processing');

      // Get file input
      const fileInput = document.getElementById('video-file') as HTMLInputElement;
      const file = fileInput?.files?.[0];
      if (!file) {
        setError('Please select a video file');
        setStep('upload');
        return;
      }

      // Upload directly to R2
      await fetch(uploadInfo.upload_url, {
        method: 'PUT',
        body: file,
        headers: {
          'Content-Type': 'video/mp4',
        },
      });

      setUploadProgress(100);

      // Confirm upload to trigger transcoding
      await api.confirmUpload(filmId);

      setStep('done');

      // Redirect after 3 seconds
      setTimeout(() => {
        router.push(`/films/${filmId}`);
      }, 3000);
    } catch (err: any) {
      setError(err.message || 'Upload failed');
      setStep('upload');
    }
  };

  if (!canUpload) {
    return (
      <div className="min-h-screen bg-background-light dark:bg-background-dark flex items-center justify-center p-4">
        <div className="text-center">
          <span className="material-icons text-6xl text-gray-600 mb-4">lock</span>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
            Creator Access Required
          </h1>
          <p className="text-slate-500 dark:text-slate-400">
            You need a Creator account to upload films.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark">
      <div className="max-w-3xl mx-auto px-4 py-12">
        {/* Progress Steps */}
        <div className="flex items-center justify-center mb-12">
          <div className="flex items-center gap-4">
            <div className={`flex items-center gap-2 ${step === 'details' ? 'text-primary' : step !== 'details' ? 'text-green-500' : 'text-gray-500'}`}>
              <span className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${step === 'details' ? 'bg-primary text-white' : step !== 'details' ? 'bg-green-500 text-white' : 'bg-gray-300 dark:bg-gray-700'}`}>
                {step !== 'details' ? '✓' : '1'}
              </span>
              <span className="text-sm font-medium">Details</span>
            </div>
            <div className={`w-16 h-0.5 ${step !== 'details' ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-700'}`} />
            <div className={`flex items-center gap-2 ${step === 'upload' || step === 'processing' || step === 'done' ? 'text-primary' : 'text-gray-500'}`}>
              <span className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${step === 'upload' || step === 'processing' ? 'bg-primary text-white' : step === 'done' ? 'bg-green-500 text-white' : 'bg-gray-300 dark:bg-gray-700'}`}>
                {step === 'done' ? '✓' : '2'}
              </span>
              <span className="text-sm font-medium">Upload</span>
            </div>
            <div className={`w-16 h-0.5 ${step === 'done' ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-700'}`} />
            <div className={`flex items-center gap-2 ${step === 'done' ? 'text-green-500' : 'text-gray-500'}`}>
              <span className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${step === 'done' ? 'bg-green-500 text-white' : 'bg-gray-300 dark:bg-gray-700'}`}>
                {step === 'done' ? '✓' : '3'}
              </span>
              <span className="text-sm font-medium">Done</span>
            </div>
          </div>
        </div>

        {/* Step Content */}
        <div className="bg-white dark:bg-surface-dark rounded-xl p-8 border border-gray-200 dark:border-white/5 shadow-sm">
          {/* Step 1: Film Details */}
          {step === 'details' && (
            <>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
                Film Details
              </h1>
              <p className="text-slate-500 dark:text-slate-400 mb-6">
                Tell us about your film before uploading
              </p>

              {error && (
                <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-500 text-sm">
                  {error}
                </div>
              )}

              <div className="space-y-4">
                <div>
                  <label htmlFor="title" className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                    Title *
                  </label>
                  <input
                    id="title"
                    type="text"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    className="w-full bg-gray-100 dark:bg-white/5 border border-gray-300 dark:border-white/10 rounded-lg py-2.5 px-4 text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary dark:text-white transition-colors"
                    placeholder="Enter your film title"
                  />
                </div>

                <div>
                  <label htmlFor="description" className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                    Description
                  </label>
                  <textarea
                    id="description"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    rows={4}
                    className="w-full bg-gray-100 dark:bg-white/5 border border-gray-300 dark:border-white/10 rounded-lg py-2.5 px-4 text-sm focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary dark:text-white transition-colors resize-none"
                    placeholder="Describe your film, genre, story, etc."
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
                    Type *
                  </label>
                  <div className="flex gap-4">
                    <label className="flex-1 flex items-center gap-2 p-3 bg-gray-100 dark:bg-white/5 rounded-lg cursor-pointer border-2 border-transparent has-[:checked]:border-primary">
                      <input
                        type="radio"
                        name="type"
                        value="SHORT_FILM"
                        checked={type === 'SHORT_FILM'}
                        onChange={() => setType('SHORT_FILM')}
                        className="accent-primary"
                      />
                      <span className="text-sm text-slate-700 dark:text-slate-300">Short Film</span>
                    </label>
                    <label className="flex-1 flex items-center gap-2 p-3 bg-gray-100 dark:bg-white/5 rounded-lg cursor-pointer border-2 border-transparent has-[:checked]:border-primary">
                      <input
                        type="radio"
                        name="type"
                        value="FEATURE_FILM"
                        checked={type === 'FEATURE_FILM'}
                        onChange={() => setType('FEATURE_FILM')}
                        className="accent-primary"
                      />
                      <span className="text-sm text-slate-700 dark:text-slate-300">Feature Film</span>
                    </label>
                  </div>
                </div>

                <button
                  onClick={handleCreateFilm}
                  className="w-full bg-primary hover:bg-red-700 text-white font-semibold py-2.5 px-4 rounded-lg transition-colors flex items-center justify-center gap-2"
                >
                  <span>Continue to Upload</span>
                  <span className="material-icons text-sm">arrow_forward</span>
                </button>
              </div>
            </>
          )}

          {/* Step 2: Upload */}
          {step === 'upload' && (
            <>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
                Upload Video
              </h1>
              <p className="text-slate-500 dark:text-slate-400 mb-6">
                Upload your video file (MP4, max 2GB)
              </p>

              {error && (
                <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-lg text-red-500 text-sm">
                  {error}
                </div>
              )}

              <div className="border-2 border-dashed border-gray-300 dark:border-white/10 rounded-xl p-12 text-center">
                <input
                  id="video-file"
                  type="file"
                  accept="video/mp4"
                  className="hidden"
                  onChange={() => {/* File selected */}}
                />
                <label
                  htmlFor="video-file"
                  className="cursor-pointer flex flex-col items-center gap-4"
                >
                  <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center">
                    <span className="material-icons text-3xl text-primary">upload</span>
                  </div>
                  <div>
                    <p className="text-slate-900 dark:text-white font-medium mb-1">
                      Click to upload or drag and drop
                    </p>
                    <p className="text-sm text-slate-500 dark:text-slate-400">
                      MP4 files up to 2GB
                    </p>
                  </div>
                </label>
              </div>

              <div className="flex gap-4 mt-6">
                <button
                  onClick={() => setStep('details')}
                  className="flex-1 bg-gray-200 dark:bg-white/10 hover:bg-gray-300 dark:hover:bg-white/20 text-slate-700 dark:text-slate-300 font-medium py-2.5 px-4 rounded-lg transition-colors"
                >
                  Back
                </button>
                <button
                  onClick={handleUpload}
                  className="flex-1 bg-primary hover:bg-red-700 text-white font-semibold py-2.5 px-4 rounded-lg transition-colors"
                >
                  Start Upload
                </button>
              </div>
            </>
          )}

          {/* Step 3: Processing */}
          {step === 'processing' && (
            <>
              <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
                Uploading...
              </h1>
              <p className="text-slate-500 dark:text-slate-400 mb-8">
                Please wait while we upload your video
              </p>

              <div className="w-full bg-gray-200 dark:bg-white/10 rounded-full h-3 mb-4">
                <div
                  className="bg-primary h-3 rounded-full transition-all duration-300"
                  style={{ width: `${uploadProgress}%` }}
                />
              </div>
              <p className="text-center text-sm text-slate-500 dark:text-slate-400">
                {uploadProgress}% complete
              </p>
            </>
          )}

          {/* Step 4: Done */}
          {step === 'done' && (
            <>
              <div className="text-center">
                <div className="w-20 h-20 rounded-full bg-green-500/10 flex items-center justify-center mx-auto mb-6">
                  <span className="material-icons text-5xl text-green-500">check_circle</span>
                </div>
                <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-2">
                  Upload Successful!
                </h1>
                <p className="text-slate-500 dark:text-slate-400 mb-6">
                  Your film is being processed. You can view it once transcoding is complete.
                </p>
                <p className="text-sm text-slate-400">
                  Redirecting to your film...
                </p>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
