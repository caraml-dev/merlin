class LoggerConfig {
  constructor() {
    this.enabled = false;
    this.mode = "";
  }
}

export class Logger {
  constructor() {
    this.model = new LoggerConfig();
    this.transformer = new LoggerConfig();
  }
}
