package spotify

import (
	"io/ioutil"
	"strconv"
)

func (service *Service) GetRecentlyPlayed(
	email string, limit int,
	before string, after string,
) (*[]byte, error) {
	credentials, err := service.GetValidToken(email)
	if err != nil {
		return nil, err
	}
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
	recentlyPlayedResp, err := service.httpClient.Request(
		"GET",
		URL,
		nil,
		"application/json",
		"Bearer "+credentials.AccessToken,
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
