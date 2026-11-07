package account

import (
	"encoding/json"
	"io"
	"net/http"
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
	if _, err := fetchUsage(strings.Repeat("s", 12), "org-last"); err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("err = %v", err)
	}
}
