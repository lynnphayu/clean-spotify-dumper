package domain

import (
	"encoding/base64"
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
	credentials, err := service.httpClient.GetCredentials(authorizationCode, secretToken)
	return credentials, err
}

func (service *Service) GetProfile(accessToken string) (*Profile, error) {
	profileResp, err := service.httpClient.GetProfile(accessToken)
	return profileResp, err
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
	remaingTokenTime := 3600 - time.Since(profile.Credentials.UpdatedAt).Seconds()
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
	refreshTokenPayload, err := service.httpClient.RefreshToken(refreshToken, secretToken)
	return refreshTokenPayload, err
}
