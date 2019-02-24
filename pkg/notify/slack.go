package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Slack implements a Notifier for Slack notifications.
type Slack struct {
	client *http.Client
}

// NewSlack returns a new Slack notification handler.
func NewSlack(c *http.Client) *Slack {
	return &Slack{
		client: c,
	}
}

// slackReq is the request for sending a slack notification.
type slackReq struct {
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	IconURL     string            `json:"icon_url,omitempty"`
	LinkNames   bool              `json:"link_names,omitempty"`
	Blocks      []block           `json:"blocks,omitempty"`
	Text        string			  `json:"text,omitempty"`
}

// block is used to display a richly-formatted message block.
type block struct {
	Type      string                `json:"type,omitempty"`
	Text      *field                 `json:"text,omitempty"`
	Fields    []*field               `json:"fields,omitempty"`
}

type field struct {
	Type      string                `json:"type"`
	Text      string                `json:"text"`
}

// Notify implements the Notifier interface for Slack Notifications
func (n *Slack) Notify(f *entity.Flag, b notify, i itemType) error {
	var err error

	slackReq := buildSlackRequest(f, b, i)

	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(slackReq); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", config.Config.SlackUrl, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("User-Agent", userAgentHeader)

	resp, err := n.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		b, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(b))
	}
	return nil
}

// titleCase returns a Title case string
func titleCase(i interface{}) string {
	if str, ok := i.(notify); ok {
		return strings.Title(strings.ToLower(string(str)))
	}
	if str, ok := i.(itemType); ok {
		return strings.Title(strings.ToLower(string(str)))
	}
	return strings.Title(strings.ToLower(""))
}

// buildSlackRequest builds the correct slack req ready for sending, based on the event in question
func buildSlackRequest(f *entity.Flag, b notify, i itemType) *slackReq {
	var blocks []block
	header := fmt.Sprintf("Flag #%d (%s)", f.ID, f.Description)

	if b == TOGGLED {
		var tmpl string
		if f.Enabled {
			tmpl = "%s has been enabled at %s"
		} else {
			tmpl = "%s has been disabled at %s"
		}
		field := newField(fmt.Sprintf(tmpl, header, time.Now().Format(time.RFC850)))
		blocks = append(blocks, block{Type: "section", Text: field})
	} else {
		titleStr := fmt.Sprintf("*%s*\n %s was %s", header, titleCase(i), titleCase(b))

		blocks = append(blocks, section(newField(titleStr)))

		variants := variants(f)
		if len(variants) > 0 {
			str := fmt.Sprintf("*Current variants*\n%s", strings.Join(variants, "\n"))
			blocks = append(blocks, section(newField(str)))
		}

		segments := segments(f)
		if len(segments) > 0 {
			blocks = append(blocks, section(segments))
		}
	}

	fallbackText := fmt.Sprintf("%s was updated", header)

	slackReq := &slackReq{
		Channel:  config.Config.SlackChannel,
		Username: "Flagr",
		Blocks:   blocks,
		Text:     fallbackText,
	}
	return slackReq
}

// variants builds an array of strings containing all variants for a flag
func variants(f *entity.Flag) []string {
	var variants []string
	for _, variant := range f.Variants {
		variants = append(variants, variant.Key)
	}
	return variants
}

// segments builds up slack fields for each segment
func segments(f *entity.Flag) []*field {
	var segments []*field
	for _, segment := range f.Segments {
		segmentsStr := fmt.Sprintf("*%s* (Current rollout %d%%)", segment.Description, segment.RolloutPercent)
		for _, distribution := range segment.Distributions {
			if segment.ID == distribution.SegmentID {
				segmentsStr = segmentsStr + fmt.Sprintf("\n%s : %d%%", distribution.VariantKey, distribution.Percent)
			}
		}
		segments = append(segments, newField(segmentsStr))
	}
	return segments
}

// newField returns a new field
func newField(input string) *field {
	return &field{Type: "mrkdwn", Text: input}
}

// section returns a new section, with either an array of fields or a singular field
func section(f interface{}) block {
	if field, ok := f.(*field); ok {
		return block{Type: "section", Text: field}
	}
	if fields, ok := f.([]*field); ok {
		return block{Type: "section", Fields: fields}
	}
	return block{Type: "section"}
}
