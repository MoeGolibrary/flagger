/*
Copyright 2020, 2022 The Flux authors

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

package loadtester

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Add a global variable to store manual traffic control state
var (
	manualStateStore = &ManualStateStore{
		states: make(map[string]*flaggerv1.CanaryManualState),
	}
)

// ManualStateStore holds the manual traffic control states
type ManualStateStore struct {
	sync.RWMutex
	states map[string]*flaggerv1.CanaryManualState
}

// SetState stores a manual state for a canary
func (m *ManualStateStore) SetState(canaryName string, state *flaggerv1.CanaryManualState) {
	m.Lock()
	defer m.Unlock()
	m.states[canaryName] = state
}

// GetState retrieves the manual state for a canary
func (m *ManualStateStore) GetState(canaryName string) *flaggerv1.CanaryManualState {
	m.RLock()
	defer m.RUnlock()
	return m.states[canaryName]
}

// ListenAndServe starts a web server and waits for SIGTERM
func ListenAndServe(port string, timeout time.Duration, logger *zap.SugaredLogger, taskRunner *TaskRunner, gate *GateStorage, authorizer *Authorizer, stopCh <-chan struct{}) {
	mux := http.DefaultServeMux
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", HandleHealthz)
	mux.HandleFunc("/gate/approve", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/gate/halt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	})
	mux.HandleFunc("/gate/check", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		canary := &flaggerv1.CanaryWebhookPayload{}
		err = json.Unmarshal(body, canary)
		if err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !authorizer.Authorize(canary) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		canaryName := fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)
		approved := gate.isOpen(canaryName)
		if approved {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Approved"))
		} else {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
		}

		logger.Infof("%s gate check: approved %v", canaryName, approved)
	})

	mux.HandleFunc("/gate/open", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		canary := &flaggerv1.CanaryWebhookPayload{}
		err = json.Unmarshal(body, canary)
		if err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !authorizer.Authorize(canary) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		canaryName := fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)
		gate.open(canaryName)

		w.WriteHeader(http.StatusAccepted)

		logger.Infof("%s gate opened", canaryName)
	})

	mux.HandleFunc("/gate/close", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		canary := &flaggerv1.CanaryWebhookPayload{}
		err = json.Unmarshal(body, canary)
		if err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !authorizer.Authorize(canary) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		canaryName := fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)
		gate.close(canaryName)

		w.WriteHeader(http.StatusAccepted)

		logger.Infof("%s gate closed", canaryName)
	})

	mux.HandleFunc("/rollback/check", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		canary := &flaggerv1.CanaryWebhookPayload{}
		err = json.Unmarshal(body, canary)
		if err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !authorizer.Authorize(canary) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		canaryName := fmt.Sprintf("rollback.%s.%s", canary.Name, canary.Namespace)
		approved := gate.isOpen(canaryName)
		if approved {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Approved"))
		} else {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
		}

		logger.Infof("%s rollback check: approved %v", canaryName, approved)
	})
	mux.HandleFunc("/rollback/open", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		canary := &flaggerv1.CanaryWebhookPayload{}
		err = json.Unmarshal(body, canary)
		if err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !authorizer.Authorize(canary) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		canaryName := fmt.Sprintf("rollback.%s.%s", canary.Name, canary.Namespace)
		gate.open(canaryName)

		w.WriteHeader(http.StatusAccepted)

		logger.Infof("%s rollback opened", canaryName)
	})
	mux.HandleFunc("/rollback/close", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		canary := &flaggerv1.CanaryWebhookPayload{}
		err = json.Unmarshal(body, canary)
		if err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !authorizer.Authorize(canary) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		canaryName := fmt.Sprintf("rollback.%s.%s", canary.Name, canary.Namespace)
		gate.close(canaryName)

		w.WriteHeader(http.StatusAccepted)

		logger.Infof("%s rollback closed", canaryName)
	})

	// Handle manual traffic control webhook requests
	mux.HandleFunc("/traffic/", HandleSetManualTrafficControl(logger, authorizer))
	mux.HandleFunc("/traffic/state", HandleGetManualTrafficControl(logger, authorizer))

	mux.HandleFunc("/", HandleNewTask(logger, taskRunner, authorizer))
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// run server in background
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatalf("HTTP server crashed %v", err)
		}
	}()

	// wait for SIGTERM or SIGINT
	<-stopCh
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("HTTP server graceful shutdown failed %v", err)
	} else {
		logger.Info("HTTP server stopped")
	}
}

// HandleHealthz handles heath check requests
func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// HandleSetManualTrafficControl handles setting manual traffic control state
func HandleSetManualTrafficControl(logger *zap.SugaredLogger, authorizer *Authorizer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			logger.Errorf("Invalid method for setting manual traffic control: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse the manual state from the request
		var manualState flaggerv1.CanaryManualState
		if err := json.Unmarshal(body, &manualState); err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Add timestamp if not provided
		if manualState.Timestamp == "" {
			manualState.Timestamp = strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
		}

		// For now, we're not doing authorization on manual traffic control
		// In a production environment, you might want to add authorization checks

		// Extract canary identifier from request headers
		canaryName := r.Header.Get("Canary-Name")
		canaryNamespace := r.Header.Get("Canary-Namespace")
		if canaryName == "" || canaryNamespace == "" {
			logger.Error("canary name or namespace not provided in headers")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storeKey := fmt.Sprintf("%s.%s", canaryNamespace, canaryName)

		logger.Infof("Setting manual traffic control for canary %s in namespace %s with weight %d and paused=%v at timestamp %s",
			canaryName, canaryNamespace, manualState.Weight, manualState.Paused, manualState.Timestamp)

		// Store the manual state in a global store
		manualStateStore.SetState(storeKey, &manualState)

		// Return the CanaryWebhookResponse with manualTrafficControl field
		response := struct {
			ManualTrafficControl struct {
				Weight    float64 `json:"weight"`
				Paused    bool    `json:"paused"`
				Timestamp string  `json:"timestamp"`
			} `json:"manualTrafficControl"`
		}{
			ManualTrafficControl: struct {
				Weight    float64 `json:"weight"`
				Paused    bool    `json:"paused"`
				Timestamp string  `json:"timestamp"`
			}{
				Weight:    0,
				Paused:    false,
				Timestamp: "",
			},
		}
		
		// Set the values properly
		if manualState.Weight != nil {
			response.ManualTrafficControl.Weight = float64(*manualState.Weight)
		}
		response.ManualTrafficControl.Paused = manualState.Paused
		response.ManualTrafficControl.Timestamp = manualState.Timestamp

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("encoding response failed", zap.Error(err))
		}
		logger.Infof("Successfully set manual traffic control for canary %s in namespace %s", canaryName, canaryNamespace)
	}
}

// HandleGetManualTrafficControl handles getting manual traffic control state
func HandleGetManualTrafficControl(logger *zap.SugaredLogger, authorizer *Authorizer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			logger.Errorf("Invalid method for getting manual traffic control: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Extract canary info from request body
		var payload flaggerv1.CanaryWebhookPayload
		err = json.Unmarshal(body, &payload)
		if err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		canaryName := payload.Name
		canaryNamespace := payload.Namespace
		storeKey := fmt.Sprintf("%s.%s", canaryNamespace, canaryName)

		logger.Infof("Getting manual traffic control state for canary %s in namespace %s", canaryName, canaryNamespace)

		// Get the stored state
		state := manualStateStore.GetState(storeKey)
		if state == nil {
			// Return empty state if none exists
			state = &flaggerv1.CanaryManualState{}
		}

		// Return the CanaryWebhookResponse with manualTrafficControl field
		response := struct {
			ManualTrafficControl struct {
				Weight    float64 `json:"weight"`
				Paused    bool    `json:"paused"`
				Timestamp string  `json:"timestamp"`
			} `json:"manualTrafficControl"`
		}{
			ManualTrafficControl: struct {
				Weight    float64 `json:"weight"`
				Paused    bool    `json:"paused"`
				Timestamp string  `json:"timestamp"`
			}{
				Weight:    0,
				Paused:    state.Paused,
				Timestamp: state.Timestamp,
			},
		}
		
		// Set the weight value properly
		if state.Weight != nil {
			response.ManualTrafficControl.Weight = float64(*state.Weight)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("encoding response failed", zap.Error(err))
		}
		logger.Infof("Successfully retrieved manual traffic control state for canary %s in namespace %s with weight %d and paused=%v at timestamp %s",
			canaryName, canaryNamespace, state.Weight, state.Paused, state.Timestamp)
	}
}

// HandleNewTask handles task creation requests
func HandleNewTask(logger *zap.SugaredLogger, taskRunner TaskRunnerInterface, authorizer *Authorizer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("reading the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		payload := &flaggerv1.CanaryWebhookPayload{}
		err = json.Unmarshal(body, payload)
		if err != nil {
			logger.Error("decoding the request body failed", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !authorizer.Authorize(payload) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Forbidden"))
			return
		}

		if len(payload.Metadata) > 0 {
			metadata := payload.Metadata
			var typ, ok = metadata["type"]
			if !ok {
				typ = TaskTypeShell
			}

			rtnCmdOutput := false
			if rtn, ok := metadata["returnCmdOutput"]; ok {
				rtnCmdOutput, err = strconv.ParseBool(rtn)
			}

			// run bash command (blocking task)
			if typ == TaskTypeBash {
				logger.With("canary", payload.Name).Infof("bash command %s", payload.Metadata["cmd"])

				bashTask := BashTask{
					command:      payload.Metadata["cmd"],
					logCmdOutput: true,
					TaskBase: TaskBase{
						canary: fmt.Sprintf("%s.%s", payload.Name, payload.Namespace),
						logger: logger,
					},
				}

				ctx, cancel := context.WithTimeout(context.Background(), taskRunner.Timeout())
				defer cancel()

				result, err := bashTask.Run(ctx)
				if !result.ok {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				w.WriteHeader(http.StatusOK)
				if rtnCmdOutput {
					w.Write(result.out)
				}
				return
			}

			// run helmv3 command (blocking task)
			if typ == TaskTypeHelmv3 || typ == TaskTypeHelm {
				helm := HelmTaskv3{
					command:      payload.Metadata["cmd"],
					logCmdOutput: true,
					TaskBase: TaskBase{
						canary: fmt.Sprintf("%s.%s", payload.Name, payload.Namespace),
						logger: logger,
					},
				}

				ctx, cancel := context.WithTimeout(context.Background(), taskRunner.Timeout())
				defer cancel()

				result, err := helm.Run(ctx)
				if !result.ok {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				w.WriteHeader(http.StatusOK)
				if rtnCmdOutput {
					w.Write(result.out)
				}
				return
			}
			//run kubectl cmd
			if typ == TaskTypeKubectl {
				kubectl := KubectlTask{
					command:      payload.Metadata["cmd"],
					logCmdOutput: true,
					TaskBase: TaskBase{
						canary: fmt.Sprintf("%s.%s", payload.Name, payload.Namespace),
						logger: logger,
					},
				}

				ctx, cancel := context.WithTimeout(context.Background(), taskRunner.Timeout())
				defer cancel()

				result, err := kubectl.Run(ctx)
				if !result.ok {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				w.WriteHeader(http.StatusOK)
				if rtnCmdOutput {
					w.Write(result.out)
				}
				return
			}
			// run concord job (blocking task)
			if typ == TaskTypeConcord {
				concord, err := NewConcordTask(payload.Metadata, fmt.Sprintf("%s.%s", payload.Name, payload.Namespace), logger)

				if err != nil {
					logger.With("canary", payload.Name).Errorf("concord task init error: %s", err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				ctx, cancel := context.WithTimeout(context.Background(), taskRunner.Timeout())
				defer cancel()

				result, err := concord.Run(ctx)
				if !result.ok {
					if err != nil {
						logger.With("canary", payload.Name).Errorf("concord task error: %s", err)
					}
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				w.WriteHeader(http.StatusOK)
				if rtnCmdOutput {
					w.Write(result.out)
				}
				return
			}

			taskFactory, ok := GetTaskFactory(typ)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("unknown task type %s", typ)))
				return
			}
			canary := fmt.Sprintf("%s.%s", payload.Name, payload.Namespace)
			task, err := taskFactory(metadata, canary, logger)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			taskRunner.Add(task)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("metadata not found in payload"))
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
