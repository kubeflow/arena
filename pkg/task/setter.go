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
// The prefix is intentionally long and unique to minimize collision risk with
// real YAML keys that might contain a similar substring.
const quotedKeyPrefix = "__ARENA_QK_"

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
func restoreQuotedKeys(m map[string]interface{}, quotedKeys map[string]string) {
	// Collect replacements first to avoid mutating the map during iteration.
	replacements := make(map[string]string)
	for key := range m {
		// Exact match: the entire key is a placeholder.
		if originalKey, isPlaceholder := quotedKeys[key]; isPlaceholder {
			replacements[key] = originalKey
			continue
		}
		// Substring match: the placeholder is part of a larger key.
		// This happens when a quoted segment is adjacent to unquoted text,
		// e.g., "foo'bar.baz'qux" → key "foo__ARENA_QK_0__qux".
		for ph, orig := range quotedKeys {
			if strings.Contains(key, ph) {
				replacements[key] = strings.ReplaceAll(key, ph, orig)
				break
			}
		}
	}
	for oldKey, newKey := range replacements {
		m[newKey] = m[oldKey]
		delete(m, oldKey)
	}

	// Recurse into nested maps and arrays.
	for _, val := range m {
		if nested, ok := val.(map[string]interface{}); ok {
			restoreQuotedKeys(nested, quotedKeys)
		}
		if arr, ok := val.([]interface{}); ok {
			for _, elem := range arr {
				if nested, ok := elem.(map[string]interface{}); ok {
					restoreQuotedKeys(nested, quotedKeys)
				}
			}
		}
	}
}

