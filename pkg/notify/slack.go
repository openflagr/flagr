package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/checkr/flagr/pkg/util"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/entity"
)

// Slack implements a Notifier for Slack notifications.
type Slack struct {
	client *Client
}

// NewSlack returns a new Slack notification handler.
func NewSlack(c *Client) *Slack {
	return &Slack{
		client: c,
	}
}

// slackReq is the request for sending a slack notification.
type slackReq struct {
	Channel   string  `json:"channel,omitempty"`
	Username  string  `json:"username,omitempty"`
	IconEmoji string  `json:"icon_emoji,omitempty"`
	IconURL   string  `json:"icon_url,omitempty"`
	LinkNames bool    `json:"link_names,omitempty"`
	Blocks    []block `json:"blocks,omitempty"`
	Text      string  `json:"text,omitempty"`
}

// block is used to display a richly-formatted message block.
type block struct {
	Type   string   `json:"type,omitempty"`
	Text   *field   `json:"text,omitempty"`
	Fields []*field `json:"fields,omitempty"`
}

type field struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Notify implements the Notifier interface for Slack Notifications
func (n *Slack) Notify(f *entity.Flag, b itemAction, i itemType, s subject) error {
	var err error

	slackReq := buildSlackRequest(f, b, i, s)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(slackReq); err != nil {
		return err
	}

	_, err = n.client.Post(config.Config.NotifySlackURL, bytes.NewReader(buf.Bytes()))

	if err != nil {
		return err
	}

	return nil
}

// titleCase changes the case
func titleCase(i interface{}) string {
	if str, ok := i.(itemAction); ok {
		return util.TitleCase(string(str))
	}
	if str, ok := i.(itemType); ok {
		return util.TitleCase(string(str))
	}
	return util.TitleCase(i)
}

// buildSlackRequest builds the correct slack req ready for sending, based on the event in question
func buildSlackRequest(f *entity.Flag, b itemAction, i itemType, s subject) *slackReq {
	var blocks []block
	header := fmt.Sprintf("Flag #%d (%s)", f.ID, f.Description)
	if s != "" {
		header = fmt.Sprintf("Flag #%d (%s) by %s", f.ID, f.Description, s)
	}

	if b == TOGGLED {
		blocks = buildToggledReq(header, blocks, f)
	} else {
		blocks = buildSectionedReq(header, i, b, blocks, f)
	}

	slackReq := &slackReq{
		Channel:  config.Config.NotifySlackChannel,
		Username: "Flagr",
		Blocks:   blocks,
		Text:     fmt.Sprintf("%s was updated", header),
	}
	return slackReq
}

// buildSectionedReq builds the sections for a crud update
func buildSectionedReq(header string, i itemType, b itemAction, blocks []block, f *entity.Flag) []block {
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
	return blocks
}

// buildToggledReq builds the text block for a toggled request
func buildToggledReq(header string, blocks []block, f *entity.Flag) []block {
	var tmpl string
	if f.Enabled {
		tmpl = "%s has been enabled at %s"
	} else {
		tmpl = "%s has been disabled at %s"
	}
	field := newField(fmt.Sprintf(tmpl, header, time.Now().Format(time.RFC850)))
	return append(blocks, block{Type: "section", Text: field})
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
