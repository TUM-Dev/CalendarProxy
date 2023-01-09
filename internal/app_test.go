package internal

import (
	ics "github.com/arran4/golang-ical"
	"io"
	"os"
	"testing"
)

func getTestData(t *testing.T, name string) (string, *App) {
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal("can't open testdata")
	}
	all, err := io.ReadAll(f)
	if err != nil {
		t.Fatal("can't read testdata")
	}
	app, err := newApp()
	if err != nil {
		t.Fatal("can't create test subject", err)
	}
	return string(all), app
}

func TestDeduplication(t *testing.T) {
	testData, app := getTestData(t, "duplication.ics")
	calendar, err := app.getCleanedCalendar([]byte(testData))
	if err != nil {
		t.Error(err)
		return
	}
	if len(calendar.Components) != 1 {
		t.Errorf("Calendar should have only 1 entry after deduplication but has %d", len(calendar.Components))
		return
	}
}

func TestNameShortening(t *testing.T) {
	testData, app := getTestData(t, "nameshortening.ics")
	calendar, err := app.getCleanedCalendar([]byte(testData))
	if err != nil {
		t.Error(err)
		return
	}
	summary := calendar.Components[0].(*ics.VEvent).GetProperty(ics.ComponentPropertySummary).Value
	if summary != "ERA" {
		t.Errorf("Einführung in die Rechnerarchitektur (IN0004) VO, Standardgruppe should be shortened to ERA but is %s", summary)
		return
	}
}
