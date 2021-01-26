package endpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"utilserver/pkg/spotify"

	"github.com/gorilla/mux"
)

func clearCookie(w *http.ResponseWriter) {
	c := &http.Cookie{
		Name:    "storage",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),

		HttpOnly: true,
	}
	http.SetCookie(*w, c)
}

// Handler - spotify authentication routes handler
func Handler(spotifyAuthService spotify.AuthService) http.Handler {
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/spotify/login", login(spotifyAuthService)).Methods(http.MethodGet)
	api.HandleFunc("/spotify/callback", loginCallback(spotifyAuthService)).Methods(http.MethodGet)
	api.HandleFunc("/spotify/recently_played", getRecentlyPlayed(spotifyAuthService)).Methods(http.MethodGet)
	api.HandleFunc("/spotify/audio_features", getAudioFeatures(spotifyAuthService)).Methods(http.MethodGet)
	return r
}

func login(authService spotify.AuthService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		base, err := url.Parse(os.Getenv("SPOTIFY_LOGIN_ENDPOINT"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		email := r.URL.Query().Get("email")
		if email != "" {
			profile, profileErr := authService.Login(email)
			if profile == nil {
				goto Redirect
			}
			if profileErr != nil {
				http.Error(w, profileErr.Error(), http.StatusInternalServerError)
			}
			profileByteArr, marshallingErr := json.Marshal(profile)
			if marshallingErr != nil {
				http.Error(w, marshallingErr.Error(), http.StatusInternalServerError)
			}
			w.Write(profileByteArr)
			return
		}
	Redirect:
		parm := url.Values{}
		parm.Add("client_id", os.Getenv("CLIENT_ID"))
		parm.Add("scope", os.Getenv("SCOPES"))
		parm.Add("response_type", "code")
		parm.Add("redirect_uri", os.Getenv("REDIRECT_URL"))
		base.RawQuery = parm.Encode()

		http.Redirect(w, r, base.String(), http.StatusTemporaryRedirect)
		return
	}
}

func loginCallback(authService spotify.AuthService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		storedStateCookie, _ := r.Cookie(os.Getenv("SPOTIFY_LOGIN_STATE_KEY"))

		if state != "" || (storedStateCookie != nil && (state != storedStateCookie.Value)) {
			fmt.Printf("STATE MISMATCH")
		} else {
			clearCookie(&w)

			profile, err := authService.AuthCallback(code)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			profileByteArr, err := json.Marshal(profile)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(profileByteArr)
		}
	}
}

func getRecentlyPlayed(authService spotify.AuthService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")
		limit := r.URL.Query().Get("limit")
		before := r.URL.Query().Get("before")
		after := r.URL.Query().Get("after")

		i, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		timeBefore, err := time.Parse("2006-01-02", before)
		timeAfter, err := time.Parse("2006-01-02", after)
		recentlyPlayed, err := authService.GetRecentlyPlayed(
			email, i,
			strconv.FormatInt(timeBefore.UnixNano()/1000000, 10),
			strconv.FormatInt(timeAfter.UnixNano()/1000000, 10),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(*recentlyPlayed)
	}
}

func getAudioFeatures(authService spotify.AuthService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		trackIDs := r.URL.Query().Get("ids")
		email := r.URL.Query().Get("email")
		trackIDsArray := strings.Split(trackIDs, ",")
		resp, err := authService.GetTracksAudioFeatures(email, trackIDsArray)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(*resp)
	}
}
