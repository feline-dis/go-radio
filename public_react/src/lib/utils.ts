export function formatTime(seconds: number): string {
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
}

export function getElapsedTime(start: Date): number {
  const now = new Date();
  return Math.floor((now.getTime() - start.getTime()) / 1000);
}
