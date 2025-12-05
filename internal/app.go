package internal

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

type Course struct {
	Summary string `json:"summary"`
	Hide    bool   `json:"hide"`
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

func customLogFormatter(params gin.LogFormatterParams) string {
	return fmt.Sprintf("[GIN] %v |%s %3d %s | %13v | %15s |%s %-7s%s %#v\n%s",
		params.TimeStamp.Format("2006/01/02 - 15:04:05"),
		params.StatusCodeColor(),
		params.StatusCode,
		params.ResetColor(),
		params.Latency,
		params.ClientIP,
		params.MethodColor(),
		params.Method,
		params.ResetColor(),
		hideTokens(params.Path),
		params.ErrorMessage,
	)
}

func hideTokens(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}

	pStud := u.Query().Get("pStud")
	pPers := u.Query().Get("pPers")
	pToken := u.Query().Get("pToken")

	if pToken == "" || (pStud == "" && pPers == "") {
		return path
	}

	manyXes := strings.Repeat("X", 12)
	tokenReplaced := pToken[:4] + manyXes
	if pStud != "" {
		return fmt.Sprintf("/?pStud=%s&pToken=%s", pStud[:4]+manyXes, tokenReplaced)
	}
	return fmt.Sprintf("/?pPers=%s&pToken=%s", pPers[:4]+manyXes, tokenReplaced)
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
	logger := gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/health"}, Formatter: customLogFormatter})
	a.engine.Use(logger, gin.Recovery())
	a.configRoutes()

	// Start the engines
	return a.engine.Run(":4321")
}

func (a *App) configRoutes() {
	a.engine.GET("/api/courses", a.handleGetCourses)
	a.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
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

func getCalendar(ctx *gin.Context) ([]byte, map[string]bool, error) {
	fetchURL := getUrl(ctx)
	if fetchURL == "" {
		return nil, nil, errors.New("no fetchable URL passed")
	}
	resp, err := http.Get(fetchURL)
	if err != nil {
		return nil, nil, fmt.Errorf("can't fetch calendar: %w", err)
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("can't read calendar: %w", err)
	}

	// Create map of all hidden courses
	hide := ctx.QueryArray("hide")
	hiddenCourses := make(map[string]bool)
	for _, course := range hide {
		hiddenCourses[course] = true
	}

	return all, hiddenCourses, nil
}

// handleIcal returns a filtered calendar with all courses that are currently offered on campus.
func (a *App) handleIcal(ctx *gin.Context) {
	allEvents, hiddenCourses, err := getCalendar(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	cleaned, err := a.getCleanedCalendar(allEvents, hiddenCourses)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response := []byte(cleaned.Serialize())
	ctx.Header("Content-Type", "text/calendar")
	ctx.Header("Content-Length", fmt.Sprintf("%d", len(response)))

	if _, err := ctx.Writer.Write(response); err != nil {
		sentry.CaptureException(err)
	}
}

// handleGetCourses returns a list of all courses that are currently offered on campus.
// This is used to populate the dropdown in the landing page for hiding courses.
func (a *App) handleGetCourses(ctx *gin.Context) {
	allEvents, hidden, err := getCalendar(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	cal, err := a.getCleanedCalendar(allEvents, map[string]bool{})
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// detect all courses, de-duplicate them by their summary (lecture name)
	courses := make(map[string]Course)
	for _, component := range cal.Components {
		switch component.(type) {
		case *ics.VEvent:
			eventSummary := cleanEventSummary(component.(*ics.VEvent).GetProperty(ics.ComponentPropertySummary).Value)
			if _, exists := courses[eventSummary]; !exists {
				courses[eventSummary] = Course{
					Summary: eventSummary,
					// Check for existing hidden course, that might want to be updated
					Hide: hidden[eventSummary],
				}
			}
			log.Printf("summaries: %s", eventSummary)
		default:
			continue
		}
	}

	ctx.JSON(http.StatusOK, courses)
}

func (a *App) getCleanedCalendar(all []byte, hiddenCourses map[string]bool) (*ics.Calendar, error) {
	cal, err := ics.ParseCalendar(strings.NewReader(string(all)))
	if err != nil {
		return nil, err
	}

	// Create map that tracks if we have already seen a lecture name & datetime (e.g. "lecturexyz-1.2.2024 10:00" -> true)
	hasLecture := make(map[string]bool)
	var newComponents []ics.Component // saves the components we keep because they are not duplicated

	for _, component := range cal.Components {
		switch component.(type) {
		case *ics.VEvent:
			event := component.(*ics.VEvent)

			// check if the summary contains any of the hidden keys, and if yes, skip it
			eventSummary := event.GetProperty(ics.ComponentPropertySummary).Value
			if hiddenCourses[eventSummary] {
				continue
			}

			// deduplicate lectures by their summary and datetime
			dedupKey := fmt.Sprintf("%s-%s", event.GetProperty(ics.ComponentPropertySummary).Value, event.GetProperty(ics.ComponentPropertyDtStart))
			if _, ok := hasLecture[dedupKey]; ok {
				continue
			}
			hasLecture[dedupKey] = true // mark event as seen

			// clean up the event
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

// Matches unique starting numbers like "0000002467 " in "0000002467 Semantik"
var reWeirdStartingNumbers = regexp.MustCompile(`^0\d+ `)

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

// matches strings like: (5612.03.017), (5612.EG.017), (5612.EG.010B)
var reNavigaTUM = regexp.MustCompile("\\(\\d{4}\\.[a-zA-Z0-9]{2}\\.\\d{3}[A-Z]?\\)")

func (a *App) cleanEvent(event *ics.VEvent) {
	summary := ""
	if s := event.GetProperty(ics.ComponentPropertySummary); s != nil {
		summary = cleanEventSummary(s.Value)
	}

	description := ""
	if d := event.GetProperty(ics.ComponentPropertyDescription); d != nil {
		description = d.Value
	}

	location := ""
	if l := event.GetProperty(ics.ComponentPropertyLocation); l != nil {
		location = event.GetProperty(ics.ComponentPropertyLocation).Value
	}

	// Remove the TAG and anything after e.g.: (IN0001) or [MA0001]
	summary = reTag.ReplaceAllString(summary, "")
	// remove location and teacher from the language course title
	summary = reLoc.ReplaceAllString(summary, "")
	summary = reSpace.ReplaceAllString(summary, "")
	for _, replace := range unneeded {
		summary = strings.ReplaceAll(summary, replace, "")
	}
	// sometimes the summary has weird numbers attached like "0000002467 " in "0000002467 Semantik"
	// What the heck? And why only sometimes???
	summary = reWeirdStartingNumbers.ReplaceAllString(summary, "")

	event.SetSummary(summary)

	// Remember the old title in the description
	description = summary + "\n" + description

	results := reRoom.FindStringSubmatch(location)
	if len(results) == 3 {
		if building, ok := a.buildingReplacements[results[2]]; ok {
			description = location + "\n" + description
			event.SetLocation(building)
		}
		if roomID := reNavigaTUM.FindString(location); roomID != "" {
			roomID = strings.Trim(roomID, "()")
			description = fmt.Sprintf("https://nav.tum.de/room/%s\n%s", roomID, description)
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

func cleanEventSummary(eventSummary string) string {
	eventSummary = strings.TrimSpace(eventSummary)
	eventSummary = strings.TrimSuffix(eventSummary, " ,")
	return eventSummary
}
