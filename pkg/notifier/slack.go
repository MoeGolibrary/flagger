/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package notifier

import (
	"errors"
	"fmt"
	"net/url"
)

type Style string

const (
	StyleDefault Style = ""
	StylePrimary Style = "primary"
	StyleDanger  Style = "danger"
)

// Slack holds the hook URL
type Slack struct {
	URL      string
	Token    string
	ProxyURL string
	Username string
	Channel  string
}

// SlackPayload holds the channel and attachments
type SlackPayload struct {
	Channel     string            `json:"channel"`
	Username    string            `json:"username"`
	IconUrl     string            `json:"icon_url"`
	IconEmoji   string            `json:"icon_emoji"`
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment holds the markdown message body
type SlackAttachment struct {
	Color      string        `json:"color"`
	AuthorName string        `json:"author_name"`
	Text       string        `json:"text"`
	MrkdwnIn   []string      `json:"mrkdwn_in"`
	Fields     []SlackField  `json:"fields"`
	Actions    []SlackAction `json:"actions"` // 新增 Actions 字段
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type SlackAction struct {
	Type     string                   `json:"type"`
	Text     *TextBlockObject         `json:"text"`
	ActionID string                   `json:"action_id,omitempty"`
	URL      string                   `json:"url,omitempty"`
	Value    string                   `json:"value,omitempty"`
	Style    Style                    `json:"style,omitempty"`
	Confirm  *ConfirmationBlockObject `json:"confirm,omitempty"`
}

type ConfirmationBlockObject struct {
	Title   *TextBlockObject `json:"title"`
	Text    *TextBlockObject `json:"text"`
	Confirm *TextBlockObject `json:"confirm"`
	Deny    *TextBlockObject `json:"deny,omitempty"`
	Style   Style            `json:"style,omitempty"`
}

type TextBlockObject struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Emoji    *bool  `json:"emoji"`
	Verbatim bool   `json:"verbatim,omitempty"`
}

// NewSlack validates the Slack URL and returns a Slack object
func NewSlack(address, token, proxyURL, username, channel string) (*Slack, error) {
	_, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, fmt.Errorf("invalid Slack hook URL %s", address)
	}

	if username == "" {
		return nil, errors.New("empty Slack username")
	}

	if channel == "" {
		return nil, errors.New("empty Slack channel")
	}

	return &Slack{
		Channel:  channel,
		URL:      address,
		Token:    token,
		ProxyURL: proxyURL,
		Username: username,
	}, nil
}

// Post Slack message
func (s *Slack) Post(workload string, namespace string, message string, fields []Field, severity string) error {
	payload := SlackPayload{
		Channel:   s.Channel,
		Username:  s.Username,
		IconEmoji: ":rocket:",
	}

	color := "good"
	if severity == "warn" {
		color = "warning"
	} else if severity == "error" {
		color = "danger"
	}

	sfields := make([]SlackField, 0)
	for _, f := range fields {
		if f.Type != "link" {
			sfields = append(sfields, SlackField{f.Name, f.Value, false})
		}
	}
	actions := make([]SlackAction, 0)
	for _, f := range fields {
		if f.Type == "link" {
			actions = append(actions, SlackAction{
				Type: "button",
				Text: &TextBlockObject{
					Text: f.Name,
				},
				URL: f.Value,
			})
		}
	}

	// TODO 优化
	actions = append(actions, SlackAction{
		Type:     "button",
		ActionID: "skip_canary",
		Value:    "skip_canary",
		Text: &TextBlockObject{
			Text: "Skip Canary (Test)",
		},
		Style: "danger",
		Confirm: &ConfirmationBlockObject{
			Title: &TextBlockObject{
				Text: "Are you sure?",
			},
			Text: &TextBlockObject{
				Text: "This will skip the canary test.",
			},
			Confirm: &TextBlockObject{
				Text: "Yes",
			},
			Deny: &TextBlockObject{
				Text: "No",
			},
		},
	}, SlackAction{
		Type:     "button",
		ActionID: "rollback_canary",
		Value:    "rollback_canary",
		Text: &TextBlockObject{
			Text: "Rollback (Test)",
		},
		Style: "danger",
		Confirm: &ConfirmationBlockObject{
			Title: &TextBlockObject{
				Text: "Are you sure?",
			},
			Text: &TextBlockObject{
				Text: "This will rollback the canary test.",
			},
			Confirm: &TextBlockObject{
				Text: "Yes",
			},
			Deny: &TextBlockObject{
				Text: "No",
			},
		},
	})

	a := SlackAttachment{
		Color:    color,
		Text:     fmt.Sprintf("Workload: %s | Namespace: %s \n", workload, namespace) + message,
		MrkdwnIn: []string{"text"},
		Fields:   sfields,
		Actions:  actions, // 填充 Actions 字段
	}

	payload.Attachments = []SlackAttachment{a}

	err := postMessage(s.URL, s.Token, s.ProxyURL, payload)
	if err != nil {
		return fmt.Errorf("postMessage failed: %w", err)
	}
	return nil
}
