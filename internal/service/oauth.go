package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/teamsphere/server/internal/config"
	"github.com/teamsphere/server/internal/model"
)

type OAuthProvider struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

type oauthProviderConfig struct {
	Name         string
	Label        string
	Enabled      bool
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	IssuerURL    string
	HostedDomain string
}

type oauthProfile struct {
	Subject   string
	Email     string
	Username  string
	AvatarURL string
}

var (
	ErrOAuthDisabled        = errors.New("oauth disabled")
	ErrOAuthProviderMissing = errors.New("oauth provider disabled or not configured")
	ErrOAuthInvalidState    = errors.New("oauth state invalid")
	ErrOAuthExchangeFailed  = errors.New("oauth exchange failed")
	ErrOAuthProfileFailed   = errors.New("oauth profile failed")
)

func (s *AuthService) ListOAuthProviders(ctx context.Context) ([]OAuthProvider, error) {
	if !s.isOAuthEnabled(ctx) {
		return []OAuthProvider{}, nil
	}
	providers := []OAuthProvider{}
	if cfg := s.loadGitHubConfig(ctx); cfg.Enabled {
		providers = append(providers, OAuthProvider{Name: cfg.Name, Label: cfg.Label, Enabled: true})
	}
	if cfg := s.loadGoogleConfig(ctx); cfg.Enabled {
		providers = append(providers, OAuthProvider{Name: cfg.Name, Label: cfg.Label, Enabled: true})
	}
	if cfg := s.loadOIDCConfig(ctx); cfg.Enabled {
		providers = append(providers, OAuthProvider{Name: cfg.Name, Label: cfg.Label, Enabled: true})
	}
	return providers, nil
}

func (s *AuthService) GetOAuthFrontendBaseURL(ctx context.Context) string {
	return s.getSetting(ctx, "oauth.frontend_base_url")
}

func (s *AuthService) StartOAuth(ctx context.Context, provider, redirect string) (string, error) {
	if !s.isOAuthEnabled(ctx) {
		return "", ErrOAuthDisabled
	}
	cfg, err := s.loadProviderConfig(ctx, provider)
	if err != nil {
		return "", err
	}

	state := s.signOAuthState(provider, redirect)
	params := url.Values{}
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(cfg.Scopes, " "))
	params.Set("state", state)
	if cfg.HostedDomain != "" && provider == "google" {
		params.Set("hd", cfg.HostedDomain)
	}

	authURL := cfg.AuthURL + "?" + params.Encode()
	return authURL, nil
}

func (s *AuthService) HandleOAuthCallback(ctx context.Context, provider, code, state string, ipAddress, userAgent *string) (*AuthResult, string, string, string, error) {
	if !s.isOAuthEnabled(ctx) {
		return nil, "", "", "", ErrOAuthDisabled
	}
	cfg, err := s.loadProviderConfig(ctx, provider)
	if err != nil {
		return nil, "", "", "", err
	}
	redirect, ok := s.verifyOAuthState(provider, state)
	if !ok {
		return nil, "", "", "", ErrOAuthInvalidState
	}
	accessToken, err := s.exchangeCode(ctx, cfg, code)
	if err != nil {
		return nil, redirect, "", "", err
	}

	profile, err := s.fetchProfile(ctx, cfg, accessToken)
	if err != nil {
		return nil, redirect, "", "", err
	}

	user, err := s.upsertOAuthUser(ctx, cfg.Name, profile)
	if err != nil {
		return nil, redirect, "", "", err
	}

	if s.Is2FARequiredForRole(ctx, user.Role) {
		enabled, _ := s.GetTOTPStatus(ctx, user.ID)
		if !enabled {
			setupToken, err := s.generateTwoFAChallenge(user.ID, "setup")
			if err != nil {
				return nil, redirect, "", "", err
			}
			return nil, redirect, "", setupToken, nil
		}
	}

	if enabled, _ := s.GetTOTPStatus(ctx, user.ID); enabled {
		challenge, err := s.generateTwoFAChallenge(user.ID, "oauth")
		if err != nil {
			return nil, redirect, "", "", err
		}
		return nil, redirect, challenge, "", nil
	}

	access, err := s.generateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, redirect, "", "", err
	}
	refresh, err := s.generateRefreshToken(ctx, user.ID, ipAddress, userAgent, nil)
	if err != nil {
		return nil, redirect, "", "", err
	}

	return &AuthResult{Token: access, RefreshToken: refresh, User: *user}, redirect, "", "", nil
}

