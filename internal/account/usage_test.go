package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

const sampleUsage = `{
	"five_hour": {"utilization": 42.0, "resets_at": "2026-11-07T04:59:59.943648+00:00"},
	"seven_day": {"utilization": 7, "resets_at": "2026-11-12T08:00:00+00:00"},
	"seven_day_opus": {"utilization": 0, "resets_at": null},
	"seven_day_oauth_apps": null
}`

func TestUsageReadsTheFiveHourAndWeeklyWindows(t *testing.T) {
	var response usageResponse
	if err := json.Unmarshal([]byte(sampleUsage), &response); err != nil {
		t.Fatal(err)
	}
	checked := time.Date(2026, 11, 7, 1, 0, 0, 0, time.UTC)
	usage := usageFrom(response, checked)
	if !usage.Known || usage.FiveHour.Percent != 42 || usage.Week.Percent != 7 || !usage.Checked.Equal(checked) {
		t.Fatalf("usage = %+v", usage)
	}
	if want := time.Date(2026, 11, 7, 4, 59, 59, 943648000, time.UTC); !usage.FiveHour.Resets.Equal(want) {
		t.Fatalf("five-hour reset = %v, want %v", usage.FiveHour.Resets, want)
	}
}

func TestUsageWithoutWindowsIsUnknown(t *testing.T) {
	var response usageResponse
	if err := json.Unmarshal([]byte(`{"five_hour": null, "seven_day": null}`), &response); err != nil {
		t.Fatal(err)
	}
	if usage := usageFrom(response, time.Now()); usage.Known {
		t.Fatalf("usage = %+v", usage)
	}
}

func TestLimitStartsOverAtItsReset(t *testing.T) {
	reset := time.Date(2026, 11, 7, 5, 0, 0, 0, time.UTC)
	limit := Limit{Percent: 90, Resets: reset}
	if got := limit.At(reset.Add(-time.Minute)); got.Percent != 90 {
		t.Fatalf("before the reset: %+v", got)
	}
	if got := limit.At(reset); got.Percent != 0 || !got.Resets.IsZero() {
		t.Fatalf("at the reset: %+v", got)
	}
	if got := (Limit{Percent: 30}).At(reset); got.Percent != 30 {
		t.Fatalf("without a reset time: %+v", got)
	}
}

// fakeClaude answers like claude.ai for the paths it knows, and only to the
// session it was given.
type fakeClaude struct {
	session string
	bodies  map[string]string
	asked   []string
}

func (f *fakeClaude) RoundTrip(request *http.Request) (*http.Response, error) {
	f.asked = append(f.asked, request.URL.Path)
	status, body := http.StatusNotFound, ""
	if request.Header.Get("Cookie") != "sessionKey="+f.session {
		status = http.StatusForbidden
	} else if known, ok := f.bodies[request.URL.Path]; ok {
		status, body = http.StatusOK, known
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: request}, nil
}

func useFakeClaude(t *testing.T, fake *fakeClaude) {
	t.Helper()
	previous := defaultClient.Transport
	defaultClient.Transport = fake
	t.Cleanup(func() { defaultClient.Transport = previous })
}

func TestFetchUsageAsksTheOrganizationClaudeLastUsed(t *testing.T) {
	session := strings.Repeat("s", 12)
	fake := &fakeClaude{session: session, bodies: map[string]string{"/api/organizations/org-last/usage": sampleUsage}}
	useFakeClaude(t, fake)
	usage, err := fetchUsage(session, "org-last")
	if err != nil || usage.FiveHour.Percent != 42 {
		t.Fatalf("usage = %+v, err = %v", usage, err)
	}
	if len(fake.asked) != 1 {
		t.Fatalf("asked %v", fake.asked)
	}
}

func TestFetchUsageFallsBackToTheSubscription(t *testing.T) {
	session := strings.Repeat("s", 12)
	fake := &fakeClaude{session: session, bodies: map[string]string{
		"/api/organizations":                `[{"uuid": "org-api", "capabilities": ["api"]}, {"uuid": "org-team", "capabilities": ["chat", "claude_team"]}]`,
		"/api/organizations/org-team/usage": sampleUsage,
	}}
	useFakeClaude(t, fake)
	usage, err := fetchUsage(session, "")
	if err != nil || usage.Week.Percent != 7 {
		t.Fatalf("usage = %+v, err = %v", usage, err)
	}
}

func TestFetchUsageReportsARefusal(t *testing.T) {
	useFakeClaude(t, &fakeClaude{session: "other", bodies: map[string]string{"/api/organizations/org-last/usage": sampleUsage}})
	_, err := fetchUsage(strings.Repeat("s", 12), "org-last")
	if err == nil || !strings.Contains(err.Error(), "403") || errors.Is(err, errBrowserCheck) {
		t.Fatalf("err = %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

// Seen live on 2026-11-08: five requests at once, and Cloudflare answered one
// with 403, cf-mitigated: challenge and its "Just a moment..." page.
func TestFetchUsageRecognizesTheBrowserCheck(t *testing.T) {
	previous := defaultClient.Transport
	defaultClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		header := http.Header{"Cf-Mitigated": []string{"challenge"}, "Content-Type": []string{"text/html; charset=UTF-8"}}
		return &http.Response{StatusCode: http.StatusForbidden, Header: header, Body: io.NopCloser(strings.NewReader("<title>Just a moment...</title>")), Request: request}, nil
	})
	t.Cleanup(func() { defaultClient.Transport = previous })
	if _, err := fetchUsage(strings.Repeat("s", 12), "org-last"); !errors.Is(err, errBrowserCheck) {
		t.Fatalf("err = %v", err)
	}
}

func TestFetchEachAsksOneAtATimeAndOnceMoreAfterABrowserCheck(t *testing.T) {
	tests := []struct {
		name    string
		checks  int
		asked   []string
		waits   int
		checked bool
	}{
		{"no check", 0, []string{"a", "b", "c"}, 0, false},
		{"one check, then an answer", 1, []string{"a", "b", "c", "b"}, 1, false},
		{"checked twice", 2, []string{"a", "b", "c", "b"}, 1, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			left := test.checks
			var asked []string
			fetch := func(path string) (Usage, error) {
				asked = append(asked, path)
				if path == "b" && left > 0 {
					left--
					return Usage{}, fmt.Errorf("%w (HTTP 403)", errBrowserCheck)
				}
				return Usage{Known: true}, nil
			}
			waits := 0
			results := fetchEach(map[string]string{"c": "c", "a": "a", "b": "b"}, fetch, func() { waits++ })
			if !slices.Equal(asked, test.asked) || waits != test.waits {
				t.Fatalf("asked %v with %d waits, want %v with %d", asked, waits, test.asked, test.waits)
			}
			if got := errors.Is(results["b"].Err, errBrowserCheck); got != test.checked || !results["a"].Known || !results["c"].Known {
				t.Fatalf("results = %+v", results)
			}
		})
	}
}
