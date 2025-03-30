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

package controller

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"net/http"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	flaggerv1 "github.com/fluxcd/flagger/pkg/apis/flagger/v1beta1"
	"github.com/fluxcd/flagger/pkg/notifier"
)

var ra = rand.New(rand.NewSource(time.Now().UnixNano()))

func (c *Controller) recordEventInfof(r *flaggerv1.Canary, template string, args ...interface{}) {
	c.logger.
		With("canary", fmt.Sprintf("%s.%s", r.Name, r.Namespace)).
		With("canary_name", r.Name).
		With("canary_namespace", r.Namespace).
		Infof(template, args...)
	c.eventRecorder.Event(r, corev1.EventTypeNormal, "Synced", fmt.Sprintf(template, args...))
	c.sendEventToWebhook(r, corev1.EventTypeNormal, template, args)
}

func (c *Controller) recordEventErrorf(r *flaggerv1.Canary, template string, args ...interface{}) {
	c.logger.With("canary", fmt.Sprintf("%s.%s", r.Name, r.Namespace)).
		With("canary_name", r.Name).
		With("canary_namespace", r.Namespace).
		Errorf(template, args...)
	c.eventRecorder.Event(r, corev1.EventTypeWarning, "Synced", fmt.Sprintf(template, args...))
	c.sendEventToWebhook(r, corev1.EventTypeWarning, template, args)
}

func (c *Controller) recordEventWarningf(r *flaggerv1.Canary, template string, args ...interface{}) {
	c.logger.With("canary", fmt.Sprintf("%s.%s", r.Name, r.Namespace)).
		With("canary_name", r.Name).
		With("canary_namespace", r.Namespace).
		Infof(template, args...)
	c.eventRecorder.Event(r, corev1.EventTypeWarning, "Synced", fmt.Sprintf(template, args...))
	c.sendEventToWebhook(r, corev1.EventTypeWarning, template, args)
}

func (c *Controller) sendEventToWebhook(r *flaggerv1.Canary, eventType, template string, args []interface{}) {
	webhookOverride := false
	for _, canaryWebhook := range r.GetAnalysis().Webhooks {
		if canaryWebhook.Type == flaggerv1.EventHook {
			webhookOverride = true
			err := CallEventWebhook(r, canaryWebhook, fmt.Sprintf(template, args...), eventType)
			if err != nil {
				c.logger.With("canary", fmt.Sprintf("%s.%s", r.Name, r.Namespace)).
					With("canary_name", r.Name).
					With("canary_namespace", r.Namespace).
					Errorf("error sending event to webhook: %s", err)
			}
		}
	}

	if c.eventWebhook != "" && !webhookOverride {
		hook := flaggerv1.CanaryWebhook{
			Name: "events",
			URL:  c.eventWebhook,
		}
		err := CallEventWebhook(r, hook, fmt.Sprintf(template, args...), eventType)
		if err != nil {
			c.logger.With("canary", fmt.Sprintf("%s.%s", r.Name, r.Namespace)).
				With("canary_name", r.Name).
				With("canary_namespace", r.Namespace).
				Errorf("error sending event to webhook: %s", err)
		}
	}
}

