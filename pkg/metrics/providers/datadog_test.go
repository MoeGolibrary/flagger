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

package providers

import (
	"errors"
	"fmt"
	"github.com/fluxcd/flagger/pkg/logger"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
)

var mylog, _ = logger.NewLogger("debug")

func TestNewDatadogProvider(t *testing.T) {
	appKey := "app-key"
	apiKey := "api-key"
	cs := map[string][]byte{
		datadogApplicationKeySecretKey: []byte(appKey),
		datadogAPIKeySecretKey:         []byte(apiKey),
	}

	mi := "100s"
	md, err := time.ParseDuration(mi)
	require.NoError(t, err)

	dp, err := NewDatadogProvider("100s", "", flaggerv1.MetricTemplateProvider{}, cs, mylog)
	require.NoError(t, err)
	assert.Equal(t, "https://api.datadoghq.com/api/v1/validate", dp.apiKeyValidationEndpoint)
	assert.Equal(t, "https://api.datadoghq.com/api/v1/query", dp.metricsQueryEndpoint)
	assert.Equal(t, int64(md.Seconds()*datadogFromDeltaMultiplierOnMetricInterval), dp.fromDelta)
	assert.Equal(t, appKey, dp.applicationKey)
	assert.Equal(t, apiKey, dp.apiKey)
}

func TestNewDatadogProvider_MissingKeys(t *testing.T) {
	// 测试无密钥时的错误处理
	cs := map[string][]byte{}
	_, err := NewDatadogProvider("1m", "", flaggerv1.MetricTemplateProvider{}, cs, mylog)
	require.Error(t, err)
	assert.EqualError(t, err, "no valid datadog keys found")
}

func TestNewDatadogProvider_InvalidInterval(t *testing.T) {
	// 测试无效时间间隔的错误
	_, err := NewDatadogProvider("invalid", "", flaggerv1.MetricTemplateProvider{}, map[string][]byte{
		datadogAPIKeySecretKey:         []byte("api-key"),
		datadogApplicationKeySecretKey: []byte("app-key"),
	}, mylog)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing metric interval")
}

func TestNewDatadogProvider_HistoryWindow(t *testing.T) {
	// 验证历史窗口配置
	dp, _ := NewDatadogProvider("1m", "2h", flaggerv1.MetricTemplateProvider{}, map[string][]byte{
		datadogAPIKeySecretKey:         []byte("api-key"),
		datadogApplicationKeySecretKey: []byte("app-key"),
	}, mylog)
	assert.Equal(t, int64(2*60*60), dp.history)
}

func TestDatadogProvider_ExecuteCurrentQuery_ErrorCases(t *testing.T) {
	appKey := "app-key"
	apiKey := "api-key"
	t.Run("invalid json response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`invalid json`))
		}))
		defer ts.Close()
		dp, _ := newTestProvider(ts.URL, appKey, apiKey)
		_, err := dp.ExecuteCurrentQuery("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error unmarshaling result")
	})

	t.Run("no series", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json := `{"series": []}`
			w.Write([]byte(json))
		}))
		defer ts.Close()
		dp, _ := newTestProvider(ts.URL, appKey, apiKey)
		_, err := dp.ExecuteCurrentQuery("")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoValuesFound)
	})

	t.Run("too many requests", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer ts.Close()
		dp, _ := newTestProvider(ts.URL, appKey, apiKey)
		_, err := dp.ExecuteCurrentQuery("")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTooManyRequests)
	})
}

func TestDatadogProvider_GetPreviousMetricValue(t *testing.T) {
	appKey := "app-key"
	apiKey := "api-key"
	t.Run("valid time window", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json := `{"series": [{"pointlist": [[1,1]]}]}`
			w.Write([]byte(json))
		}))
		defer ts.Close()
		dp, _ := NewDatadogProvider("1m", "1h", flaggerv1.MetricTemplateProvider{Address: ts.URL}, map[string][]byte{
			datadogAPIKeySecretKey:         []byte(apiKey),
			datadogApplicationKeySecretKey: []byte(appKey),
		}, mylog)
		val, err := dp.GetPreviousMetricValue("")
		require.NoError(t, err)
		assert.Equal(t, float64(1), val)
	})

	t.Run("time window not configured", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer ts.Close()
		dp, _ := NewDatadogProvider("1m", "", flaggerv1.MetricTemplateProvider{Address: ts.URL}, map[string][]byte{
			datadogAPIKeySecretKey:         []byte(apiKey),
			datadogApplicationKeySecretKey: []byte(appKey),
		}, mylog)
		_, err := dp.GetPreviousMetricValue("")
		require.Error(t, err)
		assert.EqualError(t, err, "historical time window not configured")
	})
}

