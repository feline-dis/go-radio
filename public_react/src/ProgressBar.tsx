interface ProgressBarProps {
  elapsed: number
  duration: number
}

export default function ProgressBar({ elapsed, duration }: ProgressBarProps) {
  const progress = (elapsed / duration) * 100

  return (
    <div className="h-2 w-full rounded-full bg-gray-700">
      <div className="h-full rounded-full bg-blue-500" style={{ width: `${progress}%` }}></div>
    </div>
  )
}


