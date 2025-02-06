import { EuiDescriptionList } from "@elastic/eui";
import React from "react";

export const ResourcesConfigTable = ({
  resourceRequest: {
    driver_cpu_request,
    driver_memory_request,
    executor_cpu_request,
    executor_memory_request,
    executor_replica,
  },
}) => {
  const items = [
    {
      title: "Driver CPU Request",
      description: driver_cpu_request,
    },
    {
      title: "Driver Memory Request",
      description: driver_memory_request,
    },
    {
      title: "Executor CPU Request",
      description: executor_cpu_request,
    },
    {
      title: "Executor Memory Request",
      description: executor_memory_request,
    },
    {
      title: "Executor Replicas",
      description: executor_replica,
    },
  ];

  return (
    <EuiDescriptionList
      compressed
      type="responsiveColumn"
      listItems={items}
      columnWidths={[2, 4]}
    />
  );
};
