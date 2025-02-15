export class EventBus {
  constructor() {
    this.events = {}; // Store event listeners
  }

  // Add an event listener for a specific event
  on(event, listener) {
    if (!this.events[event]) {
      this.events[event] = [];
    }
    this.events[event].push(listener);
  }

  // Remove a specific listener for an event
  off(event, listener) {
    if (!this.events[event]) return;

    this.events[event] = this.events[event].filter((l) => l !== listener);

    // Cleanup if no listeners remain
    if (this.events[event].length === 0) {
      delete this.events[event];
    }
  }

  // Emit an event with optional data
  emit(event, data) {
    if (!this.events[event]) return;

    this.events[event].forEach((listener) => {
      listener(data);
    });
  }

  // Add an event listener that only triggers once
  once(event, listener) {
    const onceListener = (data) => {
      listener(data);
      this.off(event, onceListener); // Remove listener after execution
    };
    this.on(event, onceListener);
  }
}

export const eventBus = new EventBus();
