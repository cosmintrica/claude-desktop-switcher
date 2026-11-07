//go:build !windows

package account

import "errors"

// FetchUsage reads account limits only on Windows so far; elsewhere every
// profile reports why.
func FetchUsage(_ string, cookiesByName map[string]string) map[string]UsageResult {
	results := make(map[string]UsageResult, len(cookiesByName))
	for name := range cookiesByName {
		results[name] = UsageResult{Err: errors.New("account limits are read on Windows only")}
	}
	return results
}
