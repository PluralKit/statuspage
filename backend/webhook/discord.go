package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"pluralkit/status/util"
	"strconv"
	"strings"
)

type DiscordWebhook struct {
	url        string
	notifRole  string
	httpClient *http.Client
}

func NewDiscordWebhook(config util.Config) *DiscordWebhook {
	return &DiscordWebhook{
		url:        config.NotificationWebhook,
		notifRole:  config.NotificationRole,
		httpClient: &http.Client{},
	}
}

type DiscordResponse struct {
	ID string `json:"id"`
}

func (dw *DiscordWebhook) send(content string) (int64, error) {
	url := fmt.Sprintf("%s?with_components=true&wait=true", dw.url)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(content))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := dw.httpClient.Do(req)
	if err != nil {
		return 0, err
	} else if resp.StatusCode != 200 {
		return 0, errors.New("error while sending webhook")
	}

	data := DiscordResponse{}
	_ = json.NewDecoder(resp.Body).Decode(&data)
	num, _ := strconv.ParseInt(data.ID, 10, 64)

	err = resp.Body.Close()
	return num, err
}

func (dw *DiscordWebhook) edit(msgID int64, content string) error {
	url := fmt.Sprintf("%s/messages/%d?with_components=true&wait=true", dw.url, msgID)
	req, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := dw.httpClient.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return errors.New("error while editing webhook")
	}

	err = resp.Body.Close()
	return err
}

func (dw *DiscordWebhook) genIncidentMessage(incident util.Incident) Message {
	var mentions *AllowedMentions = nil
	notifText := "new incident:"
	if dw.notifRole != "" {
		notifText = fmt.Sprintf("<@&%s> new incident:", dw.notifRole)
		mentions = &AllowedMentions{
			Roles: []string{dw.notifRole},
		}
	}
	color := 0
	switch incident.Impact {
	case util.ImpactMinor:
		color = 0xfcb700
	case util.ImpactMajor:
		color = 0xff637d
	default:
		color = 0x99c1f1
	}
	return Message{
		Components: []ComponentBase{
			{
				Type:    int(TextDisplay),
				Content: notifText,
			},
			{
				Type:        int(Container),
				AccentColor: color,
				Components: []ComponentBase{
					{
						Type:    int(TextDisplay),
						Content: "### [PluralKit Status](https://status.pluralkit.me)",
					},
					{
						Type: int(Seperator),
					},
					{
						Type:    int(TextDisplay),
						Content: fmt.Sprintf("## %s\n-# **status:** *%s*\t**impact:** *%s*", incident.Name, incident.Status, incident.Impact),
					},
					{
						Type:    int(Seperator),
						Divider: boolPtr(false),
					},
					{
						Type:    int(TextDisplay),
						Content: incident.Description,
					},
					{
						Type:    int(Seperator),
						Spacing: 2,
					},
					{
						Type:    int(TextDisplay),
						Content: fmt.Sprintf("-# incident id: `%s` · <t:%d:f>", incident.ID, incident.Timestamp.Unix()),
					},
				},
			},
		},
		Flags:           int(ComponentsV2),
		AllowedMentions: mentions,
	}
}

func (dw *DiscordWebhook) genUpdateMessage(incident util.Incident, update util.IncidentUpdate) Message {
	var mentions *AllowedMentions = nil
	notifText := "new status update:"
	if dw.notifRole != "" {
		notifText = fmt.Sprintf("<@&%s> new status update:", dw.notifRole)
		mentions = &AllowedMentions{
			Roles: []string{dw.notifRole},
		}
	}
	nameText := fmt.Sprintf("## update: %s", incident.Name)
	if update.Status != nil {
		nameText = fmt.Sprintf("## update: %s\n-# **status:** *%s*", incident.Name, incident.Status)
	}
	color := 0
	switch incident.Impact {
	case util.ImpactMinor:
		color = 0xfcb700
	case util.ImpactMajor:
		color = 0xff637d
	default:
		color = 0x99c1f1
	}
	if update.Status != nil && *update.Status == util.StatusResolved {
		color = 0x00d390
	}
	msg := Message{
		Components: []ComponentBase{
			{
				Type:    int(TextDisplay),
				Content: notifText,
			},
			{
				Type:        int(Container),
				AccentColor: color,
				Components: []ComponentBase{
					{
						Type:    int(TextDisplay),
						Content: "### [PluralKit Status](https://status.pluralkit.me)",
					},
					{
						Type: int(Seperator),
					},
					{
						Type:    int(TextDisplay),
						Content: nameText,
					},
					{
						Type:    int(Seperator),
						Divider: boolPtr(false),
					},
					{
						Type:    int(TextDisplay),
						Content: update.Text,
					},
					{
						Type:    int(Seperator),
						Spacing: 2,
					},
					{
						Type:    int(TextDisplay),
						Content: fmt.Sprintf("-# update id: `%s` · incident id: `%s` · <t:%d:f>", update.ID, incident.ID, update.Timestamp.Unix()),
					},
				},
			},
		},
		Flags:           int(ComponentsV2),
		AllowedMentions: mentions,
	}
	return msg
}

func (dw *DiscordWebhook) SendIncident(incident util.Incident) (int64, error) {
	content, err := json.Marshal(dw.genIncidentMessage(incident))
	if err != nil {
		return 0, err
	}
	id, err := dw.send(string(content))
	return id, err
}

func (dw *DiscordWebhook) SendUpdate(incident util.Incident, update util.IncidentUpdate) (int64, error) {
	content, err := json.Marshal(dw.genUpdateMessage(incident, update))
	if err != nil {
		return 0, err
	}
	id, err := dw.send(string(content))
	return id, err
}

func (dw *DiscordWebhook) EditIncident(msgID int64, incident util.Incident) error {
	content, err := json.Marshal(dw.genIncidentMessage(incident))
	if err != nil {
		return err
	}
	err = dw.edit(msgID, string(content))
	return err
}

func (dw *DiscordWebhook) EditUpdate(msgID int64, incident util.Incident, update util.IncidentUpdate) error {
	content, err := json.Marshal(dw.genUpdateMessage(incident, update))
	if err != nil {
		return err
	}
	err = dw.edit(msgID, string(content))
	return err
}

// component/helper types below

type AllowedMentions struct {
	Parse       []string `json:"parse,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Users       []string `json:"users,omitempty"`
	RepliedUser bool     `json:"replied_user,omitempty"`
}

type Message struct {
	Flags           int              `json:"flags,omitempty"`
	Components      []ComponentBase  `json:"components,omitempty"`
	AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
}

type ComponentBase struct {
	Type int `json:"type"`
	ID   int `json:"id,omitempty"`

	//Text Display
	Content string `json:"content,omitempty"`

	//Seperator
	Divider *bool `json:"divider,omitempty"`
	Spacing int   `json:"spacing,omitempty"`

	//Container
	AccentColor int             `json:"accent_color,omitempty"`
	Spoiler     *bool           `json:"spoiler,omitempty"`
	Components  []ComponentBase `json:"components,omitempty"`
}

// *sigh* thanks golang
func boolPtr(b bool) *bool {
	return &b
}

type ComponentType int

const (
	TextDisplay ComponentType = 10
	Seperator   ComponentType = 14
	Container   ComponentType = 17
)

type MessageFlags int

const (
	ComponentsV2 MessageFlags = (1 << 15)
)
