package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex
	// configuration is the active plugin configuration. Consult getConfiguration and setConfiguration for usage.
	configuration *configuration
}

func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch path := r.URL.Path; path {
	case "/api/v1/meetings":
		p.handleStartMeeting(w, r)
	default:
		http.NotFound(w, r)
	}
}

type GfycatPostRequest struct {
	ChannelID  string `json:"channel_id"`
	Personal   bool   `json:"personal"`
	Topic      string `json:"topic"`
	SearchTerm string `json:"meeting_id"`
}

func (p *Plugin) handleStartMeeting(w http.ResponseWriter, r *http.Request) {
	if err := p.getConfiguration().IsValid(); err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
		return
	}

	userID := r.Header.Get("Mattermost-User-Id")

	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var req GfycatPostRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	var user *model.User
	var err *model.AppError
	user, err = p.API.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), err.StatusCode)
	}

	if _, err := p.API.GetChannelMember(req.ChannelID, user.Id); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Set the gfycat url
	gifURL := "asdf"

	post := &model.Post{
		UserId:    user.Id,
		ChannelId: req.ChannelID,
		Message:   "",
		Type:      "custom_gfycat",
		Props: map[string]interface{}{
			"gif_url":           gifURL,
			"from_webhook":      "true",
			"override_username": "Jitsi",
			"override_icon_url": "https://s3.amazonaws.com/mattermost-plugin-media/Zoom+App.png",
		},
	}

	if _, err := p.API.CreatePost(post); err != nil {
		http.Error(w, err.Error(), err.StatusCode)
		return
	}
}