func (s *AuthService) upsertOAuthUser(ctx context.Context, provider string, profile *oauthProfile) (*model.User, error) {
	if profile.Subject == "" {
		return nil, ErrOAuthProfileFailed
	}
	if s.oauthRepo != nil {
		if existing, err := s.oauthRepo.GetByProviderSubject(ctx, provider, profile.Subject); err != nil {
			return nil, err
		} else if existing != nil {
			return s.userRepo.GetByID(ctx, existing.UserID)
		}
	}

	var user *model.User
	if profile.Email != "" {
		existingByEmail, err := s.userRepo.GetByEmail(ctx, profile.Email)
		if err != nil {
			return nil, err
		}
		if existingByEmail != nil {
			user = existingByEmail
		}
	}

	if user == nil {
		if !s.isRegisterAllowed(ctx) {
			return nil, ErrRegisterDisabled
		}
		username := s.ensureUniqueUsername(ctx, profile.Username, profile.Email)
		randomPass, _ := config.GenerateRandomHex(16)
		newUser, err := s.userRepo.Create(ctx, username, randomPass, "user", profile.Email)
		if err != nil {
			return nil, err
		}
		user = newUser
		if profile.Email != "" {
			_ = s.userRepo.SetEmailVerified(ctx, user.ID, profile.Email)
			if refreshed, err := s.userRepo.GetByID(ctx, user.ID); err == nil && refreshed != nil {
				user = refreshed
			}
		}
	}

	if s.oauthRepo != nil {
		if existing, _ := s.oauthRepo.GetByUserProvider(ctx, user.ID, provider); existing == nil {
			_, _ = s.oauthRepo.Create(ctx, &model.OAuthIdentity{
				UserID:   user.ID,
				Provider: provider,
				Subject:  profile.Subject,
				Email:    profile.Email,
			})
		}
	}

	return user, nil
}

func (s *AuthService) ensureUniqueUsername(ctx context.Context, preferred, email string) string {
	base := strings.TrimSpace(preferred)
	if base == "" && email != "" {
		base = strings.Split(email, "@")[0]
	}
	if base == "" {
		base = "user"
	}
	base = sanitizeUsername(base)
	if base == "" {
		base = "user"
	}
	if u, _ := s.userRepo.GetByUsername(ctx, base); u == nil {
		return base
	}
	for i := 0; i < 100; i++ {
		candidate := fmt.Sprintf("%s%d", base, 100+i)
		if u, _ := s.userRepo.GetByUsername(ctx, candidate); u == nil {
			return candidate
		}
	}
	return fmt.Sprintf("user%d", time.Now().Unix()%100000)
}

func sanitizeUsername(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) < 3 {
		return out
	}
	if len(out) > 32 {
		return out[:32]
	}
	return out
}

func (s *AuthService) isOAuthEnabled(ctx context.Context) bool {
	if s.settingsRepo == nil {
		return false
	}
	val, _ := s.settingsRepo.Get(ctx, "oauth.enabled")
	return val == "true"
}

func (s *AuthService) loadProviderConfig(ctx context.Context, provider string) (*oauthProviderConfig, error) {
	switch provider {
	case "github":
		cfg := s.loadGitHubConfig(ctx)
		if !cfg.Enabled || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" {
			return nil, ErrOAuthProviderMissing
		}
		return &cfg, nil
	case "google":
		cfg := s.loadGoogleConfig(ctx)
		if !cfg.Enabled || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" {
			return nil, ErrOAuthProviderMissing
		}
		return &cfg, nil
	case "oidc":
		cfg := s.loadOIDCConfig(ctx)
		if !cfg.Enabled || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" || cfg.IssuerURL == "" {
			return nil, ErrOAuthProviderMissing
		}
		return &cfg, nil
	default:
		return nil, ErrOAuthProviderMissing
	}
}

