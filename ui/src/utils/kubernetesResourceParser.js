/**
 * parse kubernetes CPU resource
 * @param {string} input kubernetes cpu resource
 * @returns number of CPU
 */
export const parseCpu = (input) => {
  const hasMilli = input.match(/^([0-9]+)m$/);
  if (hasMilli) {
    return hasMilli[1] / 1000;
  }

  return parseFloat(input);
};

const memoryMultipliers = {
  k: 1000,
  M: 1000 ** 2,
  G: 1000 ** 3,
  T: 1000 ** 4,
  P: 1000 ** 5,
  E: 1000 ** 6,
  Ki: 1024,
  Mi: 1024 ** 2,
  Gi: 1024 ** 3,
  Ti: 1024 ** 4,
  Pi: 1024 ** 5,
  Ei: 1024 ** 6,
};

/**
 * Parse kubernetes memory request and convert it as Gi unit
 * @param {string} input
 * @returns memory resource in Gi
 */
export const parseMemoryInGi = (input) => {
  const unitMatch = input.match(/^([0-9]+)([A-Za-z]{1,2})$/);
  if (unitMatch) {
    return (
      (parseInt(unitMatch[1], 10) * memoryMultipliers[unitMatch[2]]) /
      memoryMultipliers["Gi"]
    );
  }

  return parseInt(input, 10) / memoryMultipliers["Gi"];
};
