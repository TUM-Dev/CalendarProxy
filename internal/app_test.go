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
	calendar, err := app.getCleanedCalendar([]byte(testData), []string{}, map[int]int{}, map[int]int{})
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
	calendar, err := app.getCleanedCalendar([]byte(testData), []string{}, map[int]int{}, map[int]int{})
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
	calendar, err := app.getCleanedCalendar([]byte(testData), []string{}, map[int]int{}, map[int]int{})
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
	fullCalendar, err := app.getCleanedCalendar([]byte(testData), []string{}, map[int]int{}, map[int]int{})
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
	filteredCalendar, err := app.getCleanedCalendar([]byte(testData), []string{filter}, map[int]int{}, map[int]int{})
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

func TestCourseTimeAdjustment(t *testing.T) {
	testData, app := getTestData(t, "timeadjustment.ics")

    startOffsets := map[int]int {
      583745: 15,
    }

    endOffsets := map[int]int {
      583745: 0,
      583744: -14,
    }

	adjCal, err := app.getCleanedCalendar([]byte(testData), []string{}, startOffsets, endOffsets)
	if err != nil {
		t.Error(err)
		return
	}
	if len(adjCal.Components) != 4 {
		t.Errorf("Calendar should have 4 entries before time adjustment but was %d", len(adjCal.Components))
		return
	}

    // first entry (recurring id 583745): only start offset expected (+15)
    if start := adjCal.Components[0].(*ics.VEvent).GetProperty(ics.ComponentProperty(ics.PropertyDtstart)).Value; start != "20240109T171500Z" {
		t.Errorf("start (+15) should have been 20240109T171500Z but was %s", start)
		return
	}
    if end := adjCal.Components[0].(*ics.VEvent).GetProperty(ics.ComponentProperty(ics.PropertyDtend)).Value; end != "20240109T190000Z" {
		t.Errorf("end (+0) should have been 20240109T190000Z but was %s", end)
		return
	}

    // second entry (recurring id 583744): only end offset expected (-14)
    if start := adjCal.Components[1].(*ics.VEvent).GetProperty(ics.ComponentProperty(ics.PropertyDtstart)).Value; start != "20231113T170000Z" {
		t.Errorf("start (+/- n.a.) should have been 20231113T170000Z but was %s", start)
		return
	}
    if end := adjCal.Components[1].(*ics.VEvent).GetProperty(ics.ComponentProperty(ics.PropertyDtend)).Value; end != "20231113T184600Z" {
		t.Errorf("end (-14) should have been 20231113T184600Z but was %s", end)
		return
	}

    // third entry (no recurring id): expect no adjustments
    if start := adjCal.Components[2].(*ics.VEvent).GetProperty(ics.ComponentProperty(ics.PropertyDtstart)).Value; start != "20231023T160000Z" {
		t.Errorf("start (+/- n.a.) should have been 20231023T160000Z but was %s", start)
		return
	}
    if end := adjCal.Components[2].(*ics.VEvent).GetProperty(ics.ComponentProperty(ics.PropertyDtend)).Value; end != "20231023T180000Z" {
		t.Errorf("end (+/- n.a.) should have been 20231023T180000Z but was %s", end)
		return
	}

    // fourth entry (recurring id 583745): expect only start offset expected (+15)
    if start := adjCal.Components[3].(*ics.VEvent).GetProperty(ics.ComponentProperty(ics.PropertyDtstart)).Value; start != "20240206T171500Z" {
		t.Errorf("start (+/- n.a.) should have been 20240206T171500Z but was %s", start)
		return
	}
    if end := adjCal.Components[3].(*ics.VEvent).GetProperty(ics.ComponentProperty(ics.PropertyDtend)).Value; end != "20240206T190000Z" {
		t.Errorf("end (+/- n.a.) should have been 20240206T190000Z but was %s", end)
		return
	}
}
