package domain

type Repository interface {
	CreateOrUpdateProfile(profile Profile) (*Profile, error)
	GetProfileWithEmail(email string) (*Profile, error)
	UpdateCredentials(email string, credentials *Credentials) (*Profile, error)
}

type HTTPClient interface {
	GetCredentials(authorizationCode string, secretToken string) (*Credentials, error)
	GetProfile(accessToken string) (*Profile, error)
	RefreshToken(refreshToken string, secretToken string) (*Credentials, error)
	GetRecentlyPlayed(accessToken string, limit int, before string, after string) (*[]byte, error)
	GetTracksAudioFeatures(accessToken string, trackIDs []string) (*[]byte, error)
	GetTopArtistsOrTracks(accessToken string, top string, timeRange string, limit int, offset int) (*[]byte, error)
}

// AuthService - functions implemented
type ProfileService interface {
	Login(email string) (*Profile, error)
	AuthCallback(authorizationCode string) (*Profile, error)
	GetRecentlyPlayed(email string, limit int, before string, after string) (*[]byte, error)
	GetTracksAudioFeatures(email string, trackIDs []string) (*[]byte, error)
	GetValidToken(email string) (*Credentials, error)
	GetTopArtistsOrTracks(email string, top string, timeRange string, limit int, offset int) (*[]byte, error)
}

type Service struct {
	repository Repository
	httpClient HTTPClient
}

// New - instantiate profile instance with api client and repository
func New(repository Repository, httpClient HTTPClient) ProfileService {
	return &Service{repository, httpClient}
}
