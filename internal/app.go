package internal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const src = "https://campus.tum.de/tumonlinej/ws/termin/ical?pStud=%s&pToken=%s"

//go:embed courses.json
var coursesJson string

//go:embed buildings.json
var buildingsJson string

type App struct {
	engine *gin.Engine

	courseReplacements   map[string]string
	buildingReplacements map[string]string
}

func (a *App) Run() error {
	err := json.Unmarshal([]byte(coursesJson), &a.courseReplacements)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(buildingsJson), &a.buildingReplacements)
	if err != nil {
		return err
	}
	gin.SetMode("release")
	a.engine = gin.New()
	a.engine.Use(gin.Logger(), gin.Recovery())
	a.configRoutes()
	return a.engine.Run()
}

func (a *App) configRoutes() {
	a.engine.Any("/", a.handleIcal)
	a.engine.NoMethod(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotImplemented)
	})
}

func (a *App) handleIcal(c *gin.Context) {
	stud := c.Query("pStud")
	token := c.Query("pToken")
	if stud == "" || token == "" {
		return
	}
	log.Println(stud, token)
	resp, err := http.Get(fmt.Sprintf(src, stud, token))
	if err != nil {
		return
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	cal, err := ics.ParseCalendar(strings.NewReader(string(all)))
	if err != nil {
		return
	}
	hasLecture := make(map[string]bool)
	for i, event := range cal.Events() {
		dedupKey := fmt.Sprintf("%s-%s", event.GetProperty(ics.ComponentPropertySummary).Value, event.GetProperty(ics.ComponentPropertyDtStart))
		if _, ok := hasLecture[dedupKey]; ok {
			cal.Events()[i] = nil
			log.Println("skipping ", dedupKey)
			continue
		}
		hasLecture[dedupKey] = true
		a.cleanEvent(event)
	}

	response := []byte(cal.Serialize())
	c.Header("Content-Type", "text/calendar")
	c.Header("Content-Length", fmt.Sprintf("%d", len(response)))
	c.Writer.Write(response)
}

// matches tags like (IN0001) or [MA2012] and everything after.
// unfortunate also matches wrong brackets like [MA123) but hey…
var reTag = regexp.MustCompile(" ?[\\[(](MA|IN|WI|WIB)[0-9]+((_|-|,)[a-zA-Z0-9]+)*[\\])].*")

// Matches location and teacher from language course title
var reLoc = regexp.MustCompile(" ?(München|Garching|Weihenstephan).+")

// Matches repeated whitespaces
var reSpace = regexp.MustCompile(`\s\s+`)

var unneeded = []string{
	"Standardgruppe",
	"PR",
	"VO",
	"FA",
	"VI",
	"TT",
	"UE",
	"SE",
	"(Limited places)",
	"(Online)",
}

var reRoom = regexp.MustCompile("^(.*?),.*(\\d{4})\\.(?:\\d\\d|EG|UG|DG|Z\\d|U\\d)\\.\\d+")

func (a *App) cleanEvent(event *ics.VEvent) {
	summary := strings.ReplaceAll(event.GetProperty(ics.ComponentPropertySummary).Value, "\\", "")
	description := strings.ReplaceAll(event.GetProperty(ics.ComponentPropertyDescription).Value, "\\", "")
	location := strings.ReplaceAll(event.GetProperty(ics.ComponentPropertyLocation).Value, "\\", "")

	//Remove the TAG and anything after e.g.: (IN0001) or [MA0001]
	summary = reTag.ReplaceAllString(summary, "")
	//remove location and teacher from language course title
	summary = reLoc.ReplaceAllString(summary, "")
	summary = reSpace.ReplaceAllString(summary, "")
	for _, replace := range unneeded {
		summary = strings.ReplaceAll(summary, replace, "")
	}

	/*
			todo, whatever this does:
			//Clean up extra info for language course names
		    if(preg_match('/(Spanisch|Französisch)\s(A|B|C)(1|2)((\.(1|2))|(\/(A|B|C)(1|2)))?(\s)/', $summary, $matches, PREG_OFFSET_CAPTURE) === 1){
		       $summary = substr($summary, 0, $matches[10][1]);
		    }
	*/

	event.SetSummary(summary)

	//Remember the old title in the description
	description = summary + "\n" + description

	results := reRoom.FindStringSubmatch(location)
	if len(results) == 3 {
		if building, ok := a.buildingReplacements[results[2]]; ok {
			description = location + "\n" + description
			event.SetLocation(building)
		}
	}
	event.SetDescription(description)

	// set title on summary:
	for k, v := range a.courseReplacements {
		summary = strings.ReplaceAll(summary, k, v)
	}
	event.SetSummary(summary)
	switch event.GetProperty(ics.ComponentPropertyStatus).Value {
	case "CONFIRMED":
		event.SetStatus(ics.ObjectStatusConfirmed)
	case "CANCELLED":
		event.SetStatus(ics.ObjectStatusCancelled)
	case "TENTATIVE":
		event.SetStatus(ics.ObjectStatusTentative)
	}
}
