package internal

import (
	"io"
	"os"
	"strings"
	"testing"

	ics "github.com/arran4/golang-ical"
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
	calendar, err := app.getCleanedCalendar([]byte(testData), []string{})
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
	calendar, err := app.getCleanedCalendar([]byte(testData), []string{})
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
	calendar, err := app.getCleanedCalendar([]byte(testData), []string{})
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

func TestCourseFiltering(t *testing.T) {
	testData, app := getTestData(t, "coursefiltering.ics")

	// make sure the unfiltered calendar has 2 entries
	fullCalendar, err := app.getCleanedCalendar([]byte(testData), []string{})
	if err != nil {
		t.Error(err)
		return
	}
	if len(fullCalendar.Components) != 2 {
		t.Errorf("Calendar should have 2 entries before course filtering but has %d", len(fullCalendar.Components))
		return
	}

	// now filter out one course
	filteredTag := "IN0004"
	filteredCalendar, err := app.getCleanedCalendar([]byte(testData), []string{filteredTag})
	if err != nil {
		t.Error(err)
		return
	}
	if len(filteredCalendar.Components) != 1 {
		t.Errorf("Calendar should have only 1 entry after course filtering but has %d", len(filteredCalendar.Components))
		return
	}

	// make sure the summary does not contain the filtered course's name
	summary := filteredCalendar.Components[0].(*ics.VEvent).GetProperty(ics.ComponentPropertySummary).Value
	if strings.Contains(summary, filteredTag) {
		t.Errorf("Summary should not contain %s but is %s", filteredTag, summary)
		return
	}
}
