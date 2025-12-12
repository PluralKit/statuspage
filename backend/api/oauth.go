package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

type DiscordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

var discordOauthConfig *oauth2.Config

func (a *API) oauthInit() {
	gob.Register(DiscordUser{})
	baseUrl := ""
	if a.Config.RunDev {
		baseUrl = "http://localhost:5173"
	} else {
		baseUrl = "https://status.pluralkit.me"
	}
	discordOauthConfig = &oauth2.Config{
		RedirectURL:  baseUrl + "/api/v1/auth/discord/callback",
		ClientID:     os.Getenv("DISCORD_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_OAUTH_CLIENT_SECRET"),
		Scopes:       []string{"identify"},
		Endpoint:     endpoints.Discord,
	}
}

const discordUserEndpoint = "https://discordapp.com/api/users/@me"

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	err := a.Sessions.Destroy(r.Context())
	if err != nil {
		a.Logger.Error("failed to destroy session", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "login_status",
		Value:    "0",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *API) oauthDiscordLogin(w http.ResponseWriter, r *http.Request) {
	state := a.genState(w)

	url := discordOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (a *API) oauthDiscordCallback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("oauthstate")
	if err != nil {
		a.Logger.Warn("error while retrieving oauth state", slog.Any("error", err))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if r.FormValue("state") != state.Value {
		a.Logger.Warn("invalid oauth state", slog.Any("error", err))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !a.Config.RunDev,
		SameSite: http.SameSiteLaxMode,
	})

	data, err := getUserData(r.Context(), r.FormValue("code"))
	if err != nil {
		a.Logger.Error("error while getting user data", slog.Any("error", err))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var user DiscordUser
	if err := json.Unmarshal(data, &user); err != nil {
		a.Logger.Error("error parsing json", slog.Any("error", err))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if !slices.Contains(a.Config.AuthorizedUsers, user.ID) {
		http.Error(w, "You are not authorized to access this resource", http.StatusUnauthorized)
		a.Logger.Info("Unauthorized user attempted access", slog.Any("user", user))
		return
	}

	if err := a.Sessions.RenewToken(r.Context()); err != nil {
		a.Logger.Error("error renewing token", slog.Any("error", err))
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	a.Sessions.Put(r.Context(), "user_session", user)
	http.SetCookie(w, &http.Cookie{
		Name:     "login_status",
		Value:    "1",
		Path:     "/",
		HttpOnly: false,
		Secure:   !a.Config.RunDev,
		MaxAge:   86400,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *API) genState(w http.ResponseWriter) string {
	var expiration = time.Now().Add(24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   !a.Config.RunDev,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	return state
}

func getUserData(ctx context.Context, code string) ([]byte, error) {
	token, err := discordOauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", discordUserEndpoint, nil)
	if err != nil {
		return nil, err
	}

	token.SetAuthHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}