func (s *AuthService) loadGitHubConfig(ctx context.Context) oauthProviderConfig {
	return oauthProviderConfig{
		Name:         "github",
		Label:        "GitHub",
		Enabled:      s.getSettingBool(ctx, "oauth.github.enabled"),
		ClientID:     s.getSetting(ctx, "oauth.github.client_id"),
		ClientSecret: s.getSetting(ctx, "oauth.github.client_secret"),
		RedirectURL:  s.getSetting(ctx, "oauth.github.redirect_url"),
		Scopes:       s.getSettingSlice(ctx, "oauth.github.scopes", []string{"read:user", "user:email"}),
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
	}
}

func (s *AuthService) loadGoogleConfig(ctx context.Context) oauthProviderConfig {
	return oauthProviderConfig{
		Name:         "google",
		Label:        "Google",
		Enabled:      s.getSettingBool(ctx, "oauth.google.enabled"),
		ClientID:     s.getSetting(ctx, "oauth.google.client_id"),
		ClientSecret: s.getSetting(ctx, "oauth.google.client_secret"),
		RedirectURL:  s.getSetting(ctx, "oauth.google.redirect_url"),
		Scopes:       s.getSettingSlice(ctx, "oauth.google.scopes", []string{"openid", "email", "profile"}),
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		HostedDomain: s.getSetting(ctx, "oauth.google.hosted_domain"),
	}
}

func (s *AuthService) loadOIDCConfig(ctx context.Context) oauthProviderConfig {
	return oauthProviderConfig{
		Name:         "oidc",
		Label:        "OIDC",
		Enabled:      s.getSettingBool(ctx, "oauth.oidc.enabled"),
		ClientID:     s.getSetting(ctx, "oauth.oidc.client_id"),
		ClientSecret: s.getSetting(ctx, "oauth.oidc.client_secret"),
		RedirectURL:  s.getSetting(ctx, "oauth.oidc.redirect_url"),
		Scopes:       s.getSettingSlice(ctx, "oauth.oidc.scopes", []string{"openid", "email", "profile"}),
		IssuerURL:    s.getSetting(ctx, "oauth.oidc.issuer_url"),
	}
}

func (s *AuthService) getSetting(ctx context.Context, key string) string {
	if s.settingsRepo == nil {
		return ""
	}
	val, _ := s.settingsRepo.Get(ctx, key)
	return strings.TrimSpace(val)
}

func (s *AuthService) getSettingBool(ctx context.Context, key string) bool {
	return s.getSetting(ctx, key) == "true"
}

func (s *AuthService) getSettingSlice(ctx context.Context, key string, def []string) []string {
	raw := s.getSetting(ctx, key)
	if raw == "" {
		return def
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}

func (s *AuthService) signOAuthState(provider, redirect string) string {
	exp := time.Now().Add(10 * time.Minute).Unix()
	payload := map[string]any{
		"provider": provider,
		"redirect": redirect,
		"exp":      exp,
	}
	raw, _ := json.Marshal(payload)
	b64 := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(s.jwtCfg.Secret))
	mac.Write([]byte(b64))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return b64 + "." + sig
}

func (s *AuthService) verifyOAuthState(provider, token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", false
	}
	b64, sig := parts[0], parts[1]
	mac := hmac.New(sha256.New, []byte(s.jwtCfg.Secret))
	mac.Write([]byte(b64))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil {
		return "", false
	}
	var payload struct {
		Provider string `json:"provider"`
		Redirect string `json:"redirect"`
		Exp      int64  `json:"exp"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", false
	}
	if payload.Provider != provider {
		return "", false
	}
	if time.Now().Unix() > payload.Exp {
		return "", false
	}
	return payload.Redirect, true
}

func (s *AuthService) exchangeCode(ctx context.Context, cfg *oauthProviderConfig, code string) (string, error) {
	if cfg.Name == "github" {
		values := url.Values{}
		values.Set("client_id", cfg.ClientID)
		values.Set("client_secret", cfg.ClientSecret)
		values.Set("code", code)
		values.Set("redirect_uri", cfg.RedirectURL)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(values.Encode()))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", ErrOAuthExchangeFailed
		}
		defer resp.Body.Close()
		var data struct {
			AccessToken string `json:"access_token"`
		}
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &data); err != nil || data.AccessToken == "" {
			return "", ErrOAuthExchangeFailed
		}
		return data.AccessToken, nil
	}

	if cfg.Name == "oidc" {
		if err := s.resolveOIDCEndpoints(ctx, cfg); err != nil {
			return "", err
		}
	}

	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("client_secret", cfg.ClientSecret)
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", cfg.RedirectURL)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", ErrOAuthExchangeFailed
	}
	defer resp.Body.Close()
	var data struct {
		AccessToken string `json:"access_token"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &data); err != nil || data.AccessToken == "" {
		return "", ErrOAuthExchangeFailed
	}
	return data.AccessToken, nil
}

