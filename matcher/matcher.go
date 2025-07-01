package matcher

import (
	"log"
	"regexp"

	"mclog2event/config"
	"mclog2event/types"
)

// Matcher holds the compiled regex patterns for event matching.
type Matcher struct {
	eventPatterns []*regexp.Regexp
	events        []config.Event
}

// NewMatcher creates a new Matcher with compiled regex patterns.
func NewMatcher(events []config.Event) (*Matcher, error) {
	var patterns []*regexp.Regexp
	for _, event := range events {
		re, err := regexp.Compile(event.Pattern)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, re)
	}
	return &Matcher{eventPatterns: patterns, events: events}, nil
}

// Match processes a log line and returns a map of matched data or nil.
func (m *Matcher) Match(line string) *types.EventPayload {
	for i, re := range m.eventPatterns {
		match := re.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		data := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				data[name] = match[i]
			}
		}

		log.Printf("New match: %s", m.events[i].EventType)
		return &types.EventPayload{
			EventType: m.events[i].EventType,
			EventData: data,
		}
	}
	return nil
}
