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
		var payload SlackPayload
		err = json.Unmarshal(b, &payload)
		require.NoError(t, err)

		require.Equal(t, "podinfo.test", payload.Attachments[0].AuthorName)
		require.Equal(t, 2, len(payload.Attachments[0].Fields))  // 只有两个文本字段
		require.Equal(t, 1, len(payload.Attachments[0].Actions)) // 有一个链接字段

		// 检查 Fields 字段
		require.Equal(t, "name1", payload.Attachments[0].Fields[0].Title)
		require.Equal(t, "value1", payload.Attachments[0].Fields[0].Value)
		require.Equal(t, "name2", payload.Attachments[0].Fields[1].Title)
		require.Equal(t, "value2", payload.Attachments[0].Fields[1].Value)

		// 检查 Actions 字段
		require.Equal(t, "button", payload.Attachments[0].Actions[0].Type)
		require.Equal(t, "Link1", payload.Attachments[0].Actions[0].Text)
		require.Equal(t, "http://baidu.com", payload.Attachments[0].Actions[0].URL)
	}))
	defer ts.Close()

	slack, err := NewSlack(ts.URL, "", "", "test", "test")
	require.NoError(t, err)

	err = slack.Post("podinfo", "test", "test", fields, "error")
	require.NoError(t, err)
}
