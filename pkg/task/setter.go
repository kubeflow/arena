package task

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// quotedKeyPrefix is a unique placeholder for single-quoted --set key segments to prevent dot-splitting.
const quotedKeyPrefix = "__ARENA_QK_"

// maxIndex is the maximum array index allowed in --set expressions.
// This prevents accidental huge slice allocations from typos like
// items[999999999]=x, which would otherwise OOM the process.
const maxIndex = 65536

// pathSegment represents a single segment in a dotted key path.
// It can be either a map key (isArr=false) or an array index (isArr=true).
type pathSegment struct {
	key   string
	index int
	isArr bool
}

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

		if err := parseInto(processed, base); err != nil {
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

// parseInto parses a --set expression and applies it to the base map.
// The expression format is: key.path=value
// where key.path can contain dots (.) for nested keys and [N] for array indices.
func parseInto(expr string, base map[string]interface{}) error {
	// Find the first '='
	eqIdx := strings.IndexByte(expr, '=')
	if eqIdx < 0 {
		return errors.New("no '=' found in expression")
	}

	keyPart := expr[:eqIdx]
	valuePart := expr[eqIdx+1:]

	if keyPart == "" {
		return errors.New("empty key")
	}

	// Parse the key path into segments
	segments, err := parseKeyPath(keyPart)
	if err != nil {
		return err
	}

	if len(segments) == 0 {
		return errors.New("empty key path")
	}

	// Coerce the value to the appropriate type
	value := coerceValue(valuePart)

	// Set the value at the path
	return setValueAtPath(base, segments, value)
}

// parseKeyPath parses a dotted key path like "a.b[0].c" into segments.
func parseKeyPath(path string) ([]pathSegment, error) {
	segments := make([]pathSegment, 0)
	var current strings.Builder

	i := 0
	for i < len(path) {
		c := path[i]
		switch c {
		case '.':
			if current.Len() > 0 {
				segments = append(segments, pathSegment{key: current.String()})
				current.Reset()
			}
			i++
		case '[':
			// Save current key segment if any
			if current.Len() > 0 {
				segments = append(segments, pathSegment{key: current.String()})
				current.Reset()
			}
			// Find closing ']'
			j := i + 1
			for j < len(path) && path[j] != ']' {
				j++
			}
			if j >= len(path) {
				return nil, fmt.Errorf("unclosed '[' in path %q", path)
			}
			idxStr := path[i+1 : j]
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return nil, fmt.Errorf("invalid array index %q in path %q", idxStr, path)
			}
			if idx < 0 {
				return nil, fmt.Errorf("negative array index %d in path %q", idx, path)
			}
			if idx >= maxIndex {
				return nil, fmt.Errorf("array index %d exceeds maximum %d in path %q", idx, maxIndex, path)
			}
			segments = append(segments, pathSegment{index: idx, isArr: true})
			i = j + 1
			// Skip trailing '.' after ']'
			if i < len(path) && path[i] == '.' {
				i++
			}
		default:
			current.WriteByte(c)
			i++
		}
	}

	if current.Len() > 0 {
		segments = append(segments, pathSegment{key: current.String()})
	}

	return segments, nil
}

