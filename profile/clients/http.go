package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	profile "utilserver/profile/domain"
)

type HTTPClient struct {
	Timeout int
}

func New(timeout int) *HTTPClient {
	return &HTTPClient{Timeout: timeout}
}
func (client *HTTPClient) constructRequest(
	methodType string,
	URL string,
	body map[string]interface{},
	contentType string,
	auth string,
) (*http.Request, error) {
	var bodyByteArr string = ""
	if body != nil {
		if contentType == "application/x-www-form-urlencoded" {
			form := url.Values{}
			for k, v := range body {
				form.Add(k, v.(string))
			}
			bodyByteArr = form.Encode()
		} else {
			b, _ := json.Marshal(body)
			bodyByteArr = string(b)
		}
	}
	request, err := http.NewRequest(methodType, URL, strings.NewReader(bodyByteArr))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", auth)
	request.Header.Add("Content-Type", contentType)
	request.Header.Add("Content-Length", strconv.Itoa(len(bodyByteArr)))
	return request, nil
}

func (client *HTTPClient) getClient(sec int) *http.Client {
	timeout := time.Duration(time.Duration(sec) * time.Second)
	return &http.Client{
		Timeout: timeout,
	}
}

// Request - http request with parameters and return http response from endpoint
func (client *HTTPClient) Request(methodType string, URL string, body map[string]interface{}, contentType string, auth string) (*http.Response, error) {
	request, err := client.constructRequest(methodType, URL, body, contentType, auth)
	if err != nil {
		return nil, err
	}
	resp, err := client.getClient(client.Timeout).Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return resp, errors.New("Remote Server Error")
	}
	return resp, err
}

func (client *HTTPClient) GetCredentials(authorizationCode string, secretToken string) (*profile.Credentials, error) {
	resp, err := client.Request(
		"POST",
		os.Getenv("SPOTIFY_TOKEN_GENERATOR_ENTPOINT"),
		map[string]interface{}{
			"code":         authorizationCode,
			"redirect_uri": os.Getenv("REDIRECT_URL"),
			"grant_type":   "authorization_code",
		},
		"application/x-www-form-urlencoded",
		"Basic "+secretToken,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	returnBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(returnBody))
	var credentials profile.Credentials
	json.Unmarshal(returnBody, &credentials)
	return &credentials, nil
}

func (client *HTTPClient) GetProfile(accessToken string) (*profile.Profile, error) {
	profileResp, err := client.Request(
		"GET",
		os.Getenv("SPOTIFY_PROFILE_URL"), nil,
		"application/json",
		"Bearer "+accessToken,
	)

	if err != nil {
		return nil, err
	}
	defer profileResp.Body.Close()

	profileResponse, err := ioutil.ReadAll(profileResp.Body)
	if err != nil {
		return nil, err
	}

	var profile profile.Profile
	json.Unmarshal(profileResponse, &profile)
	return &profile, err
}

func (client *HTTPClient) RefreshToken(refreshToken string, secretToken string) (*profile.Credentials, error) {
	profileResp, err := client.Request(
		"POST",
		os.Getenv("SPOTIFY_TOKEN_GENERATOR_ENTPOINT"),
		map[string]interface{}{
			"refresh_token": refreshToken,
			"grant_type":    "refresh_token",
		},
		"application/x-www-form-urlencoded",
		"Basic "+secretToken,
	)

	if err != nil {
		return nil, err
	}
	defer profileResp.Body.Close()

	refreshTokenResp, err := ioutil.ReadAll(profileResp.Body)
	if err != nil {
		return nil, err
	}

	var refreshTokenPayload profile.Credentials
	json.Unmarshal(refreshTokenResp, &refreshTokenPayload)
	fmt.Println(string(refreshTokenResp))
	return &refreshTokenPayload, nil
}

func TopQueryValidator(e string, paramType string) (string, error) {
	topType := [...]string{"tracks", "artists"}
	timeRange := [...]string{"short_term", "medium_term", "long_term"}
	switch paramType {
	case "type":
		for _, v := range topType {
			if v == e {
				return e, nil
			}
		}
		return "tracks", nil
	case "time_range":
		for _, v := range timeRange {
			if v == e {
				return e, nil
			}
			return "medium_term", nil
		}
	}
	return "", errors.New("Incorrect Format Parameter")
}

func (client *HTTPClient) GetRecentlyPlayed(
	accessToken string, limit int,
	before string, after string,
) (*[]byte, error) {

	URL := "https://api.spotify.com/v1/me/player/recently-played"
	if limit != 0 {
		URL = URL + "?limit=" + strconv.Itoa(limit)
	}
	if before != "-6795364578871" {
		URL = URL + "&before=" + before
	}
	if after != "-6795364578871" {
		URL = URL + "&after=" + after
	}
	recentlyPlayedResp, err := client.Request(
		"GET",
		URL,
		nil,
		"application/json",
		"Bearer "+accessToken,
	)
	if err != nil {
		return nil, err
	}
	defer recentlyPlayedResp.Body.Close()

	recentlyPlayed, err := ioutil.ReadAll(recentlyPlayedResp.Body)
	if err != nil {
		return nil, err
	}
	return &recentlyPlayed, nil
}

func (client *HTTPClient) GetTopArtistsOrTracks(accessToken string,
	top string,
	timeRangeStr string,
	limit int,
	offset int) (*[]byte, error) {
	toptype, err := TopQueryValidator(top, "type")
	timeRange, err := TopQueryValidator(timeRangeStr, "time_range")
	if err != nil {
		return nil, err
	}
	URL := os.Getenv("SPOTIFY_PERSONAL_TOP") + "/" + toptype +
		"?limit=" + strconv.Itoa(limit) +
		"&offset=" + strconv.Itoa(offset) +
		"&time_range=" + timeRange
	topResp, err := client.Request(
		"GET",
		URL,
		nil,
		"application/json",
		"Bearer "+accessToken,
	)
	if err != nil {
		return nil, err
	}
	defer topResp.Body.Close()

	topContainer, err := ioutil.ReadAll(topResp.Body)
	if err != nil {
		return nil, err
	}
	return &topContainer, nil
}

func (client *HTTPClient) GetTracksAudioFeatures(accessToken string, trackIDs []string) (*[]byte, error) {
	URL := os.Getenv("SPOTIFY_AUDIO_FEATURES") + "?ids="
	for i, trackID := range trackIDs {
		URL = URL + trackID
		if i+1 != len(trackIDs) {
			URL = URL + ","
		}
	}
	container, err := client.Request(
		"GET",
		URL,
		nil,
		"application/json",
		"Bearer "+accessToken,
	)
	if err != nil {
		return nil, err
	}
	defer container.Body.Close()

	features, err := ioutil.ReadAll(container.Body)
	if err != nil {
		return nil, err
	}
	return &features, nil
}
