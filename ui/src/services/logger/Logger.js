const objectAssignDeep = require(`object-assign-deep`);

class LoggerConfig {
  constructor() {
    this.enabled = false;
    this.mode = "";
  }
}

class PredictionLoggerConfig {
  constructor() {
    this.enabled = false;
    this.raw_features_table = "";
    this.entities_table = "";
  }
}

export class Logger {
  constructor() {
    this.model = new LoggerConfig();
    this.transformer = new LoggerConfig();
    this.prediction = new PredictionLoggerConfig();
  }

  static fromJson(json) {
    const logger = objectAssignDeep(new Logger(), json);
    return logger;
  }
}
