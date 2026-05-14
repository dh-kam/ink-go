package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dh-kam/ink-go/internal/ttyinput"
	"github.com/dh-kam/ink-go/pkg/ink"
	"github.com/dh-kam/ink-go/pkg/terminal"
	"github.com/dh-kam/ink-go/pkg/vdom"
)

var incrementalRows = []string{
	"Server Authentication Module - Handles JWT token validation, OAuth2 flows, and session management across distributed systems",
	"Database Connection Pool - Maintains persistent connections to PostgreSQL cluster with automatic failover and load balancing",
	"API Gateway Service - Routes incoming HTTP requests to microservices with rate limiting and request transformation",
	"User Profile Manager - Caches user data in Redis with write-through policy and invalidation strategies",
	"Payment Processing Engine - Integrates with Stripe, PayPal, and Square APIs for transaction processing",
	"Email Notification Queue - Processes outbound emails through SendGrid with retry logic and delivery tracking",
	"File Storage Handler - Manages S3 bucket operations with multipart uploads and CDN integration",
	"Search Indexer Service - Maintains Elasticsearch indices with real-time document updates and reindexing",
	"Metrics Aggregation Pipeline - Collects and processes telemetry data for Prometheus and Grafana dashboards",
	"WebSocket Connection Manager - Handles real-time bidirectional communication for chat and notifications",
	"Cache Invalidation Service - Coordinates distributed cache updates across Redis cluster nodes",
	"Background Job Processor - Executes async tasks via RabbitMQ with dead letter queue handling",
	"Session Store Manager - Persists user sessions in DynamoDB with TTL and cross-region replication",
	"Rate Limiter Module - Enforces API quotas using token bucket algorithm with Redis backend",
	"Content Delivery Network - Serves static assets through Cloudflare with edge caching and GZIP compression",
	"Logging Aggregator - Streams application logs to ELK stack with structured JSON formatting",
	"Health Check Monitor - Performs periodic service health checks with circuit breaker pattern implementation",
	"Configuration Manager - Loads environment-specific settings from Consul with hot reload capability",
	"Security Scanner Service - Runs automated vulnerability scans and dependency checks on deployed applications",
	"Backup Orchestrator - Schedules and executes automated database backups with encryption and versioning",
	"Load Balancer Controller - Manages NGINX upstream servers with health-based traffic distribution",
	"Container Orchestration - Coordinates Docker container lifecycle via Kubernetes with auto-scaling policies",
	"Message Bus Coordinator - Routes events through Apache Kafka topics with guaranteed delivery semantics",
	"Analytics Data Warehouse - Aggregates business metrics in Snowflake with incremental ETL processes",
	"API Documentation Service - Generates and serves OpenAPI specs with interactive Swagger UI",
	"Feature Flag Manager - Controls feature rollouts using LaunchDarkly with user targeting and percentage rollouts",
	"Audit Trail Logger - Records all user actions and system events for compliance and security analysis",
	"Image Processing Pipeline - Resizes and optimizes uploaded images using Sharp with multiple format outputs",
	"Geolocation Service - Resolves IP addresses to geographic coordinates using MaxMind GeoIP2 database",
	"Recommendation Engine - Generates personalized content suggestions using collaborative filtering algorithms",
}

type incrementalState struct {
	selectedIndex int
	timestamp     string
	counter       int
	fps           int
	progress1     int
	progress2     int
	progress3     int
	randomValue   int
	logLines      []string
	frameCount    int
	lastFPSAt     time.Time
	random        *rand.Rand
}

type incrementalSnapshot struct {
	selectedIndex int
	timestamp     string
	counter       int
	fps           int
	progress1     int
	progress2     int
	progress3     int
	randomValue   int
	logLines      []string
}

var incrementalStateStore = struct {
	sync.Mutex
	state incrementalState
}{}

