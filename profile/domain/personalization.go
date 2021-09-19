package domain

import (
	"errors"
)

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
		}
		return "medium_term", nil
	}
	return "", errors.New("incorrect format parameter")
}

func (service *Service) GetRecentlyPlayed(
	email string, limit int,
	before string, after string,
) (*[]byte, error) {
	credentials, err := service.GetValidToken(email)
	if err != nil {
		return nil, err
	}
	recentlyPlayed, err := service.httpClient.GetRecentlyPlayed(credentials.AccessToken, limit, before, after)
	return recentlyPlayed, err
}

func (service *Service) GetTopArtistsOrTracks(email string,
	top string,
	timeRangeStr string,
	limit int,
	offset int) (*[]byte, error) {
	credentials, err := service.GetValidToken(email)
	if err != nil {
		return nil, err
	}
	topContainer, err := service.httpClient.GetTopArtistsOrTracks(credentials.AccessToken, top, timeRangeStr, limit, offset)
	return topContainer, err
}
