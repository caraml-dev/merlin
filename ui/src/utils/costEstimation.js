import { parseCpu, parseMemoryInGi } from "./kubernetesResourceParser";
import { costEstimationConfig } from "../config";

/**
 * Calculate cost estimation for given number of replica, cpu, and memory resource
 * @param {int} replica number of replica
 * @param {string} cpu cpu resource
 * @param {string} memory memory resource
 * @returns estimated cost
 */
export const calculateCost = (replica, cpu, memory) => {
  const parsed_cpu = parseCpu(cpu);
  const parsed_memory_gb = parseMemoryInGi(memory);

  return (
    replica *
    (parsed_cpu * costEstimationConfig.cpuCost +
      parsed_memory_gb * costEstimationConfig.memoryCost)
  );
};
