package account

import (
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"time"
)

const (
	usageEndpoint    = "https://claude.ai/api/organizations/%s/usage"
	browserCheckWait = 5 * time.Second
)

// fetchEach asks for each account's limits one at a time, because claude.ai's
// Cloudflare guard answers some requests that arrive together with a browser
// check. An account that got the check is asked once more after wait.
func fetchEach(paths map[string]string, fetch func(path string) (Usage, error), wait func()) map[string]UsageResult {
	results := make(map[string]UsageResult, len(paths))
	var checked []string
	for _, name := range slices.Sorted(maps.Keys(paths)) {
		usage, err := fetch(paths[name])
		results[name] = UsageResult{Usage: usage, Err: err}
		if errors.Is(err, errBrowserCheck) {
			checked = append(checked, name)
		}
	}
	if len(checked) > 0 {
		wait()
		for _, name := range checked {
			usage, err := fetch(paths[name])
			results[name] = UsageResult{Usage: usage, Err: err}
		}
	}
	return results
}

// Limit is how much of one usage window an account has used, in percent, and
// when that window starts over.
type Limit struct {
	Percent float64
	Resets  time.Time
	Known   bool
}

// At is the limit as it stands at now: once its reset time has passed, the
// window has started over.
func (l Limit) At(now time.Time) Limit {
	if !l.Resets.IsZero() && !now.Before(l.Resets) {
		return Limit{Known: l.Known}
	}
	return l
}

// Usage holds an account's 5-hour and weekly limits as Claude's Usage page
// shows them. Known is false when claude.ai sent neither.
type Usage struct {
	FiveHour, Week Limit
	Known          bool
	Checked        time.Time
}

// UsageResult is one account's answer: its usage, or why it could not be read.
type UsageResult struct {
	Usage
	Err error
}

type usageWindow struct {
	Utilization *float64   `json:"utilization"`
	ResetsAt    *time.Time `json:"resets_at"`
}

type usageResponse struct {
	FiveHour *usageWindow `json:"five_hour"`
	SevenDay *usageWindow `json:"seven_day"`
}

func (w *usageWindow) limit() (Limit, bool) {
	if w == nil || w.Utilization == nil {
		return Limit{}, false
	}
	limit := Limit{Percent: *w.Utilization, Known: true}
	if w.ResetsAt != nil {
		limit.Resets = *w.ResetsAt
	}
	return limit, true
}

func usageFrom(response usageResponse, checked time.Time) Usage {
	fiveHour, knownFiveHour := response.FiveHour.limit()
	week, knownWeek := response.SevenDay.limit()
	return Usage{FiveHour: fiveHour, Week: week, Known: knownFiveHour || knownWeek, Checked: checked}
}

// fetchUsage reads the limits of the organization the account last used in
// Claude, or of its first subscription when that is not known.
func fetchUsage(sessionKey, org string) (Usage, error) {
	if org == "" {
		var err error
		if org, err = subscriptionOrg(sessionKey); err != nil {
			return Usage{}, err
		}
	}
	response, err := getJSON[usageResponse](fmt.Sprintf(usageEndpoint, url.PathEscape(org)), sessionKey)
	if err != nil {
		return Usage{}, err
	}
	return usageFrom(response, time.Now()), nil
}

func subscriptionOrg(sessionKey string) (string, error) {
	orgs, err := getJSON[[]orgResponse](orgsEndpoint, sessionKey)
	if err != nil {
		return "", err
	}
	for _, org := range orgs {
		if org.UUID != "" && parsePlan(org) != "" {
			return org.UUID, nil
		}
	}
	if len(orgs) > 0 && orgs[0].UUID != "" {
		return orgs[0].UUID, nil
	}
	return "", errors.New("the account has no organization")
}