func (c *Controller) alert(canary *flaggerv1.Canary, message string, metadata bool, severity flaggerv1.AlertSeverity) {
	var fields []notifier.Field

	if c.clusterName != "" {
		fields = append(fields,
			notifier.Field{
				Name:  "Cluster",
				Value: c.clusterName,
			},
		)
	}

	if metadata {
		fields = append(fields, alertMetadata(canary)...)
	}
	if c.sendAt {
		var err error
		var githubUrl, githubActionUrl string
		message, githubUrl, githubActionUrl, err = c.getCommitters(canary, message, severity)
		if err != nil {
			c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
				With("canary_name", canary.Name).
				With("canary_namespace", canary.Namespace).
				Errorf("alert parse msg: %v", err)
		}
		if githubUrl != "" {
			fields = append(fields, notifier.Field{
				Name:  "Github Repo",
				Value: githubUrl,
				Type:  "link",
			})
		}
		if githubActionUrl != "" {
			fields = append(fields, notifier.Field{
				Name:  "Github Action",
				Value: githubActionUrl,
				Type:  "link",
			})
		}
	}

	// add canary dashboard link
	canaryURL := getCanaryURL(canary)
	fields = append(fields, notifier.Field{
		Name:  "Canary Dashboard",
		Value: canaryURL,
		Type:  "link",
	}, notifier.Field{
		Name:  "Datadog Service Page",
		Value: getServiceURL(canary),
		Type:  "link",
	})
	// set alert message
	if severity == flaggerv1.SeverityError {
		message = fmt.Sprintf("%s \n\n Please check the **Canary Dashboard**  to find out why error.\n", message)
	}

	// get canaryId
	canaryId := ""
	for _, canaryWebhook := range canary.GetAnalysis().Webhooks {
		if canaryWebhook.Type == flaggerv1.SkipHook || canaryWebhook.Type == flaggerv1.RollbackHook {
			canaryId = canary.CanaryChecksum()
			break
		}
	}

	// send alert with the global notifier
	if len(canary.GetAnalysis().Alerts) == 0 {
		err := c.notifier.Post(canary.Name, canary.Namespace, message, fields, string(severity), canaryId)
		if err != nil {
			c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
				With("canary_name", canary.Name).
				With("canary_namespace", canary.Namespace).
				Errorf("alert can't be sent: %v", err)
			return
		}
		return
	}

	// send canary alerts
	for _, alert := range canary.GetAnalysis().Alerts {
		// determine if alert should be sent based on severity level
		shouldAlert := false
		if alert.Severity == flaggerv1.SeverityInfo {
			shouldAlert = true
		} else {
			if severity == alert.Severity {
				shouldAlert = true
			}
			if severity == flaggerv1.SeverityError && alert.Severity == flaggerv1.SeverityWarn {
				shouldAlert = true
			}
		}
		if !shouldAlert {
			continue
		}

		// determine alert provider namespace
		providerNamespace := canary.GetNamespace()
		if alert.ProviderRef.Namespace != canary.Namespace && alert.ProviderRef.Namespace != "" {
			providerNamespace = alert.ProviderRef.Namespace
		}

		// find alert provider
		provider, err := c.flaggerInformers.AlertInformer.Lister().AlertProviders(providerNamespace).Get(alert.ProviderRef.Name)
		if err != nil {
			c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
				With("canary_name", canary.Name).
				With("canary_namespace", canary.Namespace).
				Errorf("alert provider %s.%s error: %v", alert.ProviderRef.Name, providerNamespace, err)
			continue
		}

		// set hook URL address
		url := provider.Spec.Address

		// set the token which will be sent in the header
		// https://datatracker.ietf.org/doc/html/rfc6750
		token := ""

		// extract address from secret
		if provider.Spec.SecretRef != nil {
			secret, err := c.kubeClient.CoreV1().Secrets(providerNamespace).Get(context.TODO(), provider.Spec.SecretRef.Name, metav1.GetOptions{})
			if err != nil {
				c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
					With("canary_name", canary.Name).
					With("canary_namespace", canary.Namespace).
					Errorf("alert provider %s.%s secretRef error: %v", alert.ProviderRef.Name, providerNamespace, err)
				continue
			}
			if address, ok := secret.Data["address"]; ok {
				url = string(address)
			} else {
				c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
					With("canary_name", canary.Name).
					With("canary_namespace", canary.Namespace).
					Errorf("alert provider %s.%s secret does not contain an address", alert.ProviderRef.Name, providerNamespace)
				continue
			}

			if tokenFromSecret, ok := secret.Data["token"]; ok {
				token = string(tokenFromSecret)
			}
		}

		// set defaults
		username := "flagger"
		if provider.Spec.Username != "" {
			username = provider.Spec.Username
		}
		channel := "general"
		if provider.Spec.Channel != "" {
			channel = provider.Spec.Channel
		}
		proxy := ""
		if provider.Spec.Proxy != "" {
			proxy = provider.Spec.Proxy
		}

		// create notifier based on provider type
		f := notifier.NewFactory(url, token, proxy, username, channel)
		n, err := f.Notifier(provider.Spec.Type)
		if err != nil {
			c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
				With("canary_name", canary.Name).
				With("canary_namespace", canary.Namespace).
				Errorf("alert provider %s.%s error: %v", alert.ProviderRef.Name, providerNamespace, err)
			continue
		}

		// send alert
		err = n.Post(canary.Name, canary.Namespace, message, fields, string(severity), canaryId)
		if err != nil {
			c.logger.With("canary", fmt.Sprintf("%s.%s", canary.Name, canary.Namespace)).
				With("canary_name", canary.Name).
				With("canary_namespace", canary.Namespace).
				Errorf("alert provider $s.%s send error: %v", alert.ProviderRef.Name, providerNamespace, err)
		}

	}
}

