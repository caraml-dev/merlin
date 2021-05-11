import { Logger } from "./Logger";

test("test new Logger()", () => {
  const logger = new Logger();

  expect(logger.model).toEqual({ enabled: false, mode: "" });
  expect(logger.transformer).toEqual({ enabled: false, mode: "" });
});
