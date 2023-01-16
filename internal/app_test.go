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

func TestLocationReplacement(t *testing.T) {
	testData, app := getTestData(t, "location.ics")
	calendar, err := app.getCleanedCalendar([]byte(testData))
	if err != nil {
		t.Error(err)
		return
	}
	location := calendar.Components[0].(*ics.VEvent).GetProperty(ics.ComponentPropertyLocation).Value
	if location != "Boltzmannstr. 15\\, 85748 Garching b. München" {
		t.Errorf("MW 1801\\, Ernst-Schmidt-Hörsaal (5508.02.801) should be shortened to Boltzmannstr. 15\\, 85748 Garching b. München but is %s", location)
		return
	}
	desc := calendar.Components[0].(*ics.VEvent).GetProperty(ics.ComponentPropertyDescription).Value
	if desc != "MW 1801\\, Ernst-Schmidt-Hörsaal (5508.02.801)\\nEinführung in die Rechnerarchitektur\\nfix\\; Abhaltung\\;" {
		t.Errorf("Description should be MW 1801\\, Ernst-Schmidt-Hörsaal (5508.02.801)\\nEinführung in die Rechnerarchitektur\\nfix\\; Abhaltung\\; but is %s", desc)
		return
	}
}