// https://docs.datadoghq.com/api/
const (
	dashboardBaseURL = "https://us5.datadoghq.com/dashboard/5pp-9u8-u3i/moego-canary"
	queryParams      = "?fromUser=false&refresh_mode=sliding&live=true"
	tplVarCanary     = "tpl_var_canary%5B0%5D="
	tplVarNamespace  = "tpl_var_namespace%5B0%5D="
	tplVarPrimary    = "tpl_var_primary%5B0%5D="

	datadogKeysSecretKey           = "datadog_keys"
	datadogAPIKeyHeaderKey         = "DD-API-KEY"
	datadogApplicationKeyHeaderKey = "DD-APPLICATION-KEY"
)

type DatadogKey struct {
	ApiKey         string `json:"api_key" yaml:"apiKey"`
	ApplicationKey string `json:"application_key" yaml:"applicationKey"`
}

type datadogCiPipelineEventAttributes struct {
	Git struct {
		Commit struct {
			Committer struct {
				Date          string `json:"date"`
				Name          string `json:"name"`
				Email         string `json:"email"`
				DateTimestamp int64  `json:"date_timestamp"`
			} `json:"committer"`
			Author struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
			Message string `json:"message"`
			Sha     string `json:"sha"`
		} `json:"commit"`
		RepositoryUrl string `json:"repository_url"`
		DefaultBranch string `json:"default_branch"`
		Branch        string `json:"branch"`
		Repository    struct {
			Path   string `json:"path"`
			Scheme string `json:"scheme"`
			IdV2   string `json:"id_v2"`
			Name   string `json:"name"`
			Host   string `json:"host"`
			Id     string `json:"id"`
		} `json:"repository"`
	} `json:"git"`
	Github struct {
		Conclusion string `json:"conclusion"`
		NodeGroup  string `json:"node_group"`
		HtmlUrl    string `json:"html_url"`
		RunAttempt string `json:"run_attempt"`
		AppId      string `json:"app_id"`
	} `json:"github"`
	Ci struct {
		Pipeline struct {
			FingerprintV2 string `json:"fingerprint_v2"`
			Fingerprint   string `json:"fingerprint"`
			Name          string `json:"name"`
			Id            string `json:"id"`
			Url           string `json:"url"`
		} `json:"pipeline"`
		Node struct {
			Name   string   `json:"name"`
			Labels []string `json:"labels"`
		} `json:"node"`
		Provider struct {
			Instance string `json:"instance"`
			Name     string `json:"name"`
		} `json:"provider"`
		Step struct {
			Name string `json:"name"`
		} `json:"step"`
		Job struct {
			Name string `json:"name"`
			Id   string `json:"id"`
			Url  string `json:"url"`
		} `json:"job"`
		Status string `json:"status"`
	} `json:"ci"`
	Start int64 `json:"start"`
}

func (c *Controller) initDatadogKeys() error {
	// get datadog keys
	var keys []DatadogKey
	secret, err := c.kubeClient.CoreV1().Secrets("flagger-system").Get(context.TODO(), "datadog-secret", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("alert datadog no secret keys ")
	}
	credentials := secret.Data

	if b, ok := credentials[datadogKeysSecretKey]; ok {
		if err := json.Unmarshal(b, &keys); err != nil {
			return fmt.Errorf("alert datadog no secret keys ")
		}
	}
	if len(keys) == 0 {
		return fmt.Errorf("alert datadog no secret keys ")
	}
	c.datadogKeys = keys
	return nil
}

