package internal

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
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

	courseReplacements   map[string]string
	buildingReplacements map[string]string
}

func newApp() (*App, error) {
	a := App{}
	err := json.Unmarshal([]byte(coursesJson), &a.courseReplacements)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(buildingsJson), &a.buildingReplacements)
	if err != nil {
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
	newApp, err := newApp()
	if err != nil {
		return err
	}
	a = newApp

	// Setup Gin with sentry traces, logger and routes
	gin.SetMode("release")
	a.engine = gin.New()
	a.engine.Use(sentrygin.New(sentrygin.Options{}))
	a.engine.Use(gin.Logger(), gin.Recovery())
	a.configRoutes()

	// Start the engines
	return a.engine.Run(":80")
}

func (a *App) configRoutes() {
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return ""
		}
		io.Copy(c.Writer, f)
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
	cleaned, err := a.getCleanedCalendar(all)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response := []byte(cleaned.Serialize())
	c.Header("Content-Type", "text/calendar")
	c.Header("Content-Length", fmt.Sprintf("%d", len(response)))
	c.Writer.Write(response)
}

func (a *App) getCleanedCalendar(all []byte) (*ics.Calendar, error) {
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
