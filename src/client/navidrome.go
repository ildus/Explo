package client

import (
	"bytes"
	"encoding/json"
	"fmt"

	"explo/src/models"
	"explo/src/util"
)

type NavidromeLoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type NavidromeLoginResponse struct {
	ID            string `json:"id"`
	IsAdmin       bool   `json:"isAdmin"`
	LastFMApiKey  string `json:"lastFMApiKey"`
	Name          string `json:"name"`
	SubsonicSalt  string `json:"subsonicSalt"`
	SubsonicToken string `json:"subsonicToken"`
	Token         string `json:"token"`
	Username      string `json:"username"`
}

type Navidrome struct {
	Auth *map[string]string
	Sub  *Subsonic
}

type NavidromeTrack struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Rating int    `json:"rating"`
	Path   string `json:"path"`
}

func NewNavidrome(sub *Subsonic) *Navidrome {
	return &Navidrome{Auth: nil, Sub: sub}
}

func (c *Navidrome) GetAuth() error { // Generate salt and token
	payload := NavidromeLoginUser{
		Username: c.Sub.Cfg.Creds.User,
		Password: c.Sub.Cfg.Creds.Password,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %s", err.Error())
	}

	body, err := c.Sub.HttpClient.MakeRequest("POST",
		fmt.Sprintf("https://%s/auth/login", c.Sub.Cfg.URL),
		bytes.NewBuffer(payloadBytes), nil)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	var auth NavidromeLoginResponse
	err = util.ParseResp(body, &auth)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	c.Sub.Token = auth.SubsonicToken
	c.Sub.Salt = auth.SubsonicSalt
	c.Sub.Cfg.ClientID = auth.ID
	c.Sub.Cfg.Creds.APIKey = auth.Token

	return nil
}

func (c *Navidrome) AddHeader() error {
	cfg := &c.Sub.Cfg
	if cfg.Creds.Headers == nil {
		cfg.Creds.Headers = make(map[string]string)
		cfg.Creds.Headers["x-nd-client-unique-id"] = cfg.ClientID
	}

	if cfg.Creds.APIKey != "" {
		cfg.Creds.Headers["x-nd-authorization"] = fmt.Sprintf("Bearer %s", cfg.Creds.APIKey)
	}

	return nil
}

func (c *Navidrome) GetLibrary() error {
	return nil
}

func (c *Navidrome) AddLibrary() error {
	return nil
}

func (c *Navidrome) GetPlaylists() ([]*models.Playlist, error) {
	return c.Sub.GetPlaylists()
}

func (c *Navidrome) GetPlaylist(ID string) ([]*models.Track, error) {
	api := fmt.Sprintf("playlist/%s/tracks", ID)
	body, err := c.navidromeRequest(api)

	if err != nil {
		return nil, err
	}

	var tracks []NavidromeTrack
	if err = util.ParseResp(body, &tracks); err != nil {
		return nil, err
	}

	fmt.Printf("%s, %d tracks\n", ID, len(tracks))

	return nil, nil
}

func (c *Navidrome) SearchSongs(tracks []*models.Track) error {
	return c.Sub.SearchSongs(tracks)
}

func (c *Navidrome) RefreshLibrary() error {
	return c.Sub.RefreshLibrary()
}

func (c *Navidrome) CreatePlaylist(tracks []*models.Track) error {
	return c.Sub.CreatePlaylist(tracks)
}

func (c *Navidrome) SearchPlaylist() error {
	return c.Sub.SearchPlaylist()
}

func (c *Navidrome) UpdatePlaylist() error {
	return c.Sub.UpdatePlaylist()
}

func (c *Navidrome) DeletePlaylist() error {
	return c.Sub.DeletePlaylist()
}

func (c *Navidrome) navidromeRequest(apiUrl string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/api/%s", c.Sub.Cfg.URL, apiUrl)
	body, err := c.Sub.HttpClient.MakeRequest("GET", reqURL, nil, c.Sub.Cfg.Creds.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to make request %s", err.Error())
	}

	var checkResp FailedResp
	if err = util.ParseResp(body, &checkResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request %s", err.Error())
	} else if checkResp.SubsonicResponse.Status == "failed" {
		return nil, fmt.Errorf("%s", checkResp.SubsonicResponse.Error.Message)
	}
	return body, nil
}