func (c *Controller) getCommitters(canary *flaggerv1.Canary, message string, severity flaggerv1.AlertSeverity) (string, string, string, error) {
	targetMessage := message
	var githubUrl, githubActionUrl string

	_, _, _, labels, _ := c.canaryFactory.Controller(canary.Spec.TargetRef.Kind).GetMetadata(canary)

	if labels == nil {
		return targetMessage, githubUrl, githubActionUrl, fmt.Errorf("alert datadog no labels")
	}
	repo := labels["repo-name"]
	if repo == "" {
		repo = labels["tags.datadoghq.com/service"]
	}
	if repo == "" {
		return targetMessage, githubUrl, githubActionUrl, fmt.Errorf("alert datadog no secret keys ")
	}
	branch := labels["branch"]

	// 随机从keys选择一个
	key := c.datadogKeys[ra.Intn(len(c.datadogKeys))]
	c.logger.Debugf("Datadog API Key: %s", key.ApiKey)

	ctx := context.WithValue(
		context.Background(),
		datadog.ContextServerVariables,
		map[string]string{
			"site": "us5.datadoghq.com",
		},
	)
	configuration := datadog.NewConfiguration()
	configuration.Debug = c.logger.Level() == zapcore.DebugLevel
	configuration.RetryConfiguration.EnableRetry = true
	configuration.RetryConfiguration.MaxRetries = 5
	configuration.AddDefaultHeader(datadogAPIKeyHeaderKey, key.ApiKey)
	configuration.AddDefaultHeader(datadogApplicationKeyHeaderKey, key.ApplicationKey)

	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV2.NewCIVisibilityPipelinesApi(apiClient)

	body := *datadogV2.
		NewListCIAppPipelineEventsOptionalParameters().
		WithFilterQuery(fmt.Sprintf("ci_level:pipeline @git.repository.id:\"github.com/MoeGolibrary/%s\" @git.branch:%s", repo, branch)).
		WithFilterFrom(time.Now().Add(time.Hour * 24 * -30)).
		WithFilterTo(time.Now()).
		WithPageLimit(1000).
		WithSort(datadogV2.CIAPPSORT_TIMESTAMP_ASCENDING)
	resp, r, err := api.ListCIAppPipelineEvents(ctx, body)

	if err != nil || r.StatusCode != 200 {
		return targetMessage, githubUrl, githubActionUrl, fmt.Errorf("alert datadog error: %v", err)
	}
	data, ok := resp.GetDataOk()
	if !ok {
		return targetMessage, githubUrl, githubActionUrl, fmt.Errorf("alert datadog no events")
	}
	targetMessage += "\n*Commits*\n"
	var emails = make(map[string]string)
	for index, event := range *data {
		eventAttributes, ok := event.GetAttributesOk()
		if !ok {
			continue
		}
		var attributes datadogCiPipelineEventAttributes
		jsonbyte, err := json.Marshal(eventAttributes.GetAttributes())
		if err != nil {
			continue
		}
		err = json.Unmarshal(jsonbyte, &attributes)
		if err != nil {
			continue
		}
		if githubUrl == "" {
			githubUrl = attributes.Git.RepositoryUrl
		}
		if githubActionUrl == "" {
			githubActionUrl = attributes.Ci.Pipeline.Url
		}
		if _, ok := emails[attributes.Git.Commit.Author.Email]; !ok {
			emails[attributes.Git.Commit.Author.Email] = c.getSlackUserByEmail(attributes.Git.Commit.Author.Email)
			if severity == flaggerv1.SeverityError || severity == flaggerv1.SeverityWarn {
				barkMessageContent := fmt.Sprintf(
					"*%s %s %s: %s » %s (%s)* \n <%s|Github Repo> <%s|Github Action> <%s|Canary Dashboard> <%s|Datadog Srvice>",
					getEmojiMsg(severity),
					"Canary ",
					severity,
					attributes.Git.Repository.Name,
					attributes.Git.Branch,
					message,
					attributes.Git.RepositoryUrl,
					attributes.Ci.Pipeline.Url,
					getCanaryURL(canary),
					getServiceURL(canary),
				)

				go c.barkMessage(attributes.Git.Commit.Author.Email, barkMessageContent)
			}
		}
		timeString := time.Unix(attributes.Start, 0).In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04:05 (CST)")
		targetMessage += fmt.Sprintf(
			"%d. <https://github.com/%s/commit/%s|%s> %s %s<%s> %s <@%s>\n",
			index+1,
			attributes.Git.Repository.Name,
			attributes.Git.Commit.Sha[:7],
			attributes.Git.Commit.Sha[:7],
			timeString,
			attributes.Git.Commit.Author.Name,
			attributes.Git.Commit.Author.Email,
			strings.Replace(attributes.Git.Commit.Message, "\n", " ", -1),
			emails[attributes.Git.Commit.Author.Email],
		)
	}

	c.logger.Debugf("fainl Message: %s", targetMessage)

	return targetMessage, githubUrl, githubActionUrl, nil
}

func getEmojiMsg(severity flaggerv1.AlertSeverity) string {
	switch severity {
	case flaggerv1.SeveritySuccess:
		return "🎉"
	case flaggerv1.SeverityInfo:
		return "🟠"
	case flaggerv1.SeverityWarn:
		return "⚠️"
	case flaggerv1.SeverityError:
		return "❌"
	default:
		return "🟠"
	}
}

