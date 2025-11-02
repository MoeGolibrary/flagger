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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
)

func TestSlack_Post(t *testing.T) {
	fields := []Field{
		{Name: "name1", Value: "value1", Type: "text"},
		{Name: "name2", Value: "value2", Type: "text"},
		{Name: "Link1", Value: "http://baidu.com", Type: "link"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		fmt.Printf("received payload: %s", string(b))
		var payload slack.WebhookMessage
		err = json.Unmarshal(b, &payload)
		require.NoError(t, err)

		// 验证 Blocks 结构
		require.NotNil(t, payload.Blocks)
		blocks := payload.Blocks.BlockSet

		// 应该有 header, section, input 和 actions 四个 blocks
		require.Equal(t, 4, len(blocks))

		// 检查 header block
		headerBlock, ok := blocks[0].(*slack.HeaderBlock)
		require.True(t, ok)
		require.Equal(t, "[ERROR] podinfo | test", headerBlock.Text.Text)

		// 检查 section block
		sectionBlock, ok := blocks[1].(*slack.SectionBlock)
		require.True(t, ok)
		require.Equal(t, "New revision detected !test", sectionBlock.Text.Text)

		// 检查 actions block
		actionsBlock, ok := blocks[3].(*slack.ActionBlock)
		require.True(t, ok)

		// 应该有 6 个按钮: Link1, Skip Canary, Rollback, Pause at Weight, Resume, Set Weight
		require.Equal(t, 6, len(actionsBlock.Elements.ElementSet))

		// 检查按钮名称
		buttons := actionsBlock.Elements.ElementSet
		require.Equal(t, "Link1", buttons[0].(*slack.ButtonBlockElement).Text.Text)
		require.Equal(t, "Skip Canary", buttons[1].(*slack.ButtonBlockElement).Text.Text)
		require.Equal(t, "Rollback", buttons[2].(*slack.ButtonBlockElement).Text.Text)
		require.Equal(t, "Pause at Weight", buttons[3].(*slack.ButtonBlockElement).Text.Text)
		require.Equal(t, "Set Weight", buttons[4].(*slack.ButtonBlockElement).Text.Text)
		require.Equal(t, "Resume", buttons[5].(*slack.ButtonBlockElement).Text.Text)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	slack, err := NewSlack(ts.URL, "", "", "test", "test")
	require.NoError(t, err)

	err = slack.Post("podinfo", "test", "New revision detected !test", fields, "error", "canary-id")
	require.NoError(t, err)
}

func TestSlack_Post_WithoutManualControlButtons(t *testing.T) {
	fields := []Field{
		{Name: "name1", Value: "value1", Type: "text"},
		{Name: "Link1", Value: "http://baidu.com", Type: "link"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload slack.WebhookMessage
		err = json.Unmarshal(b, &payload)
		require.NoError(t, err)

		// 验证 Blocks 结构
		require.NotNil(t, payload.Blocks)
		blocks := payload.Blocks.BlockSet

		// 应该有 header, section 和 actions 三个 blocks
		require.Equal(t, 3, len(blocks))

		// 检查 actions block 中的按钮
		actionsBlock, ok := blocks[2].(*slack.ActionBlock)
		require.True(t, ok)

		// 当消息不是"New revision detected"开头时，应该有 1 个按钮: Link1
		require.Equal(t, 1, len(actionsBlock.Elements.ElementSet))

		// 检查按钮名称
		buttons := actionsBlock.Elements.ElementSet
		require.Equal(t, "Link1", buttons[0].(*slack.ButtonBlockElement).Text.Text)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	slack, err := NewSlack(ts.URL, "", "", "test", "test")
	require.NoError(t, err)

	// 使用不以"New revision detected"开头的消息，应该只显示人工介入按钮
	err = slack.Post("podinfo", "test", "Test message without canary controls", fields, "info", "canary-id")
	require.NoError(t, err)
}

func TestSlack_Post_WithoutCanaryId(t *testing.T) {
	fields := []Field{
		{Name: "name1", Value: "value1", Type: "text"},
		{Name: "Link1", Value: "http://baidu.com", Type: "link"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var payload slack.WebhookMessage
		err = json.Unmarshal(b, &payload)
		require.NoError(t, err)

		// 验证 Blocks 结构
		require.NotNil(t, payload.Blocks)
		blocks := payload.Blocks.BlockSet

		// 应该有 header, section 和 actions 三个 blocks
		require.Equal(t, 3, len(blocks))

		// 检查 actions block 中的按钮
		actionsBlock, ok := blocks[2].(*slack.ActionBlock)
		require.True(t, ok)

		// 当没有 canary-id 时，应该只有 1 个按钮: Link1
		require.Equal(t, 1, len(actionsBlock.Elements.ElementSet))

		// 检查按钮名称
		buttons := actionsBlock.Elements.ElementSet
		require.Equal(t, "Link1", buttons[0].(*slack.ButtonBlockElement).Text.Text)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	slack, err := NewSlack(ts.URL, "", "", "test", "test")
	require.NoError(t, err)

	// 使用不包含 canary-id 的消息，不应该显示人工介入按钮
	err = slack.Post("podinfo", "test", "Test message without canary controls", fields, "info", "")
	require.NoError(t, err)
}
