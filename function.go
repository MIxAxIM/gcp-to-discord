package function

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	colorGreen = 3066993  // Green for resolved incidents
	colorRed   = 15158332 // Red for open incidents
	colorGrey  = 9807270  // Grey for unknown/default
)

type Notification struct {
	Incident Incident `json:"incident"`
	Version  string   `json:"version"`
}

type Incident struct {
	IncidentID    string `json:"incident_id"`
	ResourceID    string `json:"resource_id"`
	ResourceName  string `json:"resource_name"`
	State         string `json:"state"`
	StartedAt     int64  `json:"started_at"`
	EndedAt       int64  `json:"ended_at,omitempty"`
	PolicyName    string `json:"policy_name"`
	ConditionName string `json:"condition_name"`
	URL           string `json:"url"`
	Summary       string `json:"summary"`
}

type DiscordWebhook struct {
	Embeds []Embed `json:"embeds,omitempty"`
}

type Embed struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Fields      []Field `json:"fields,omitempty"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func toDiscord(notification Notification) DiscordWebhook {
	startedAt := "-"
	endedAt := "-"

	if st := notification.Incident.StartedAt; st > 0 {
		startedAt = time.Unix(st, 0).String()
	}

	if et := notification.Incident.EndedAt; et > 0 {
		endedAt = time.Unix(et, 0).String()
	}

	policyName := notification.Incident.PolicyName
	if policyName == "" {
		policyName = "-"
	}

	conditionName := notification.Incident.ConditionName
	if conditionName == "" {
		conditionName = "-"
	}

	colour := colorGrey
	if notification.Incident.State == "open" {
		colour = colorRed
	} else if notification.Incident.State == "closed" {
		colour = colorGreen
	}

	return DiscordWebhook{
		Embeds: []Embed{
			{
				Title: notification.Incident.Summary,
				URL:   notification.Incident.URL,
				Color: colour,
				Fields: []Field{
					{
						Name:  "Incident ID",
						Value: notification.Incident.IncidentID,
					},
					{
						Name:   "Policy",
						Value:  policyName,
						Inline: true,
					},
					{
						Name:   "Condition",
						Value:  conditionName,
						Inline: true,
					},
					{
						Name:  "Started At",
						Value: startedAt,
					},
					{
						Name:  "Ended At",
						Value: endedAt,
					},
				},
			},
		},
	}
}

var (
	authToken         = os.Getenv("AUTH_TOKEN")
	discordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
	// httpClient is a shared HTTP client for making requests to Discord.
	// It has a timeout to prevent requests from hanging indefinitely.
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

func F(w http.ResponseWriter, r *http.Request) {
	if authToken == "" {
		log.Printf("Error: `AUTH_TOKEN` is not set in the environment")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("auth_token") != authToken {
		log.Printf("Error: Invalid auth_token provided")
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	if discordWebhookURL == "" {
		log.Printf("Error: `DISCORD_WEBHOOK_URL` is not set in the environment")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err := url.Parse(discordWebhookURL); err != nil {
		log.Printf("Error parsing DISCORD_WEBHOOK_URL: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" {
		log.Printf("Error: Invalid request method or content type. Method: %s, Content-Type: %s", r.Method, r.Header.Get("Content-Type"))
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	var notification Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		log.Printf("Error decoding notification payload: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	discordWebhook := toDiscord(notification)

	payload, err := json.Marshal(discordWebhook)
	if err != nil {
		log.Printf("Error marshalling Discord webhook payload: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error sending webhook to Discord: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Printf("Error: Unexpected status code from Discord: %d, payload: %s", res.StatusCode, string(payload))
		http.Error(w, "Failed to send to Discord", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