func getCanaryURL(canary *flaggerv1.Canary) string {
	return fmt.Sprintf(
		"%s%s&%s%s&%s%s&%s%s-primary",
		dashboardBaseURL,
		queryParams,
		tplVarCanary,
		canary.GetName(),
		tplVarNamespace,
		canary.GetNamespace(),
		tplVarPrimary,
		canary.GetName(),
	)
}

func getServiceURL(canary *flaggerv1.Canary) string {
	const serviceURL = "https://us5.datadoghq.com/apm/entity/service%3A"
	return fmt.Sprintf(
		"%s%s?env=%s",
		serviceURL,
		canary.GetName(),
		canary.GetNamespace(),
	)
}

func alertMetadata(canary *flaggerv1.Canary) []notifier.Field {
	var fields []notifier.Field

	fields = append(fields,
		notifier.Field{
			Name:  "Target",
			Value: fmt.Sprintf("%s/%s.%s", canary.Spec.TargetRef.Kind, canary.Spec.TargetRef.Name, canary.Namespace),
		},
		notifier.Field{
			Name:  "Failed checks threshold",
			Value: fmt.Sprintf("%v", canary.GetAnalysisThreshold()),
		},
		notifier.Field{
			Name:  "Progress deadline",
			Value: fmt.Sprintf("%vs", canary.GetProgressDeadlineSeconds()),
		},
	)

	if canary.GetAnalysis().StepWeight > 0 {
		fields = append(fields, notifier.Field{
			Name: "Traffic routing",
			Value: fmt.Sprintf("Weight step: %v max: %v interval: %v",
				canary.GetAnalysis().StepWeight,
				canary.GetAnalysis().MaxWeight,
				canary.GetAnalysis().Interval),
		})
	} else if len(canary.GetAnalysis().StepWeights) > 0 {
		fields = append(fields, notifier.Field{
			Name: "Traffic routing",
			Value: fmt.Sprintf("Weight steps: %s max: %v",
				strings.Trim(strings.Join(strings.Fields(fmt.Sprint(canary.GetAnalysis().StepWeights)), ","), "[]"),
				canary.GetAnalysis().MaxWeight),
		})
	} else if len(canary.GetAnalysis().Match) > 0 {
		fields = append(fields, notifier.Field{
			Name:  "Traffic routing",
			Value: "A/B Testing",
		})
	} else if canary.GetAnalysis().Iterations > 0 {
		fields = append(fields, notifier.Field{
			Name:  "Traffic routing",
			Value: "Blue/Green",
		})
	}
	return fields
}

type barkContent struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

func (c *Controller) barkMessage(email, message string) error {
	c.logger.Debugf("Bark message to %s: %s", email, message)
	jsonBody, err := json.Marshal(barkContent{
		Email:   email,
		Message: message,
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal content")
	}

	body := bytes.NewReader(jsonBody)
	http.Post("https://pearl.baobo.me/api/ci/notify", "application/json", body)
	return nil
}

type SlackProfile struct {
	Email string `json:"email"`
}

type SlackUser struct {
	ID      string       `json:"id"`
	Profile SlackProfile `json:"profile"`
}

type SlackUsersResponse struct {
	Ok      bool         `json:"ok"`
	Members []*SlackUser `json:"members"`
}

//go:embed slack-users.json
var slackUsersJSON string

func (c *Controller) initSlackUsers() error {
	c.slackUsersMap = make(map[string]*SlackUser)
	var slackUsersResponse SlackUsersResponse

	// 输出JSON内容以验证加载
	c.logger.Debugf("Loaded slack-users.json content: %s", slackUsersJSON)

	if err := json.Unmarshal([]byte(slackUsersJSON), &slackUsersResponse); err != nil {
		return errors.Wrap(err, "failed to unmarshal slack users")
	}

	// 输出解析后的数据以验证映射
	for _, user := range slackUsersResponse.Members {
		c.logger.Debugf("Adding user to map: Email=%s, ID=%s", user.Profile.Email, user.ID)
		c.slackUsersMap[user.Profile.Email] = user
	}

	return nil
}

func (c *Controller) getSlackUserByEmail(email string) string {
	if user, ok := c.slackUsersMap[email]; ok {
		return user.ID
	}
	return email
}
