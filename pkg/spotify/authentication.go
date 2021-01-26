package spotify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Login - login logic
func (service *Service) Login(email string) (*Profile, error) {
	profile, profileErr := service.repository.GetProfileWithEmail(email)
	return profile, profileErr
}

func (service *Service) GetCredentials(authorizationCode string) (*Credentials, error) {
	secretToken := base64.StdEncoding.EncodeToString([]byte(os.Getenv("CLIENT_ID") + ":" + os.Getenv("CLIENT_SECRET")))
	resp, err := service.httpClient.Request(
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
	var credentials Credentials
	json.Unmarshal(returnBody, &credentials)
	return &credentials, nil
}

func (service *Service) GetProfile(accessToken string) (*Profile, error) {
	profileResp, err := service.httpClient.Request(
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

	var profile Profile
	json.Unmarshal(profileResponse, &profile)
	return &profile, err
}

// AuthCallback - callback function when spotify hit the  authorization endpoint
func (service *Service) AuthCallback(authorizationCode string) (*Profile, error) {
	credentials, err := service.GetCredentials(authorizationCode)

	if err != nil {
		return nil, err
	}

	profile, err := service.GetProfile(credentials.AccessToken)
	if err != nil {
		return nil, err
	}

	profile.Credentials = *credentials
	_, createError := service.repository.CreateOrUpdateProfile(*profile)
	if createError != nil {
		return nil, createError
	}
	return profile, nil
}

// GetValidToken - return credentials with valid token meaning if token is expred, token will be refreshed
func (service *Service) GetValidToken(email string) (*Credentials, error) {
	profile, err := service.repository.GetProfileWithEmail(email)
	if err != nil {
		return nil, err
	}
	remaingTokenTime := 3600 - time.Now().Sub(profile.Credentials.UpdatedAt).Seconds()
	if remaingTokenTime <= 10 {
		refreshCredentials, err := service.RefreshToken(profile.Credentials.RefreshToken)
		if err != nil {
			return nil, err
		}
		_, updateErr := service.repository.UpdateCredentials(email, refreshCredentials)
		if updateErr != nil {
			return nil, updateErr
		}
		return refreshCredentials, nil
	}
	return &profile.Credentials, nil
}

func (service *Service) RefreshToken(refreshToken string) (*Credentials, error) {
	secretToken := base64.StdEncoding.EncodeToString([]byte(os.Getenv("CLIENT_ID") + ":" + os.Getenv("CLIENT_SECRET")))
	profileResp, err := service.httpClient.Request(
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

	var refreshTokenPayload Credentials
	json.Unmarshal(refreshTokenResp, &refreshTokenPayload)
	fmt.Println(string(refreshTokenResp))
	return &refreshTokenPayload, nil
}
