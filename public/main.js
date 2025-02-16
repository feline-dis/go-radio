import { WebSocketClient } from "./modules/websocket.js";
import { eventBus } from "./modules/eventBus.js";
import { audioController } from "./modules/audio.js";
import { differenceInSeconds } from "https://cdn.jsdelivr.net/npm/date-fns@4.1.0/+esm";

// elements
const startButton = document.getElementById("start");
const audio = document.getElementById("audio");
let container = null;
let elapsedTimer = null;
let elapsedTimerId = null;

let currSong = null;

eventBus.on("connected", () => {
  console.log("Connected to server");
});

function startElapsedTimer() {
  if (elapsedTimerId) {
    clearInterval(elapsedTimerId);
  }

  elapsedTimerId = setInterval(() => {
    if (currSong) {
      const elapsedSeconds = differenceInSeconds(
        new Date(),
        new Date(currSong.start_time),
      );

      document.getElementById("elapsed").innerText = elapsedSeconds;
    } else {
      document.getElementById("elapsed").innerText = "0:00";
    }
  }, 1000);
}

eventBus.on("current_song", async (data) => {
  console.log("Current song", data);
  currSong = data;
  await audioController.setAudioSource("/file/" + data.id);

  const secondsElapsed = differenceInSeconds(
    new Date(),
    new Date(data.start_time),
  );

  startElapsedTimer();

  audioController.play();
  audioController.seek(secondsElapsed);
});

startButton.addEventListener("click", async () => {
  console.log("Starting audio player");

  startButton.remove();

  container = document.createElement("div");
  container.id = "container";

  elapsedTimer = document.createElement("div");
  elapsedTimer.id = "elapsed";

  container.appendChild(elapsedTimer);
  document.body.appendChild(container);

  socket = new WebSocketClient("ws://localhost:8080/ws");
});
