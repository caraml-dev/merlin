import { EuiInMemoryTable } from "@elastic/eui";

export const SourceFeatures = ({ features, searchFeature }) => {
  const items = [];
  features.forEach((feature, index) => {
    console.log("feature", feature, searchFeature);
    if (searchFeature === "" || feature.includes(searchFeature)) {
      items.push({
        id: index,
        feature: feature,
      });
    }
  });

  const columns = [
    {
      field: "feature",
      name: "Feature Name",
      sortable: true,
    },
  ];

  return (
    <EuiInMemoryTable items={items} columns={columns} pagination sorting />
  );
};
