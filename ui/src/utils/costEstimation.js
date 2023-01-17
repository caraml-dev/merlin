import { parseCpu, parseMemoryInGi } from "./kubernetesResourceParser";

// TODO: move as env var
const CPU_COST_PER_MONTH = 28.46927;
const MEMORY_COST_PER_GI_PER_MONTH = 3.81498;

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
    (parsed_cpu * CPU_COST_PER_MONTH +
      parsed_memory_gb * MEMORY_COST_PER_GI_PER_MONTH)
  );
};