func TestDatadogProvider_IsOnline_ErrorCases(t *testing.T) {
	appKey := "app-key"
	apiKey := "api-key"
	t.Run("invalid status code", func(t *testing.T) {
		for _, code := range []int{http.StatusInternalServerError, http.StatusForbidden} {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
			}))
			defer ts.Close()
			dp, _ := newTestProvider(ts.URL, appKey, apiKey)
			_, err := dp.IsOnline()
			require.Error(t, err)
		}
	})

	t.Run("invalid response body", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		}))
		defer ts.Close()
		dp, _ := newTestProvider(ts.URL, appKey, apiKey)
		_, err := dp.IsOnline()
		require.Error(t, err)
	})
}

// 辅助函数
func newTestProvider(url, appKey, apiKey string) (*DatadogProvider, error) {
	return NewDatadogProvider("1m", "", flaggerv1.MetricTemplateProvider{Address: url}, map[string][]byte{
		datadogAPIKeySecretKey:         []byte(apiKey),
		datadogApplicationKeySecretKey: []byte(appKey),
	}, mylog)
}
func TestDatadogProvider_RunQuery(t *testing.T) {
	appKey := "app-key"
	apiKey := "api-key"
	t.Run("ok", func(t *testing.T) {
		expected := 1.11111
		eq := `avg:system.cpu.user{*}by{host}`
		now := time.Now().Unix()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			aq := r.URL.Query().Get("query")
			assert.Equal(t, eq, aq)
			assert.Equal(t, appKey, r.Header.Get(datadogApplicationKeyHeaderKey))
			assert.Equal(t, apiKey, r.Header.Get(datadogAPIKeyHeaderKey))

			from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
			if assert.NoError(t, err) {
				assert.Less(t, from, now)
			}

			to, err := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
			if assert.NoError(t, err) {
				assert.GreaterOrEqual(t, to, now)
			}

			json := fmt.Sprintf(`{"series": [{"pointlist": [[1577232000000,29325.102158814265],[1577318400000,56294.46758591842],[1577404800000,%f]]}]}`, expected)
			w.Write([]byte(json))
		}))
		defer ts.Close()

		dp, err := NewDatadogProvider("1m",
			"",
			flaggerv1.MetricTemplateProvider{Address: ts.URL},
			map[string][]byte{
				datadogApplicationKeySecretKey: []byte(appKey),
				datadogAPIKeySecretKey:         []byte(apiKey),
			},
			mylog,
		)
		require.NoError(t, err)

		f, err := dp.ExecuteCurrentQuery(eq)
		require.NoError(t, err)
		assert.Equal(t, expected, f)
	})

	t.Run("no values", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json := `{"series": [{"pointlist": []}]}`
			w.Write([]byte(json))
		}))
		defer ts.Close()

		dp, err := NewDatadogProvider("1m",
			"",
			flaggerv1.MetricTemplateProvider{Address: ts.URL},
			map[string][]byte{
				datadogApplicationKeySecretKey: []byte(appKey),
				datadogAPIKeySecretKey:         []byte(apiKey),
			},
			mylog,
		)
		require.NoError(t, err)
		_, err = dp.ExecuteCurrentQuery("")
		require.True(t, errors.Is(err, ErrNoValuesFound))
	})
}

func TestDatadogProvider_IsOnline(t *testing.T) {
	for _, c := range []struct {
		code        int
		errExpected bool
	}{
		{code: http.StatusOK, errExpected: false},
		{code: http.StatusUnauthorized, errExpected: true},
	} {
		t.Run(fmt.Sprintf("%d", c.code), func(t *testing.T) {
			appKey := "app-key"
			apiKey := "api-key"
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, appKey, r.Header.Get(datadogApplicationKeyHeaderKey))
				assert.Equal(t, apiKey, r.Header.Get(datadogAPIKeyHeaderKey))
				w.WriteHeader(c.code)
				if c.code == http.StatusOK {
					w.Write([]byte(`{"valid": true}`))
				}
			}))
			defer ts.Close()

			dp, err := NewDatadogProvider("1m",
				"",
				flaggerv1.MetricTemplateProvider{Address: ts.URL},
				map[string][]byte{
					datadogApplicationKeySecretKey: []byte(appKey),
					datadogAPIKeySecretKey:         []byte(apiKey),
					datadogKeysSecretKey:           []byte(`[{"api_key":"api-key","application_key":"app-key"}]`),
				},
				mylog,
			)
			require.NoError(t, err)

			_, err = dp.IsOnline()
			if c.errExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
