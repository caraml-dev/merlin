const objectAssignDeep = require(`object-assign-deep`);

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

  static fromJson(json) {
    const logger = objectAssignDeep(new Logger(), json);
    return logger;
  }
}