func (s *AuthService) resolveOIDCEndpoints(ctx context.Context, cfg *oauthProviderConfig) error {
	if cfg.IssuerURL == "" {
		return ErrOAuthProviderMissing
	}
	issuer := strings.TrimSuffix(cfg.IssuerURL, "/")
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, issuer+"/.well-known/openid-configuration", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ErrOAuthExchangeFailed
	}
	defer resp.Body.Close()
	var data struct {
		AuthURL     string `json:"authorization_endpoint"`
		TokenURL    string `json:"token_endpoint"`
		UserInfoURL string `json:"userinfo_endpoint"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &data); err != nil {
		return ErrOAuthExchangeFailed
	}
	if data.AuthURL != "" {
		cfg.AuthURL = data.AuthURL
	}
	if data.TokenURL != "" {
		cfg.TokenURL = data.TokenURL
	}
	if data.UserInfoURL != "" {
		cfg.UserInfoURL = data.UserInfoURL
	}
	return nil
}

func (s *AuthService) fetchProfile(ctx context.Context, cfg *oauthProviderConfig, accessToken string) (*oauthProfile, error) {
	switch cfg.Name {
	case "github":
		return s.fetchGitHubProfile(ctx, accessToken)
	case "google":
		return s.fetchGoogleProfile(ctx, accessToken)
	case "oidc":
		return s.fetchOIDCProfile(ctx, cfg, accessToken)
	default:
		return nil, ErrOAuthProfileFailed
	}
}

func (s *AuthService) fetchGitHubProfile(ctx context.Context, accessToken string) (*oauthProfile, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrOAuthProfileFailed
	}
	defer resp.Body.Close()
	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, ErrOAuthProfileFailed
	}

	email := user.Email
	if email == "" {
		req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
		req2.Header.Set("Authorization", "token "+accessToken)
		req2.Header.Set("Accept", "application/vnd.github+json")
		resp2, err := http.DefaultClient.Do(req2)
		if err == nil {
			defer resp2.Body.Close()
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if data, _ := io.ReadAll(resp2.Body); json.Unmarshal(data, &emails) == nil {
				for _, e := range emails {
					if e.Primary && e.Verified {
						email = e.Email
						break
					}
				}
				if email == "" && len(emails) > 0 {
					email = emails[0].Email
				}
			}
		}
	}

	return &oauthProfile{
		Subject:   fmt.Sprintf("%d", user.ID),
		Email:     email,
		Username:  user.Login,
		AvatarURL: user.AvatarURL,
	}, nil
}

func (s *AuthService) fetchGoogleProfile(ctx context.Context, accessToken string) (*oauthProfile, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrOAuthProfileFailed
	}
	defer resp.Body.Close()
	var user struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Picture  string `json:"picture"`
		Verified bool   `json:"verified_email"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, ErrOAuthProfileFailed
	}
	email := ""
	if user.Verified {
		email = user.Email
	}
	return &oauthProfile{
		Subject:   user.ID,
		Email:     email,
		Username:  user.Name,
		AvatarURL: user.Picture,
	}, nil
}

func (s *AuthService) fetchOIDCProfile(ctx context.Context, cfg *oauthProviderConfig, accessToken string) (*oauthProfile, error) {
	if cfg.UserInfoURL == "" {
		if err := s.resolveOIDCEndpoints(ctx, cfg); err != nil {
			return nil, ErrOAuthProfileFailed
		}
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, ErrOAuthProfileFailed
	}
	defer resp.Body.Close()
	var user struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, ErrOAuthProfileFailed
	}
	return &oauthProfile{
		Subject:   user.Sub,
		Email:     user.Email,
		Username:  user.Name,
		AvatarURL: user.Picture,
	}, nil
}
