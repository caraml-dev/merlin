package utils

// MergeMaps combine 2 maps string of string and return a new map
// keys in 'right' map is given priority if there is duplicate key in 'left' map
func MergeMaps(left map[string]string, right map[string]string) map[string]string {
	output := make(map[string]string)

	for k, v := range left {
		output[k] = v
	}

	for k, v := range right {
		output[k] = v
	}

	return output
}

// ExcludeKeys create a new map that exclude certain keys in the original map
func ExcludeKeys(srcMap map[string]string, keysToExclude []string) map[string]string {
	output := make(map[string]string)
	keysToExcludeMap := make(map[string]bool)

	for _, key := range keysToExclude {
		keysToExcludeMap[key] = true
	}

	for k, v := range srcMap {
		if keysToExcludeMap[k] == true {
			continue
		}
		output[k] = v
	}
	return output
}
