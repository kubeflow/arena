package task

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/strvals"
)

// quotedKeyPrefix is the placeholder prefix used for single-quoted key segments
// in --set expressions. Quoted segments are extracted before strvals parsing
// (which treats all dots as path separators) and restored afterwards.
const quotedKeyPrefix = "__QK_"

// ApplySetOverrides merges Helm-style --set expressions into raw YAML bytes.
// Each expression uses dot-notation paths: "worker.replicas=4", "envs.KEY=val".
// Single-quoted segments preserve dots literally, e.g.
// "worker.resources.'nvidia.com/gpu'=4" treats "nvidia.com/gpu" as one key.
// Returns the merged YAML bytes ready for LoadFromBytes.
func ApplySetOverrides(yamlData []byte, expressions []string) ([]byte, error) {
	if len(expressions) == 0 {
		return yamlData, nil
	}

	var base map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &base); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	if base == nil {
		base = make(map[string]interface{})
	}

	for _, expr := range expressions {
		// Validate that expression has a non-empty key
		if len(expr) == 0 || expr[0] == '=' {
			return nil, fmt.Errorf("failed to parse --set %q: empty key", expr)
		}

		processed, quotedKeys, err := preprocessQuotedKeys(expr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse --set %q: %w", expr, err)
		}

		if err := strvals.ParseInto(processed, base); err != nil {
			return nil, fmt.Errorf("failed to parse --set %q: %w", expr, err)
		}

		if len(quotedKeys) > 0 {
			restoreQuotedKeys(base, quotedKeys)
		}
	}

	merged, err := yaml.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged YAML: %w", err)
	}
	return merged, nil
}

// preprocessQuotedKeys scans a --set expression for single-quoted segments and
// replaces each with a dot-free placeholder so that strvals.ParseInto does not
// split them on dots. It returns the processed expression and a mapping from
// placeholder to the original quoted content.
//
// Example:
//
//	"worker.resources.'nvidia.com/gpu'=4"
//	→ "worker.resources.__QK_0__=4", {"__QK_0__": "nvidia.com/gpu"}
func preprocessQuotedKeys(expr string) (string, map[string]string, error) {
	// Only preprocess quoted segments in the key portion (before the first '=').
	// Values after '=' must not be touched — quotes there are literal user data.
	eqIdx := strings.IndexByte(expr, '=')
	keyPart := expr
	valueSuffix := ""
	if eqIdx >= 0 {
		keyPart = expr[:eqIdx]
		valueSuffix = expr[eqIdx:] // includes the '='
	}

	quotedKeys := make(map[string]string)
	var result strings.Builder
	counter := 0
	i := 0

	for i < len(keyPart) {
		if keyPart[i] == '\'' {
			// Find the closing quote (starting after the opening quote).
			end := strings.IndexByte(keyPart[i+1:], '\'')
			if end == -1 {
				return "", nil, fmt.Errorf("mismatched single quote in expression %q", expr)
			}
			// Content between the quotes.
			content := keyPart[i+1 : i+1+end]
			if content == "" {
				return "", nil, fmt.Errorf("empty quoted segment in expression %q", expr)
			}

			placeholder := fmt.Sprintf("%s%d%s", quotedKeyPrefix, counter, "__")
			counter++
			quotedKeys[placeholder] = content
			result.WriteString(placeholder)
			i = i + 1 + end + 1 // advance past the closing quote
		} else {
			result.WriteByte(keyPart[i])
			i++
		}
	}

	result.WriteString(valueSuffix)
	return result.String(), quotedKeys, nil
}

// restoreQuotedKeys walks a map produced by strvals.ParseInto and replaces any
// placeholder keys with the original quoted content from the mapping. It
// recurses into nested maps so that placeholders at any depth are restored.
//
// To avoid mutating the map during range iteration, it rebuilds the map into a
// fresh copy and then replaces the original map's contents.
func restoreQuotedKeys(m map[string]interface{}, quotedKeys map[string]string) {
	rebuilt := rebuildMap(m, quotedKeys)

	// Clear the original map and copy rebuilt entries back in place so that
	// callers who hold a reference to m see the updated contents.
	for k := range m {
		delete(m, k)
	}
	for k, v := range rebuilt {
		m[k] = v
	}
}

// rebuildMap creates a new map with all placeholder keys resolved.
// It recurses into nested maps and array-element maps.
func rebuildMap(m map[string]interface{}, quotedKeys map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for key, val := range m {
		realKey := resolveKey(key, quotedKeys)

		// Recurse into nested maps.
		if nested, ok := val.(map[string]interface{}); ok {
			restoreQuotedKeys(nested, quotedKeys)
			result[realKey] = nested
			continue
		}

		// Recurse into array elements that are maps.
		if arr, ok := val.([]interface{}); ok {
			for _, elem := range arr {
				if nested, ok := elem.(map[string]interface{}); ok {
					restoreQuotedKeys(nested, quotedKeys)
				}
			}
			result[realKey] = arr
			continue
		}

		result[realKey] = val
	}
	return result
}

// resolveKey replaces any placeholder substrings in key with the original
// quoted content. It first checks for an exact placeholder match, then falls
// back to substring replacement for keys where strvals embedded a placeholder
// inside a larger path segment (e.g., "foo__QK_0__qux" → "foobar.bazqux").
func resolveKey(key string, quotedKeys map[string]string) string {
	if orig, ok := quotedKeys[key]; ok {
		return orig
	}
	for ph, orig := range quotedKeys {
		if strings.Contains(key, ph) {
			return strings.ReplaceAll(key, ph, orig)
		}
	}
	return key
}

