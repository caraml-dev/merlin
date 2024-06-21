import { EuiHealth, EuiSearchBar } from "@elastic/eui";
import { jobStatuses } from "./JobStatus";

const JobSearchBar = ({ placeholder, onChange }) => {
  const filters = [
    {
      type: "field_value_selection",
      field: "status",
      name: "Status",
      multiSelect: false,
      operator: "exact",
      options: jobStatuses.map((status) => ({
        value: status.name,
        view: <EuiHealth color={status.color}>{status.name}</EuiHealth>,
      })),
    },
  ];

  return (
    <EuiSearchBar
      box={{
        placeholder: placeholder,
      }}
      filters={filters}
      onChange={onChange}
    />
  );
};

export default JobSearchBar;
