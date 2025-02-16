import { useState, useEffect, useRef } from "react";
import { getElapsedTime } from "./lib/utils";

interface SongInfo {
  artist: string;
  title: string;
  art_url: string;
  duration: number;
  start_time: string;
  end_time: string;
  id: string;
}

interface Message {
  type: string;
  payload: SongInfo;
}

const audioContext = new AudioContext();
const gainNode = audioContext.createGain();
gainNode.connect(audioContext.destination);

export const useRadioPlayer = () => {
  const [songInfo, setSongInfo] = useState<SongInfo | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const audioSourceRef = useRef<AudioBufferSourceNode | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [volume, setVolume] = useState(() => {
    if (typeof window !== "undefined") {
      return parseFloat(localStorage.getItem("radio-volume") || "1");
    }
    return 1;
  });

  const getAudioData = async (id: string) => {
    const audioData = await fetch(`/file/${id}`);
    const buffer = await audioData.arrayBuffer();
    const audioBuffer = await audioContext.decodeAudioData(buffer);
    const source = audioContext.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(gainNode); // Connect to gain node instead of destination
    return source;
  };

  useEffect(() => {
    wsRef.current = new WebSocket("/ws");
    wsRef.current.onopen = () => {
      console.log("WebSocket connected");
    };
    wsRef.current.onerror = (error) => {
      console.error("WebSocket error:", error);
    };
    wsRef.current.onmessage = async (event) => {
      const data = JSON.parse(event.data) as Message;
      const source = await getAudioData(data.payload.id);
      const elapsed = getElapsedTime(new Date(data.payload.start_time));

      if (!source) return;

      if (audioSourceRef.current) {
        audioSourceRef.current.stop();
        audioSourceRef.current.disconnect();
      }

      audioSourceRef.current = source;
      audioSourceRef.current.start(0, elapsed);

      const srcObj = audioContext.createMediaStreamDestination();
      audioSourceRef.current.connect(srcObj);

      if (audioRef.current) {
        audioRef.current.srcObject = srcObj.stream;
      }

      setIsPlaying(true);
      setSongInfo(data.payload);
      setElapsed(0);
    };

    return () => {
      console.log("WebSocket disconnected");
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  useEffect(() => {
    if (songInfo) {
      const interval = setInterval(() => {
        const now = new Date();
        const start = new Date(songInfo.start_time);
        setElapsed(Math.floor((now.getTime() - start.getTime()) / 1000));
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [songInfo]);

  const togglePausePlay = () => {
    if (!audioSourceRef.current) return;

    if (isPlaying) {
      audioSourceRef.current.stop();
      audioSourceRef.current.disconnect();
      setElapsed(getElapsedTime(new Date(songInfo!.start_time)));
    } else {
      const bfrSrc = audioContext.createBufferSource();
      bfrSrc.buffer = audioSourceRef.current.buffer!;
      bfrSrc.connect(gainNode); // Connect to gain node instead of destination
      audioSourceRef.current = bfrSrc;
      audioSourceRef.current.start(0, elapsed);
    }

    setIsPlaying(!isPlaying);
  };

  useEffect(() => {
    gainNode.gain.value = volume; // Update gain node value

    if (typeof window !== "undefined") {
      localStorage.setItem("radio-volume", volume.toString());
    }
  }, [volume]);

  return {
    songInfo,
    isPlaying,
    elapsed,
    audioRef,
    setIsPlaying,
    togglePausePlay,
    volume,
    setVolume,
  };
};
