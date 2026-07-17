package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chmuri/terraform-provider-freeipa/client"
)

func isEmptyModlistError(err error) bool {
	if err == nil {
		return false
	}
	if rpcErr, ok := err.(*client.RPCError); ok {
		return rpcErr.Code == 4202 // EmptyModlist
	}
	return strings.Contains(strings.ToLower(err.Error()), "empty modest") || strings.Contains(strings.ToLower(err.Error()), "no modifications")
}

// parseStringVal extracts a single string from various possible JSON formats
func parseStringVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		if len(val) > 0 {
			return parseStringVal(val[0])
		}
	case []string:
		if len(val) > 0 {
			return val[0]
		}
	case map[string]interface{}:
		if dnsName, ok := val["__dns_name__"].(string); ok {
			return dnsName
		}
	}
	return fmt.Sprintf("%v", v)
}

// parseIntVal extracts an int64 from various possible JSON formats
func parseIntVal(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	case []interface{}:
		if len(val) > 0 {
			return parseIntVal(val[0])
		}
	case []string:
		if len(val) > 0 {
			return parseIntVal(val[0])
		}
	}
	return 0
}

// parseBoolVal extracts a boolean from various possible JSON formats
func parseBoolVal(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return strings.ToUpper(val) == "TRUE" || val == "1"
	case []interface{}:
		if len(val) > 0 {
			return parseBoolVal(val[0])
		}
	case []bool:
		if len(val) > 0 {
			return val[0]
		}
	case []string:
		if len(val) > 0 {
			return parseBoolVal(val[0])
		}
	}
	return false
}

// parseStringSlice extracts a slice of strings from various possible JSON formats
func parseStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		return []string{val}
	case []interface{}:
		res := make([]string, 0, len(val))
		for _, item := range val {
			res = append(res, parseStringVal(item))
		}
		return res
	case []string:
		return val
	}
	return []string{fmt.Sprintf("%v", v)}
}

// contains returns true if slice contains val
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// difference returns elements in slice1 that are not in slice2
func difference(slice1 []string, slice2 []string) []string {
	var diff []string
	m := make(map[string]bool)

	for _, val := range slice2 {
		m[val] = true
	}

	for _, val := range slice1 {
		if !m[val] {
			diff = append(diff, val)
		}
	}
	return diff
}
