package internal

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"

	ics "github.com/arran4/golang-ical"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

//go:embed courses.json
var coursesJson string

//go:embed buildings.json
var buildingsJson string

//go:embed static
var static embed.FS

// Version is injected at build time by the compiler with the correct git-commit-sha or "dev" in development
var Version = "dev"

type App struct {
	engine *gin.Engine

	courseReplacements   []*Replacement
	buildingReplacements map[string]string
}

type Replacement struct {
	key   string
	value string
}

// for sorting replacements by length, then alphabetically
func (r1 *Replacement) isLessThan(r2 *Replacement) bool {
	if len(r1.key) != len(r2.key) {
		return len(r1.key) > len(r2.key)
	}
	if r1.key != r2.key {
		return r1.key < r2.key
	}
	return r1.value < r2.value
}

func newApp() (*App, error) {
	a := App{}

	// courseReplacements is a map of course names to shortened names.
	// We sort it by length, then alphabetically to ensure a consistent execution order
	var rawCourseReplacements map[string]string
	if err := json.Unmarshal([]byte(coursesJson), &rawCourseReplacements); err != nil {
		return nil, err
	}
	for key, value := range rawCourseReplacements {
		a.courseReplacements = append(a.courseReplacements, &Replacement{key, value})
	}
	sort.Slice(a.courseReplacements, func(i, j int) bool { return a.courseReplacements[i].isLessThan(a.courseReplacements[j]) })
	// buildingReplacements is a map of room numbers to building names
	if err := json.Unmarshal([]byte(buildingsJson), &a.buildingReplacements); err != nil {
		return nil, err
	}
	return &a, nil
}

func (a *App) Run() error {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://2fbc80ad1a99406cb72601d6a47240ce@glitch.exgen.io/4",
		Release:          Version,
		AttachStacktrace: true,
		EnableTracing:    true,
		// Specify a fixed sample rate: 10% will do for now
		TracesSampleRate: 0.1,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	// Create app struct
	a, err := newApp()
	if err != nil {
		return err
	}

	// Setup Gin with sentry traces, logger and routes
	gin.SetMode("release")
	a.engine = gin.New()
	a.engine.Use(sentrygin.New(sentrygin.Options{}))
	a.engine.Use(gin.Logger(), gin.Recovery())
	a.configRoutes()

	// Start the engines
	return a.engine.Run(":4321")
}

func (a *App) configRoutes() {
	a.engine.GET("/api/courses", a.handleGetCourses)
	a.engine.Any("/", a.handleIcal)
	f := http.FS(static)
	a.engine.StaticFS("/files/", f)
	a.engine.NoMethod(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotImplemented)
	})
}

func getUrl(c *gin.Context) string {
	stud := c.Query("pStud")
	pers := c.Query("pPers")
	token := c.Query("pToken")
	if (stud == "" && pers == "") || token == "" {
		// Missing parameters: just serve our landing page
		f, err := static.Open("static/index.html")
		if err != nil {
			sentry.CaptureException(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return ""
		}

		if _, err := io.Copy(c.Writer, f); err != nil {
			sentry.CaptureException(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return ""
		}
		return ""
	}
	if stud == "" {
		return fmt.Sprintf("https://campus.tum.de/tumonlinej/ws/termin/ical?pPers=%s&pToken=%s", pers, token)
	}
	return fmt.Sprintf("https://campus.tum.de/tumonlinej/ws/termin/ical?pStud=%s&pToken=%s", stud, token)
}

func (a *App) handleIcal(c *gin.Context) {
	url := getUrl(c)
	if url == "" {
		return
	}
	resp, err := http.Get(url)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	filters := c.QueryArray("filter")
	cleaned, err := a.getCleanedCalendar(calData, filters)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response := []byte(cleaned.Serialize())
	c.Header("Content-Type", "text/calendar")
	c.Header("Content-Length", fmt.Sprintf("%d", len(response)))

	if _, err := c.Writer.Write(response); err != nil {
		sentry.CaptureException(err)
	}
}

func (a *App) handleGetCourses(c *gin.Context) {
	url := getUrl(c)
	if url == "" {
		return
	}
	resp, err := http.Get(url)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	cal, err := ics.ParseCalendar(strings.NewReader(string(all)))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	// detect all courses, de-duplicate them by their summary (lecture name)
	courses := make(map[string]bool)
	for _, component := range cal.Components {
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)
			eventSummary := event.GetProperty(ics.ComponentPropertySummary).Value
			courses[eventSummary] = true

		default:
			continue
		}
	}

	c.JSON(http.StatusOK, courses)
}

func stringEqualsOneOf(target string, listOfStrings []string) bool {
	for _, element := range listOfStrings {
		if target == element {
			return true
		}
	}
	return false
}

func (a *App) getCleanedCalendar(all []byte, filters []string) (*ics.Calendar, error) {
	cal, err := ics.ParseCalendar(strings.NewReader(string(all)))
	if err != nil {
		return nil, err
	}

	// Create map that tracks if we have allready seen a lecture name & datetime (e.g. "lecturexyz-1.2.2024 10:00" -> true)
	hasLecture := make(map[string]bool)
	var newComponents []ics.Component // saves the components we keep because they are not duplicated

	for _, component := range cal.Components {
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)

			// check if the summary contains any of the filtered keys, and if yes, skip it
			eventSummary := event.GetProperty(ics.ComponentPropertySummary).Value
			shouldFilterEvent := stringEqualsOneOf(eventSummary, filters)
			if shouldFilterEvent {
				continue
			}

			dedupKey := fmt.Sprintf("%s-%s", event.GetProperty(ics.ComponentPropertySummary).Value, event.GetProperty(ics.ComponentPropertyDtStart))
			if _, ok := hasLecture[dedupKey]; ok {
				continue
			}
			hasLecture[dedupKey] = true // mark event as seen
			a.cleanEvent(event)
			newComponents = append(newComponents, event)
		default: // keep everything that is not an event (metadata etc.)
			newComponents = append(newComponents, component)
		}
	}
	cal.Components = newComponents
	return cal, nil
}

// matches tags like (IN0001) or [MA2012] and everything after.
// unfortunate also matches wrong brackets like [MA123) but hey…
var reTag = regexp.MustCompile(" ?[\\[(](ED|MW|SOM|CIT|MA|IN|WI|WIB)[0-9]+((_|-|,)[a-zA-Z0-9]+)*[\\])].*")

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
	summary := ""
	if s := event.GetProperty(ics.ComponentPropertySummary); s != nil {
		summary = strings.ReplaceAll(s.Value, "\\", "")
	}

	description := ""
	if d := event.GetProperty(ics.ComponentPropertyDescription); d != nil {
		description = strings.ReplaceAll(d.Value, "\\", "")
	}

	location := ""
	if l := event.GetProperty(ics.ComponentPropertyLocation); l != nil {
		location = strings.ReplaceAll(event.GetProperty(ics.ComponentPropertyLocation).Value, "\\", "")
	}

	//Remove the TAG and anything after e.g.: (IN0001) or [MA0001]
	summary = reTag.ReplaceAllString(summary, "")
	//remove location and teacher from language course title
	summary = reLoc.ReplaceAllString(summary, "")
	summary = reSpace.ReplaceAllString(summary, "")
	for _, replace := range unneeded {
		summary = strings.ReplaceAll(summary, replace, "")
	}

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
	for _, repl := range a.courseReplacements {
		summary = strings.ReplaceAll(summary, repl.key, repl.value)
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