func IncrementalRenderingDemo() *vdom.Node {
	app := ink.UseApp()
	stdout := ink.UseStdout()
	terminalHeight := stdout.Rows
	if terminalHeight <= 0 {
		terminalHeight = 24
	}

	availableLines := maxInt(terminalHeight-15, 10)
	logLineCount := maxInt(availableLines*3/10, 3)
	serviceCount := minInt(maxInt(availableLines*7/10, 5), len(incrementalRows))
	snapshot := incrementalSnapshotFor(logLineCount, serviceCount)

	ink.UseInput(func(input string, key ink.InputKey) {
		switch {
		case key.UpArrow:
			moveIncrementalSelection(-1, serviceCount)
		case key.DownArrow:
			moveIncrementalSelection(1, serviceCount)
		case input == "q":
			app.Exit()
		}
	})

	progressBar := func(value int) string {
		filled := value / 5
		return strings.Repeat("█", filled) + strings.Repeat("░", 20-filled)
	}

	serviceRows := make([]*vdom.Node, 0, serviceCount+1)
	serviceRows = append(serviceRows,
		ink.Text(vdom.Props{"bold": true, "color": "magenta"}, fmt.Sprintf("System Services Monitor (%d of %d services):", serviceCount, len(incrementalRows))),
	)
	for index, row := range incrementalRows[:serviceCount] {
		prefix := "  "
		color := "white"
		if index == snapshot.selectedIndex {
			prefix = "> "
			color = "blue"
		}
		serviceRows = append(serviceRows, ink.Text(vdom.Props{"color": color}, prefix+row))
	}

	logRows := []*vdom.Node{
		ink.Text(vdom.Props{"bold": true, "color": "yellow"}, "Live Logs (only 1-2 lines update per frame):"),
	}
	for _, line := range snapshot.logLines {
		logRows = append(logRows, ink.Text(vdom.Props{"color": "green"}, line))
	}

	selected := ""
	if snapshot.selectedIndex >= 0 && snapshot.selectedIndex < serviceCount {
		selected = incrementalRows[snapshot.selectedIndex]
	}

	return ink.Box(vdom.Props{"flexDirection": "column"},
		ink.Box(vdom.Props{"borderStyle": "round", "borderColor": "cyan", "paddingX": 2, "paddingY": 1},
			ink.Box(vdom.Props{"flexDirection": "column"},
				ink.Text(vdom.Props{"bold": true, "color": "cyan"}, "Incremental Rendering Demo - incrementalRendering=true"),
				ink.Text(vdom.Props{"dimColor": true}, fmt.Sprintf("Use ↑/↓ arrows to navigate • Press q to quit • FPS: %d", snapshot.fps)),
				ink.Text(
					"Time: ", ink.Text(vdom.Props{"color": "green"}, snapshot.timestamp),
					" • Updates: ", ink.Text(vdom.Props{"color": "yellow"}, snapshot.counter),
					" • Random: ", ink.Text(vdom.Props{"color": "cyan"}, snapshot.randomValue),
				),
				ink.Text("Progress 1: ", ink.Text(vdom.Props{"color": "green"}, progressBar(snapshot.progress1)), fmt.Sprintf(" %d%%", snapshot.progress1)),
				ink.Text("Progress 2: ", ink.Text(vdom.Props{"color": "yellow"}, progressBar(snapshot.progress2)), fmt.Sprintf(" %d%%", snapshot.progress2)),
				ink.Text("Progress 3: ", ink.Text(vdom.Props{"color": "red"}, progressBar(snapshot.progress3)), fmt.Sprintf(" %d%%", snapshot.progress3)),
			),
		),
		ink.Box(vdom.Props{"borderStyle": "single", "borderColor": "yellow", "paddingX": 2, "paddingY": 1, "marginTop": 1},
			ink.Box(vdom.Props{"flexDirection": "column"}, logRows...),
		),
		ink.Box(vdom.Props{"borderStyle": "single", "borderColor": "gray", "paddingX": 2, "paddingY": 1, "marginTop": 1, "flexDirection": "column"}, serviceRows...),
		ink.Box(vdom.Props{"borderStyle": "round", "borderColor": "magenta", "paddingX": 2, "marginTop": 1},
			ink.Text("Selected: ", ink.Text(vdom.Props{"bold": true, "color": "magenta"}, selected)),
		),
	)
}

func incrementalSnapshotFor(logLineCount int, serviceCount int) incrementalSnapshot {
	incrementalStateStore.Lock()
	defer incrementalStateStore.Unlock()

	ensureIncrementalStateLocked(logLineCount)
	if incrementalStateStore.state.selectedIndex >= serviceCount {
		incrementalStateStore.state.selectedIndex = serviceCount - 1
	}
	if incrementalStateStore.state.selectedIndex < 0 {
		incrementalStateStore.state.selectedIndex = 0
	}

	logLines := append([]string(nil), incrementalStateStore.state.logLines...)
	return incrementalSnapshot{
		selectedIndex: incrementalStateStore.state.selectedIndex,
		timestamp:     incrementalStateStore.state.timestamp,
		counter:       incrementalStateStore.state.counter,
		fps:           incrementalStateStore.state.fps,
		progress1:     incrementalStateStore.state.progress1,
		progress2:     incrementalStateStore.state.progress2,
		progress3:     incrementalStateStore.state.progress3,
		randomValue:   incrementalStateStore.state.randomValue,
		logLines:      logLines,
	}
}

