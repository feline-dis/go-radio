import { WebSocketClient } from "./modules/websocket.js";
import { eventBus } from "./modules/eventBus.js";
import { audioController } from "./modules/audio.js";

let socket;
const startButton = document.getElementById("start");

eventBus.on("connected", () => {
  console.log("Connected to server");
});

eventBus.on("current_song", async (data) => {
  await audioController.setAudioSource("/file/" + data.id);
  console.log(data);
  const startTime = new Date(data.start_time);
  console.log("Start time:", startTime);
  const currentTime = Date.now();

  const secondsElapsed = (currentTime - startTime) / 1000;

  console.log("Seconds elapsed:", secondsElapsed);

  audioController.play();
  audioController.seek(secondsElapsed);
});

startButton.addEventListener("click", async () => {
  console.log("Starting audio player");
  socket = new WebSocketClient("ws://localhost:8080/ws");
});
