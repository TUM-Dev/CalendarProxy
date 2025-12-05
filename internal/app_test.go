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
	calendar, err := app.getCleanedCalendar([]byte(testData), map[string]bool{})
	if err != nil {
		t.Error(err)
		return
	}
	if len(calendar.Components) != 1 {
		t.Errorf("Calendar should have only 1 entry after deduplication but has %d", len(calendar.Components))
		return
	}

	// Verify that the additional room from the deduplicated event is in the description
	desc := calendar.Components[0].(*ics.VEvent).GetProperty(ics.ComponentPropertyDescription).Value
	if !strings.Contains(desc, "Additional rooms:") {
		t.Error("Description should contain 'Additional rooms:' when events are deduplicated with different locations")
		return
	}
	if !strings.Contains(desc, "MI HS 1") {
		t.Error("Description should contain the additional room 'MI HS 1' from the deduplicated event")
		return
	}
}

func TestMultipleRooms(t *testing.T) {
	// Setup app with building replacements
	app, err := newApp()
	if err != nil {
		t.Fatal(err)
	}

	// Create a dummy event with multiple rooms
	event := ics.NewEvent("test-uid")

	// Using 5508 and 5612 which are present in buildings.json
	// 5508 -> Boltzmannstr. 15, 85748 Garching b. München
	// 5612 -> Boltzmannstr. 3, 85748 Garching b. München

	// Construct a location string with multiple rooms.
	// RoomA, DescA (5508.01.001), RoomB, DescB (5612.01.001)

	location := "RoomA, DescA (5508.01.001), RoomB, DescB (5612.01.001)"
	event.SetProperty(ics.ComponentPropertyLocation, location)
	event.SetProperty(ics.ComponentPropertySummary, "Test Event")
	event.SetProperty(ics.ComponentPropertyDescription, "Original Description")
	event.SetProperty(ics.ComponentPropertyStatus, "CONFIRMED")

	app.cleanEvent(event)

	desc := event.GetProperty(ics.ComponentPropertyDescription).Value
	loc := event.GetProperty(ics.ComponentPropertyLocation).Value

	// Check if both rooms are present in description or nav links
	if !strings.Contains(desc, "5508.01.001") {
		t.Errorf("Description should contain first room ID")
	}
	if !strings.Contains(desc, "5612.01.001") {
		t.Errorf("Description should contain second room ID")
	}

	// Check if nav links are generated for both
	// 5508.01.001 -> https://nav.tum.de/room/5508.01.001
	// 5612.01.001 -> https://nav.tum.de/room/5612.01.001

	if !strings.Contains(desc, "https://nav.tum.de/room/5508.01.001") {
		t.Error("Missing nav link for first room")
	}
	if !strings.Contains(desc, "https://nav.tum.de/room/5612.01.001") {
		t.Error("Missing nav link for second room")
	}

	// With non-greedy regex, the location should be the first building (5508)
	expectedLoc := "Boltzmannstr. 15, 85748 Garching b. München"
	if loc != expectedLoc {
		t.Errorf("Location should be %s but is %s", expectedLoc, loc)
	}
}

func TestNameShortening(t *testing.T) {
	testData, app := getTestData(t, "nameshortening.ics")
	calendar, err := app.getCleanedCalendar([]byte(testData), map[string]bool{})
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
	calendar, err := app.getCleanedCalendar([]byte(testData), map[string]bool{})
	if err != nil {
		t.Error(err)
		return
	}
	location := calendar.Components[0].(*ics.VEvent).GetProperty(ics.ComponentPropertyLocation).Value
	expectedLocation := "Boltzmannstr. 15, 85748 Garching b. München"
	if location != expectedLocation {
		t.Errorf("Location should be shortened to %s but is %s", expectedLocation, location)
		return
	}
	desc := calendar.Components[0].(*ics.VEvent).GetProperty(ics.ComponentPropertyDescription).Value
	expectedDescription := "Additional rooms:\nMI HS 1\n\nhttps://nav.tum.de/room/5508.02.801\nMW 1801, Ernst-Schmidt-Hörsaal (5508.02.801)\nEinführung in die Rechnerarchitektur\nfix; Abhaltung;"
	if desc != expectedDescription {
		t.Errorf("Description should be %s but is %s", expectedDescription, desc)
		return
	}
}

func TestCourseFiltering(t *testing.T) {
	testData, app := getTestData(t, "coursefiltering.ics")

	// make sure the unfiltered calendar has 2 entries
	fullCalendar, err := app.getCleanedCalendar([]byte(testData), map[string]bool{})
	if err != nil {
		t.Error(err)
		return
	}
	if len(fullCalendar.Components) != 2 {
		t.Errorf("Calendar should have 2 entries before course filtering but has %d", len(fullCalendar.Components))
		return
	}

	// now filter out one course
	filter := "Einführung in die Rechnerarchitektur (IN0004) VO, Standardgruppe"
	filteredCalendar, err := app.getCleanedCalendar([]byte(testData), map[string]bool{filter: true})
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