func updateIncrementalState(logLineCount int) {
	incrementalStateStore.Lock()
	defer incrementalStateStore.Unlock()

	ensureIncrementalStateLocked(logLineCount)
	state := &incrementalStateStore.state
	state.progress1 = (state.progress1 + 1) % 101
	state.progress2 = (state.progress2 + 2) % 101
	state.progress3 = (state.progress3 + 3) % 101
	state.randomValue = state.random.Intn(1000)

	if len(state.logLines) > 0 {
		updateIndex := state.random.Intn(len(state.logLines))
		state.logLines[updateIndex] = generateIncrementalLogLine(updateIndex, state.random.Intn(1000), state.random)
	}

	state.frameCount++
	now := time.Now()
	if now.Sub(state.lastFPSAt) >= time.Second {
		state.timestamp = now.Format("3:04:05 PM")
		state.counter++
		state.fps = state.frameCount
		state.frameCount = 0
		state.lastFPSAt = now
	}
}

func moveIncrementalSelection(delta int, serviceCount int) {
	incrementalStateStore.Lock()
	defer incrementalStateStore.Unlock()

	ensureIncrementalStateLocked(3)
	if serviceCount <= 0 {
		incrementalStateStore.state.selectedIndex = 0
		return
	}

	next := incrementalStateStore.state.selectedIndex + delta
	if next < 0 {
		next = serviceCount - 1
	}
	if next >= serviceCount {
		next = 0
	}
	incrementalStateStore.state.selectedIndex = next
}

func ensureIncrementalStateLocked(logLineCount int) {
	state := &incrementalStateStore.state
	if state.random == nil {
		state.random = rand.New(rand.NewSource(time.Now().UnixNano()))
		state.timestamp = time.Now().Format("3:04:05 PM")
		state.lastFPSAt = time.Now()
	}

	for len(state.logLines) < logLineCount {
		index := len(state.logLines)
		state.logLines = append(state.logLines, generateIncrementalLogLine(index, 0, state.random))
	}
	if len(state.logLines) > logLineCount {
		state.logLines = state.logLines[:logLineCount]
	}
}

func resetIncrementalState() {
	incrementalStateStore.Lock()
	defer incrementalStateStore.Unlock()

	incrementalStateStore.state = incrementalState{}
}

func generateIncrementalLogLine(index int, value int, random *rand.Rand) string {
	actions := []string{"PROCESSING", "COMPLETED", "UPDATING", "SYNCING", "VALIDATING", "EXECUTING"}
	return fmt.Sprintf("[%s] Worker-%d %s: Batch=%d Throughput=%.0freq/s Memory=%.1fMB CPU=%.1f%%",
		time.Now().Format("3:04:05 PM"),
		index,
		actions[random.Intn(len(actions))],
		value,
		random.Float64()*1000,
		random.Float64()*512,
		random.Float64()*100,
	)
}

func main() {
	resetIncrementalState()

	if !terminal.StdinIsTerminal() {
		app := ink.NewApp(IncrementalRenderingDemo)
		fmt.Println(app.RenderOnce())
		return
	}

	instance, err := ink.RenderWithOptions(IncrementalRenderingDemo, ink.RenderOptions{
		IncrementalRendering: true,
	})
	if err != nil {
		panic(err)
	}

	inputDone := make(chan error, 1)
	go func() {
		inputDone <- ttyinput.Run(os.Stdin, instance.HandleInput, func(input string) bool {
			return strings.Contains(input, "q") || strings.Contains(input, "\x03")
		})
	}()

	exitDone := make(chan error, 1)
	go func() {
		exitDone <- instance.WaitUntilExit()
	}()

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			updateIncrementalState(3)
			if err := instance.Rerender(IncrementalRenderingDemo); err != nil {
				fmt.Println(err)
				return
			}
		case err := <-inputDone:
			if err != nil {
				fmt.Println(err)
			}
			return
		case err := <-exitDone:
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
