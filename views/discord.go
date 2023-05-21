package views

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/xremming/abborre/esox"
	"github.com/xremming/abborre/esox/csrf"
	"golang.org/x/oauth2"
)

func DiscordLogin(oauth2Config oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		csrfStruct := csrf.FromContext(r.Context())

		state := csrfStruct.Generate()
		http.Redirect(
			w, r,
			oauth2Config.AuthCodeURL(state),
			http.StatusFound,
		)
	}
}

type DiscordUser struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
	MFA      bool    `json:"mfa_enabled"`
	Email    string  `json:"email"`
	Verified bool    `json:"verified"`
}

var loginFailureTmpl = esox.GetTemplate("login_failure.html", "base.html")

func DiscordCallback(oauth2Config oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := zerolog.Ctx(r.Context())
		csrfStruct := csrf.FromContext(r.Context())

		state := r.URL.Query().Get("state")
		err := csrfStruct.Validate(r.Context(), state)
		if err != nil {
			log.Err(err).Msg("failed to validate oauth2 state")
			renderError(w, r, 400, "Login failed, please try again.")
			return
		}

		code := r.URL.Query().Get("code")
		token, err := oauth2Config.Exchange(r.Context(), code)
		if err != nil {
			log.Err(err).Msg("failed to exchange oauth2 code")
			renderError(w, r, 500, "Login failed, please try again.")
			return
		}

		client := oauth2Config.Client(r.Context(), token)
		resp, err := client.Get("https://discord.com/api/v10/users/@me")
		if err != nil {
			log.Err(err).Msg("failed to get user info from discord")
			renderError(w, r, 500, "Login failed, please try again.")
			return
		}
		defer resp.Body.Close()

		var user DiscordUser
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			log.Err(err).Msg("failed to decode user info from discord")
			renderError(w, r, 500, "Login failed, please try again.")
			return
		}

		logWithDiscordUser := log.With().
			Str("user_id", user.ID).
			Str("username", user.Username).
			Logger()

		if !user.Verified {
			logWithDiscordUser.Warn().Msg("user is not verified")
			loginFailureTmpl.Render(w, r, 400, &data{
				Data: "You must verify your Discord email address before you can log in.",
			})
			return
		}

		// TODO: create or update user in database
		// TODO: create session and save the session to cookie

		logWithDiscordUser.Info().Msg("authentication succeeded")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
