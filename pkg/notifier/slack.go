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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"net/url"
	"strings"
)

// Slack action constants
const (
	SkipCanaryAction     = "skip_canary"
	RollbackCanaryAction = "rollback_canary"
	PauseAtWeightAction  = "pause_at_weight"
	SetWeightAction      = "set_weight"
	ResumeCanaryAction   = "resume_canary"
	GetStatusAction      = "get_status"
	WeightInputBlockID   = "weight_input_block"
	WeightInputElementID = "weight_input"
	ActionsBlockID       = "actions"

	// Button labels
	SkipCanaryLabel    = "Skip Canary"
	RollbackLabel      = "Rollback"
	PauseAtWeightLabel = "Pause at Weight"
	SetWeightLabel     = "Set Weight"
	ResumeLabel        = "Resume"
	GetStatusLabel     = "Get Status"
	WeightLabel        = "Weight"
	WeightPlaceholder  = "Enter weight"
	WeightHint         = "Enter weight(0.0-100.0)"

	// Confirmation dialog texts
	ConfirmTitle        = "Are you sure?"
	SkipConfirmText     = "This will skip the canary test.\n *Workload:* %s \n *Namespace:* %s \n"
	RollbackConfirmText = "This will rollback the canary test.\n *Workload:* %s \n *Namespace:* %s \n"
	YesLabel            = "Yes"
	NoLabel             = "No"
)

// Slack holds the hook URL
type Slack struct {
	URL      string
	Token    string
	ProxyURL string
	Username string
	Channel  string
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
func (s *Slack) Post(workload string, namespace string, message string, fields []Field, severity string, canaryId string) error {

	// Create blocks
	blocks := make([]slack.Block, 0)

	// add header block
	headerBlock := slack.NewHeaderBlock(
		slack.NewTextBlockObject("plain_text", fmt.Sprintf("[%s] %s | %s", strings.ToUpper(severity), workload, namespace), false, false),
	)
	blocks = append(blocks, headerBlock)

	// Add fields as context blocks
	sfields := make([]*slack.TextBlockObject, 0)
	// add fields workload and namespace
	sfields = append(sfields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Workload:* %s", workload), false, false))
	sfields = append(sfields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Namespace:* %s", namespace), false, false))

	if len(fields) > 0 {
		for _, f := range fields {
			if f.Type != "link" {
				sfields = append(sfields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s:* %s", f.Name, f.Value), false, false))
			}
		}
	}

	// Add section block for message
	sectionBlock := slack.NewSectionBlock(
		slack.NewTextBlockObject("mrkdwn", message, false, false),
		sfields,
		nil,
	)
	blocks = append(blocks, sectionBlock)

	// Add actions as action block
	if len(fields) > 0 {
		var elements []slack.BlockElement
		for _, f := range fields {
			if f.Type == "link" {
				elements = append(elements, slack.NewButtonBlockElement(
					f.Name,
					f.Name,
					slack.NewTextBlockObject("plain_text", f.Name, false, false),
				).WithURL(f.Value))
			}
		}

		// 如果message以New revision detected开头
		if canaryId != "" && strings.HasPrefix(message, "New revision detected") {

			// Add additional buttons
			elements = append(elements, slack.NewButtonBlockElement(
				SkipCanaryAction,
				canaryId,
				slack.NewTextBlockObject("plain_text", SkipCanaryLabel, false, false),
			).WithStyle(slack.StyleDanger).WithConfirm(
				slack.NewConfirmationBlockObject(
					slack.NewTextBlockObject("plain_text", ConfirmTitle, false, false),
					slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(SkipConfirmText,
						workload, namespace), false, false),
					slack.NewTextBlockObject("plain_text", YesLabel, false, false),
					slack.NewTextBlockObject("plain_text", NoLabel, false, false),
				),
			))

			elements = append(elements, slack.NewButtonBlockElement(
				RollbackCanaryAction,
				canaryId,
				slack.NewTextBlockObject("plain_text", RollbackLabel, false, false),
			).WithStyle(slack.StyleDanger).WithConfirm(
				slack.NewConfirmationBlockObject(
					slack.NewTextBlockObject("plain_text", ConfirmTitle, false, false),
					slack.NewTextBlockObject("mrkdwn", fmt.Sprintf(RollbackConfirmText,
						workload, namespace), false, false),
					slack.NewTextBlockObject("plain_text", YesLabel, false, false),
					slack.NewTextBlockObject("plain_text", NoLabel, false, false),
				),
			))

			//添加人工介入控制按钮和一个权重输入框

			// 创建输入框元素
			weightInput := slack.NewPlainTextInputBlockElement(
				&slack.TextBlockObject{
					Type: "plain_text",
					Text: WeightLabel,
				},
				WeightInputElementID,
			)
			weightInput.Placeholder = &slack.TextBlockObject{
				Type: "plain_text",
				Text: WeightPlaceholder,
			}

			// 创建输入框区块
			inputBlock := slack.NewInputBlock(
				WeightInputBlockID,
				&slack.TextBlockObject{
					Type: "plain_text",
					Text: WeightLabel,
				},
				&slack.TextBlockObject{
					Type: "plain_text",
					Text: WeightHint,
				},
				weightInput,
			)
			blocks = append(blocks, inputBlock)

			// Pause at Weight button
			elements = append(elements, slack.NewButtonBlockElement(
				PauseAtWeightAction,
				canaryId,
				slack.NewTextBlockObject("plain_text", PauseAtWeightLabel, false, false),
			).WithStyle(slack.StylePrimary))

			// Set Weight button
			elements = append(elements, slack.NewButtonBlockElement(
				SetWeightAction,
				canaryId,
				slack.NewTextBlockObject("plain_text", SetWeightLabel, false, false),
			).WithStyle(slack.StylePrimary))

			// Resume button
			elements = append(elements, slack.NewButtonBlockElement(
				ResumeCanaryAction,
				canaryId,
				slack.NewTextBlockObject("plain_text", ResumeLabel, false, false),
			).WithStyle(slack.StylePrimary))

			//// Get Status button
			//elements = append(elements, slack.NewButtonBlockElement(
			//	GetStatusAction,
			//	canaryId,
			//	slack.NewTextBlockObject("plain_text", GetStatusLabel, false, false),
			//).WithStyle(slack.StylePrimary))
		}

		if len(elements) > 0 {
			actionsBlock := slack.NewActionBlock(
				ActionsBlockID,
				elements...,
			)
			blocks = append(blocks, actionsBlock)
		}
	}

	msg := slack.WebhookMessage{
		Blocks: &slack.Blocks{
			BlockSet: blocks,
		},
	}

	err := slack.PostWebhook(s.URL, &msg)

	if err != nil {
		// 输出msg json
		b, _ := json.Marshal(msg)
		fmt.Printf("Slack WebhookMessage: %s \n", string(b))
		return fmt.Errorf("postMessage failed: %s, err: %w", string(b), err)
	}
	return nil
}
