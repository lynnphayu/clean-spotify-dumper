package domain

func (service *Service) GetTracksAudioFeatures(email string, trackIDs []string) (*[]byte, error) {
	profile, err := service.repository.GetProfileWithEmail(email)
	if err != nil {
		return nil, err
	}
	features, err := service.httpClient.GetTracksAudioFeatures(profile.Credentials.AccessToken, trackIDs)
	return features, err
}
