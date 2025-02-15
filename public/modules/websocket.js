import { eventBus } from "./eventBus.js";

export class WebSocketClient {
  constructor(url) {
    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      eventBus.emit("connected");
    };

    this.socket.onmessage = (message) => {
      try {
        const data = JSON.parse(message.data);

        switch (data.type) {
          case "current_song": {
            eventBus.emit("current_song", data.payload);
            break;
          }
          default: {
            console.error("Unknown message type:", data.type);
          }
        }
      } catch (error) {
        console.error("Failed to parse message:", error);
      }
    };

    this.socket.onclose = () => {
      console.log("Disconnected from server");
    };
  }

  send(data) {
    this.socket.send(JSON.stringify(data));
  }
}
