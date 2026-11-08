//go:build darwin

package account

import (
	"errors"
	"sync"
)

// FetchUsage asks claude.ai for each profile's limits with the sign-in saved
// in its Cookies copy, opened with Claude's key from the login keychain. Keys
// are profile names; values are their Cookies paths.
func FetchUsage(_ string, cookiesByName map[string]string) map[string]UsageResult {
	results := make(map[string]UsageResult, len(cookiesByName))
	cm, err := NewCookieManager()
	if err != nil {
		for name := range cookiesByName {
			results[name] = UsageResult{Err: err}
		}
		return results
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for name, path := range cookiesByName {
		wg.Add(1)
		go func() {
			defer wg.Done()
			usage, err := cm.usage(path)
			mu.Lock()
			results[name] = UsageResult{Usage: usage, Err: err}
			mu.Unlock()
		}()
	}
	wg.Wait()
	return results
}

func (cm *CookieManager) usage(cookiesPath string) (Usage, error) {
	sessionKey, err := cm.Cookie(cookiesPath, "sessionKey")
	if err != nil {
		return Usage{}, err
	}
	if sessionKey == "" {
		return Usage{}, errors.New("no saved sign-in")
	}
	org, err := cm.Cookie(cookiesPath, "lastActiveOrg")
	if err != nil {
		org = ""
	}
	return fetchUsage(sessionKey, org)
}
