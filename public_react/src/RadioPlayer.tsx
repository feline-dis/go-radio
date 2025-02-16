"use client"
import { formatTime } from "./lib/utils"
import { useRadioPlayer } from "./useRadioPlayer"

export default function RadioPlayer() {
  const {
    songInfo,
    isPlaying,
    elapsed,
    audioRef,
    volume,
    togglePausePlay,
    setIsPlaying,
    setVolume
  } = useRadioPlayer()

  if (!songInfo) {
    return (
      <div className="w-full max-w-xl h-32 bg-black/95 animate-pulse rounded-lg" />
    )
  }

  // Handle mute/unmute
  const handleMuteToggle = () => {
    const newVolume = volume === 0 ? 1 : 0;
    setVolume(newVolume);
  };

  // Handle volume change
  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newVolume = parseFloat(e.target.value);
    setVolume(newVolume);
  };

  return (
    <div className="w-full max-w-xl bg-black/95 text-white p-6 rounded-lg border-white border-1">
      <div className="flex items-start gap-4">
        {/* Album Art */}
        <img
          src={songInfo.art_url || "/fallback.jpg"}
          alt={`${songInfo.artist} - ${songInfo.title}`}
          className="w-20 h-20 rounded-md object-cover self-center"
        />

        {/* Main Content */}
        <div className="flex-1 min-w-0">
          {/* Track Info */}
          <div className="mb-4">
            <h2 className="text-sm font-medium truncate">{songInfo.title}</h2>
            <p className="text-xs text-gray-400 truncate">{songInfo.artist}</p>
          </div>

          {/* Progress Bar */}
          <div className="space-y-2">
            <div className="h-1 bg-gray-800 rounded-full overflow-hidden">
              <div
                className="h-full bg-white transition-all duration-300"
                style={{ width: `${(elapsed / songInfo.duration) * 100}%` }}
              />
            </div>

            {/* Time */}
            <div className="flex justify-between text-xs text-gray-400">
              <span>{formatTime(elapsed)}</span>
              <span>{formatTime(songInfo.duration)}</span>
            </div>
          </div>

          {/* Volume Control */}
          <div className="flex items-center gap-2 mt-4">
            <button
              className="p-1 hover:bg-white/10 rounded-sm transition-colors"
              onClick={handleMuteToggle}
              aria-label={volume === 0 ? "Unmute" : "Mute"}
            >
              {volume === 0 ? (
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M11 5 6 9H2v6h4l5 4zM22 9l-6 6M16 9l6 6" />
                </svg>
              ) : (
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
                  <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
                  <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />
                </svg>
              )}
            </button>
            <input
              type="range"
              min="0"
              max="1"
              step="0.01"
              value={volume}
              onChange={handleVolumeChange}
              className="w-20 h-1 bg-gray-800 rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-2 [&::-webkit-slider-thumb]:h-2 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-white"
              aria-label="Volume"
            />
          </div>
        </div>

        {/* Play/Pause Control */}
        <button
          onClick={togglePausePlay}
          className="ml-2 p-2 hover:bg-white/10 transition-colors rounded-sm"
          aria-label={isPlaying ? "Pause" : "Play"}
        >
          {isPlaying ? (
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <rect x="6" y="4" width="4" height="16" />
              <rect x="14" y="4" width="4" height="16" />
            </svg>
          ) : (
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polygon points="5 3 19 12 5 21 5 3" />
            </svg>
          )}
        </button>
      </div>
      <audio ref={audioRef} onEnded={() => setIsPlaying(false)} />
    </div>
  )
}
