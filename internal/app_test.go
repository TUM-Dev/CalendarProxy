package internal

import (
	"bytes"
	"io"
	"os"
	"slices"
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

func TestReplacement(t *testing.T) {
	r1 := Replacement{"b", "b"}
	if r1.isLessThan(&r1) {
		t.Error("Replacement should not be less than itself")
		return
	}
	if r1.isLessThan(&Replacement{key: "longer key first"}) {
		t.Error("Replacement should sort longer prefix first")
		return
	}
	if !r1.isLessThan(&Replacement{key: ""}) {
		t.Error("Replacement should sort longer prefix first")
		return
	}
	if r1.isLessThan(&Replacement{key: "a"}) {
		t.Error("Replacement should sort alphabetically")
		return
	}
	if !r1.isLessThan(&Replacement{key: "c"}) {
		t.Error("Replacement should sort alphabetically")
		return
	}
	if r1.isLessThan(&Replacement{key: r1.key, value: "a"}) {
		t.Error("Replacement with equal key should sort alphabetically by value")
		return
	}
	if !r1.isLessThan(&Replacement{key: r1.key, value: "c"}) {
		t.Error("Replacement with equal key should sort alphabetically by value")
		return
	}

}

func TestDeduplication(t *testing.T) {
	testData, app := getTestData(t, "duplication.ics")
	calendar, err := app.getCleanedCalendar([]byte(testData), []string{})
	if err != nil {
		t.Error(err)
		return
	}
	if len(calendar.Components) != 2 {
		t.Errorf("Calendar should have only 2 entry after deduplication but has %d", len(calendar.Components))
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
	expectedLocation := "Boltzmannstr. 15\\, 85748 Garching b. München"
	if location != expectedLocation {
		t.Errorf("Location should be shortened to %s but is %s", expectedLocation, location)
		return
	}
	desc := calendar.Components[0].(*ics.VEvent).GetProperty(ics.ComponentPropertyDescription).Value
	expectedDescription := "https://nav.tum.de/room/5508.02.801\\nMW 1801\\, Ernst-Schmidt-Hörsaal (5508.02.801)\\nEinführung in die Rechnerarchitektur\\nfix\\; Abhaltung\\;"
	if desc != expectedDescription {
		t.Errorf("Description should be %s but is %s", expectedDescription, desc)
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
	filter := "Einführung in die Rechnerarchitektur (IN0004) VO\\, Standardgruppe"
	filteredCalendar, err := app.getCleanedCalendar([]byte(testData), []string{filter})
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
	if strings.Contains(summary, filter) {
		t.Errorf("Summary should not contain %s but is %s", filter, summary)
		return
	}
}

func TestGetCourses(t *testing.T) {
	testData, app := getTestData(t, "duplication.ics")
	courses, err := app.getCourses(bytes.NewReader([]byte(testData)))
	if err != nil {
		t.Error(err)
		return
	}
	should := []string{"Einführung in die Rechnerarchitektur (IN0004) VO\\, Standardgruppe", "Practical Course: Open Source Lab (IN0012\\, IN2106\\, IN4308) PR\\, Standardgruppe"}
	if !slices.Equal(courses, should) {
		t.Errorf("get courses failed, expected: %v, got %v", should, courses)
	}
}