// setValueAtPath sets a value in a nested map/slice structure at the given path.
func setValueAtPath(base map[string]interface{}, segments []pathSegment, value interface{}) error {
	if len(segments) == 0 {
		return errors.New("empty path")
	}

	// Handle the simple case: single key, no arrays
	if len(segments) == 1 && !segments[0].isArr {
		base[segments[0].key] = value
		return nil
	}

	// Navigate through all segments except the last, building the container chain
	current := interface{}(base)

	for i := range segments[:len(segments)-1] {
		seg := segments[i]
		nextSeg := segments[i+1]

		if seg.isArr {
			// Current should be a slice
			arr, ok := current.([]interface{})
			if !ok {
				return fmt.Errorf("expected array at index %d, got %T", seg.index, current)
			}
			// Ensure index exists
			for len(arr) <= seg.index {
				arr = append(arr, nil)
			}
			// Create next level if nil
			if arr[seg.index] == nil {
				if nextSeg.isArr {
					arr[seg.index] = []interface{}{}
				} else {
					arr[seg.index] = make(map[string]interface{})
				}
			}
			// Write back the slice in case append grew it (append may return a
			// new slice header that is not reflected in the parent container).
			if err := updateSliceInParent(base, segments[:i], arr); err != nil {
				return err
			}
			current = arr[seg.index]
		} else {
			// Current should be a map
			m, ok := current.(map[string]interface{})
			if !ok {
				return fmt.Errorf("expected map at key %q, got %T", seg.key, current)
			}
			// Create next level if nil
			if m[seg.key] == nil {
				if nextSeg.isArr {
					m[seg.key] = []interface{}{}
				} else {
					m[seg.key] = make(map[string]interface{})
				}
			}
			current = m[seg.key]
		}
	}

	// Now set the final value
	lastSeg := segments[len(segments)-1]
	if lastSeg.isArr {
		arr, ok := current.([]interface{})
		if !ok {
			return fmt.Errorf("expected array at index %d, got %T", lastSeg.index, current)
		}
		for len(arr) <= lastSeg.index {
			arr = append(arr, nil)
		}
		arr[lastSeg.index] = value
		// Update the slice in its parent
		if err := updateSliceInParent(base, segments[:len(segments)-1], arr); err != nil {
			return err
		}
	} else {
		m, ok := current.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map at key %q, got %T", lastSeg.key, current)
		}
		m[lastSeg.key] = value
	}

	return nil
}

// updateSliceInParent updates a slice reference in its parent container.
// This is needed because append() may return a new slice header.
func updateSliceInParent(base map[string]interface{}, parentSegments []pathSegment, newSlice []interface{}) error {
	if len(parentSegments) == 0 {
		return nil
	}

	// Navigate to the parent of the slice
	current := interface{}(base)
	for i := range parentSegments[:len(parentSegments)-1] {
		seg := parentSegments[i]
		if seg.isArr {
			arr, ok := current.([]interface{})
			if !ok {
				return fmt.Errorf("expected array at path segment, got %T", current)
			}
			current = arr[seg.index]
		} else {
			m, ok := current.(map[string]interface{})
			if !ok {
				return fmt.Errorf("expected map at path segment, got %T", current)
			}
			current = m[seg.key]
		}
	}

	// Update the slice in the last parent
	lastParentSeg := parentSegments[len(parentSegments)-1]
	if lastParentSeg.isArr {
		arr, ok := current.([]interface{})
		if !ok {
			return fmt.Errorf("expected array at path segment, got %T", current)
		}
		arr[lastParentSeg.index] = newSlice
	} else {
		m, ok := current.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected map at path segment, got %T", current)
		}
		m[lastParentSeg.key] = newSlice
	}
	return nil
}

// coerceValue converts a string value to the appropriate Go type.
// - "true"/"True"/"TRUE" etc. → bool (case-insensitive, matches Helm strvals)
// - "false"/"False"/"FALSE" etc. → bool
// - "null" → nil (exact match only)
// - Valid integers → int
// - Otherwise → string
func coerceValue(s string) interface{} {
	if strings.EqualFold(s, "true") {
		return true
	}
	if strings.EqualFold(s, "false") {
		return false
	}
	if s == "null" {
		return nil
	}
	// Try to parse as integer
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return s
}

// preprocessQuotedKeys replaces single-quoted segments with placeholders before parsing, returning a restore mapping.
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
		if keyPart[i] != '\'' {
			result.WriteByte(keyPart[i])
			i++
			continue
		}
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

		placeholder := quotedKeyPrefix + strconv.Itoa(counter) + "__"
		counter++
		quotedKeys[placeholder] = content
		result.WriteString(placeholder)
		i = i + 1 + end + 1 // advance past the closing quote
	}

	result.WriteString(valueSuffix)
	return result.String(), quotedKeys, nil
}

// restoreQuotedKeys recursively replaces placeholder keys with their original quoted content.
func restoreQuotedKeys(m map[string]interface{}, quotedKeys map[string]string) {
	for key, val := range m {
		realKey := resolveKey(key, quotedKeys)
		if realKey != key {
			m[realKey] = val
			delete(m, key)
		}

		if nested, ok := val.(map[string]interface{}); ok {
			restoreQuotedKeys(nested, quotedKeys)
			continue
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

// resolveKey replaces any placeholder substrings in key with the original
// quoted content. It first checks for an exact placeholder match, then falls
// back to substring replacement for keys where a placeholder was embedded
// inside a larger path segment (e.g., "foo__ARENA_QK_0__qux" → "foobar.bazqux").
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
