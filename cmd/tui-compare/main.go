package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dh-kam/ink-go/internal/tuitest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	leftPath := flag.String("left", "", "left transcript JSON path")
	rightPath := flag.String("right", "", "right transcript JSON path")
	mode := flag.String("mode", "screen", "comparison mode: screen, plain, raw, raw-escaped")
	candidatePath := flag.String("candidate", "", "optional path to write right-side candidate content")
	flag.Parse()

	if *leftPath == "" {
		return errors.New("-left is required")
	}
	if *rightPath == "" {
		return errors.New("-right is required")
	}

	left, err := readTranscript(*leftPath)
	if err != nil {
		return err
	}
	right, err := readTranscript(*rightPath)
	if err != nil {
		return err
	}

	if len(left.Frames) != len(right.Frames) {
		return fmt.Errorf("frame count mismatch: left=%d right=%d", len(left.Frames), len(right.Frames))
	}
	if left.Scenario != right.Scenario {
		return fmt.Errorf("scenario mismatch: left=%q right=%q", left.Scenario, right.Scenario)
	}
	if left.App != right.App {
		return fmt.Errorf("app mismatch: left=%q right=%q", left.App, right.App)
	}
	if left.Viewport != right.Viewport {
		return fmt.Errorf("viewport mismatch: left=%dx%d right=%dx%d",
			left.Viewport.Width, left.Viewport.Height, right.Viewport.Width, right.Viewport.Height)
	}

	for index := range left.Frames {
		leftFrame := left.Frames[index]
		rightFrame := right.Frames[index]
		if leftFrame.Step != rightFrame.Step {
			return fmt.Errorf("frame %d step mismatch: left=%q right=%q", index, leftFrame.Step, rightFrame.Step)
		}

		leftValue, err := frameValue(leftFrame, *mode)
		if err != nil {
			return err
		}
		rightValue, err := frameValue(rightFrame, *mode)
		if err != nil {
			return err
		}
		if leftValue != rightValue {
			if *candidatePath != "" {
				if writeErr := os.WriteFile(*candidatePath, []byte(rightValue), 0o644); writeErr != nil {
					return fmt.Errorf("write candidate %q: %w", *candidatePath, writeErr)
				}
			}
			return fmt.Errorf("frame %d step %q mismatch in %s at line %d", index, leftFrame.Step, *mode, firstDifferingLine(leftValue, rightValue))
		}
	}

	fmt.Printf("transcripts match: scenario=%q mode=%s frames=%d\n", left.Scenario, normalizeMode(*mode), len(left.Frames))
	return nil
}

func readTranscript(path string) (*tuitest.Transcript, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read transcript %q: %w", path, err)
	}

	var transcript tuitest.Transcript
	if err := json.Unmarshal(data, &transcript); err != nil {
		return nil, fmt.Errorf("decode transcript %q: %w", path, err)
	}
	if transcript.SchemaVersion != tuitest.TranscriptSchemaVersion {
		return nil, fmt.Errorf("unsupported transcript schemaVersion %q in %q", transcript.SchemaVersion, path)
	}
	return &transcript, nil
}

func frameValue(frame tuitest.TranscriptFrame, mode string) (string, error) {
	switch normalizeMode(mode) {
	case "screen":
		if frame.ScreenPlain != "" {
			return frame.ScreenPlain, nil
		}
		return frame.Plain, nil
	case "plain":
		return frame.Plain, nil
	case "raw":
		return frame.Raw, nil
	case "raw-escaped", "rawescaped":
		return frame.RawEscaped, nil
	default:
		return "", fmt.Errorf("unsupported comparison mode %q", mode)
	}
}

func normalizeMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "screen", "screen-plain", "screenplain":
		return "screen"
	case "plain":
		return "plain"
	case "raw":
		return "raw"
	case "raw-escaped", "rawescaped":
		return "raw-escaped"
	default:
		return mode
	}
}

func firstDifferingLine(left string, right string) int {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	lineCount := len(leftLines)
	if len(rightLines) > lineCount {
		lineCount = len(rightLines)
	}
	for index := 0; index < lineCount; index++ {
		leftLine := ""
		if index < len(leftLines) {
			leftLine = leftLines[index]
		}
		rightLine := ""
		if index < len(rightLines) {
			rightLine = rightLines[index]
		}
		if leftLine != rightLine {
			return index + 1
		}
	}
	return 0
}
