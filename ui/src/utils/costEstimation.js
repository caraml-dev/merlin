import { costEstimationConfig } from "../config";
import { parseCpu, parseMemoryInGi } from "./kubernetesResourceParser";

/**
 * Calculate cost estimation for given number of replica, cpu, and memory resource
 * @param {int} replica number of replica
 * @param {string} cpu cpu resource
 * @param {string} memory memory resource
 * @param {string} gpu gpu resource
 * @param {float} monthlyCostPerGPU monthlyl cost per gpu
 * @returns estimated cost
 */
export const calculateCost = (
  replica,
  cpu,
  memory,
  gpu = 0,
  monthlyCostPerGPU = 0
) => {
  const parsed_cpu = parseCpu(cpu);
  const parsed_memory_gb = parseMemoryInGi(memory);
  const parsed_gpu = parseFloat(gpu) || 0;

  return (
    replica *
    (parsed_cpu * costEstimationConfig.cpuCost +
      parsed_memory_gb * costEstimationConfig.memoryCost +
      parsed_gpu * monthlyCostPerGPU)
  );
};
