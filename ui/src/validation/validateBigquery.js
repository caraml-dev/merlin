/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Project ID: https://cloud.google.com/resource-manager/reference/rest/v1beta1/projects
// - It must be 6 to 30 lowercase letters, digits, or hyphens.
// - It must start with a letter.
// - Trailing hyphens are prohibited.
//
// Dataset: https://cloud.google.com/bigquery/docs/datasets#dataset-naming
// - May contain up to 1,024 characters
// - Can contain letters (upper or lower case), numbers, and underscores
// - Cannot contain spaces or special characters such as -, &, @, or %
//
// Table: https://cloud.google.com/bigquery/docs/tables
// - Contain up to 1,024 characters
// - Contain letters (upper or lower case), numbers, and underscores
//
// Column: https://cloud.google.com/bigquery/docs/schemas#column_names
// - A column name must contain only letters (a-z, A-Z), numbers (0-9), or underscores (_)
// - It must start with a letter or underscore.

// Format: project-name.dataset_name.table_name
export const validateBigqueryTable = table => {
  const expression = /^[a-z][a-z0-9-]+\.\w+([_]?\w)+\.\w+([_]?\w)+$/;
  return expression.test(String(table).toLowerCase());
};

export const validateBigqueryColumn = table => {
  const expression = /^[a-z_][a-z0-9_]*$/;
  return expression.test(String(table).toLowerCase());
};
