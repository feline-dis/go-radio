export class AudioController {
  constructor(audioElementId) {
    if (!window.AudioContext && !window.webkitAudioContext) {
      throw new Error("Web Audio API is not supported in this browser.");
    }

    this.audioContext = new (window.AudioContext ||
      window.webkitAudioContext)();
    this.audioElement = document.getElementById(audioElementId);

    if (!this.audioElement) {
      throw new Error(`Audio element with ID "${audioElementId}" not found.`);
    }

    this.audioSource = null;
  }

  // Fetches and decodes audio data
  async loadAudio(url) {
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`Failed to fetch audio from ${url}`);
      }

      const arrayBuffer = await response.arrayBuffer();
      const audioBuffer = await this.audioContext.decodeAudioData(arrayBuffer);
      return audioBuffer;
    } catch (error) {
      console.error("Error loading audio:", error);
      throw error;
    }
  }

  // Set the audio source from the decoded audio data
  async setAudioSource(url) {
    const audioBuffer = await this.loadAudio(url);
    this.audioBuffer = audioBuffer;

    // Stop and disconnect any existing source
    if (this.audioSource) {
      this.audioSource.stop();
      this.audioSource.disconnect();
    }

    console.log("Setting audio source:", url);

    // Create a new source node
    this.audioSource = this.audioContext.createBufferSource();
    this.audioSource.buffer = audioBuffer;
    this.audioSource.connect(this.audioContext.destination);

    const srcObj = this.audioContext.createMediaStreamDestination();
    this.audioSource.connect(srcObj);

    this.audioElement.srcObject = srcObj.stream;
  }

  // Plays the audio
  async play() {
    if (!this.audioSource) {
      console.warn("No audio source is set.");
    }

    this.audioSource.start();
  }

  // Pauses the audio
  pause() {
    if (!this.audioSource) {
      console.warn("No audio source is set.");
    }

    this.audioSource.stop();
  }

  connectSource() {
    if (!this.audioBuffer) {
      console.warn("No audio buffer is set.");
      return;
    }

    this.audioSource = this.audioContext.createBufferSource();
  }

  seek(time) {
    if (!this.audioSource) {
      return;
    }

    this.audioSource.stop();
    this.audioSource.disconnect();
    this.audioSource = this.audioContext.createBufferSource();
    this.audioSource.buffer = this.audioBuffer;
    this.audioSource.connect(this.audioContext.destination);
    this.audioSource.start(0, time);
  }
}

export const audioController = new AudioController("audio");
