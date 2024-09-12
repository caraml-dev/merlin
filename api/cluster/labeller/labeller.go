package labeller

import (
	"fmt"
	"regexp"
)

const (
	LabelAppName          = "app"
	LabelComponent        = "component"
	LabelEnvironment      = "environment"
	LabelOrchestratorName = "orchestrator"
	LabelStreamName       = "stream"
	LabelTeamName         = "team"
	LabelManaged          = "managed"
)

var reservedKeys = map[string]bool{
	LabelAppName:          true,
	LabelComponent:        true,
	LabelEnvironment:      true,
	LabelOrchestratorName: true,
	LabelStreamName:       true,
	LabelTeamName:         true,
	LabelManaged:          true,
}

var (
	prefix           string
	nsPrefix         string
	environment      string
	validPrefixRegex = regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?(/)$")
)

// InitKubernetesLabeller builds a new KubernetesLabeller Singleton
func InitKubernetesLabeller(p, ns, e string) error {
	if err := isValidPrefix(p, "prefix"); err != nil {
		return err
	}
	if err := isValidPrefix(ns, "namespace prefix"); err != nil {
		return err
	}

	prefix = p
	nsPrefix = ns
	environment = e
	return nil
}

// isValidPrefix checks if the given prefix is valid
func isValidPrefix(prefix, name string) error {
	if len(prefix) > 253 {
		return fmt.Errorf("length of %s is greater than 253 characters", name)
	}
	if isValidPrefix := validPrefixRegex.MatchString(prefix); !isValidPrefix {
		return fmt.Errorf("%s name violates kubernetes label's prefix constraint", name)
	}
	return nil
}

// GetLabelName prefixes the label with the config specified label and returns the formatted label prefix
func GetLabelName(name string) string {
	return fmt.Sprintf("%s%s", prefix, name)
}

// GetNamespaceLabel prefixes the label with the config specified namespace label and returns the formatted label prefix
func GetNamespaceLabel(name string) string {
	return fmt.Sprintf("%s%s", nsPrefix, name)
}

// GetPrefix returns the labeller prefix
func GetPrefix() string {
	return prefix
}

// GetEnvironment returns the environment
func GetEnvironment() string {
	return environment
}

// IsReservedKey checks if the key is a reserved key
func IsReservedKey(key string) bool {
	_, usingReservedKeys := reservedKeys[key]
	return usingReservedKeys
}
