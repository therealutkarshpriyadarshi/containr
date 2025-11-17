package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/therealutkarshpriyadarshi/containr/pkg/events"
)

var eventsCmd = &cobra.Command{
	Use:   "events [OPTIONS]",
	Short: "Get real-time events from containers",
	Long:  `Stream or display container lifecycle events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		since, _ := cmd.Flags().GetString("since")
		until, _ := cmd.Flags().GetString("until")
		filter, _ := cmd.Flags().GetStringArray("filter")
		format, _ := cmd.Flags().GetString("format")

		// Create event manager
		em, err := events.NewEventManager(stateDir)
		if err != nil {
			return fmt.Errorf("failed to create event manager: %w", err)
		}

		// Build filters
		filters := events.EventFilters{}

		// Parse time filters
		if since != "" {
			t, err := parseTimeFilter(since)
			if err != nil {
				return fmt.Errorf("invalid --since value: %w", err)
			}
			filters.Since = &t
		}

		if until != "" {
			t, err := parseTimeFilter(until)
			if err != nil {
				return fmt.Errorf("invalid --until value: %w", err)
			}
			filters.Until = &t
		}

		// Parse other filters
		for _, f := range filter {
			// TODO: Parse filter format (e.g., "container=abc", "type=start")
			_ = f
		}

		// Get events
		eventList := em.GetEvents(filters)

		// Display events
		if format == "json" {
			data, err := json.MarshalIndent(eventList, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
		} else {
			// Human-readable format
			for _, event := range eventList {
				fmt.Println(events.FormatEvent(event))
			}
		}

		return nil
	},
}

func init() {
	eventsCmd.Flags().String("since", "", "Show events created since timestamp (e.g., 2023-01-01T00:00:00Z)")
	eventsCmd.Flags().String("until", "", "Show events created before timestamp")
	eventsCmd.Flags().StringArray("filter", nil, "Filter output based on conditions (e.g., container=<name>)")
	eventsCmd.Flags().String("format", "", "Format output (json)")
}

// parseTimeFilter parses a time filter string
func parseTimeFilter(s string) (time.Time, error) {
	// Try RFC3339 format first
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	// Try other common formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
