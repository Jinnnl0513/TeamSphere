package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/teamsphere/server/internal/config"
	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken        = errors.New("username already exists")
	ErrIDTaken              = errors.New("id already taken")
	ErrEmailTaken           = errors.New("email already registered")
	ErrInvalidCredentials   = errors.New("invalid username or password")
	ErrLoginLocked          = errors.New("login temporarily locked")
	ErrWeakPassword         = errors.New("password does not meet complexity requirements")
	ErrRegisterDisabled     = errors.New("registration is disabled")
	ErrCodeTooFrequent      = errors.New("please wait before requesting another code")
	ErrCodeNotFound         = errors.New("no valid verification code found")
	ErrCodeExpired          = errors.New("verification code has expired")
	ErrCodeInvalid          = errors.New("invalid verification code")
	ErrCodeTooManyRetries   = errors.New("too many attempts, please request a new code")
	ErrEmailNotVerified     = errors.New("email verification required")
	ErrInvalidVerifyToken   = errors.New("invalid or expired verification token")
	ErrRefreshTokenInvalid  = errors.New("invalid or expired refresh token")
	ErrRefreshTokenRevoked  = errors.New("refresh token has been revoked")
	ErrTOTPRequired         = errors.New("totp required")
	ErrTOTPInvalid          = errors.New("totp invalid")
	ErrTOTPNotSetup         = errors.New("totp not setup")
	ErrTOTPChallengeInvalid = errors.New("totp challenge invalid")
	ErrTOTPPolicyRequired   = errors.New("totp required by policy")
)

const (
	verificationCodeLength   = 6
	verificationCodeExpiry   = 5 * time.Minute
	verificationCodeCooldown = 60 * time.Second
	verificationMaxAttempts  = 5
	verificationTokenExpiry  = 10 * time.Minute
)

type AuthService struct {
	userRepo         repository.UserRepository
	oauthRepo        repository.OAuthIdentityRepository
	totpRepo         repository.UserTOTPSecretRepository
	recoveryRepo     repository.RecoveryCodeRepository
	settingsRepo     repository.SettingsRepository
	emailService     *EmailService
	jwtCfg           *config.JWTConfig
	loginAttemptRepo repository.LoginAttemptRepository
	encryptionKey    string
}

func NewAuthService(userRepo repository.UserRepository, oauthRepo repository.OAuthIdentityRepository, totpRepo repository.UserTOTPSecretRepository, recoveryRepo repository.RecoveryCodeRepository, settingsRepo repository.SettingsRepository, jwtCfg *config.JWTConfig, loginAttemptRepo repository.LoginAttemptRepository, encryptionKey string) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		oauthRepo:        oauthRepo,
		totpRepo:         totpRepo,
		recoveryRepo:     recoveryRepo,
		settingsRepo:     settingsRepo,
		jwtCfg:           jwtCfg,
		loginAttemptRepo: loginAttemptRepo,
		encryptionKey:    encryptionKey,
	}
}

func (s *AuthService) SetEmailDeps(settingsRepo repository.SettingsRepository, emailService *EmailService) {
	s.settingsRepo = settingsRepo
	s.emailService = emailService
}

type AuthClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type TwoFAChallengeClaims struct {
	UserID  int64  `json:"user_id"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

type AuthResult struct {
	Token        string     `json:"token"`
	RefreshToken string     `json:"refresh_token"`
	User         model.User `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TOTPSetupRequiredError indicates the user must complete 2FA setup before login can succeed.
type TOTPSetupRequiredError struct {
	Challenge string
}

func (e *TOTPSetupRequiredError) Error() string {
	return "totp setup required"
}

func (s *AuthService) Register(ctx context.Context, username, password, email, verificationToken string, ipAddress, userAgent *string) (*AuthResult, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if !s.isRegisterAllowed(ctx) {
		return nil, ErrRegisterDisabled
	}

	existing, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("check username: %w", err)
	}
	if existing != nil {
		return nil, ErrUsernameTaken
	}

	emailRequired := s.isEmailRequired(ctx)
	verifiedEmail := ""
	if email != "" {
		existingByEmail, err := s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("check email: %w", err)
		}
		if existingByEmail != nil {
			return nil, ErrEmailTaken
		}

		if verificationToken == "" {
			if emailRequired {
				return nil, ErrEmailNotVerified
			}
		} else {
			if !s.validateVerificationToken(email, verificationToken) {
				return nil, ErrInvalidVerifyToken
			}
			verifiedEmail = email
		}
	} else if emailRequired {
		return nil, ErrEmailNotVerified
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, username, string(hash), "user", verifiedEmail)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if verifiedEmail != "" {
		if err := s.userRepo.SetEmailVerified(ctx, user.ID, verifiedEmail); err != nil {
			slog.Error("failed to set email verified", "error", err)
		}
		if refreshedUser, err := s.userRepo.GetByID(ctx, user.ID); err != nil {
			slog.Error("failed to refresh user after email verification", "error", err)
		} else if refreshedUser != nil {
			user = refreshedUser
		}
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.generateRefreshToken(ctx, user.ID, ipAddress, userAgent, nil)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: accessToken, RefreshToken: refreshToken, User: *user}, nil
}

func (s *AuthService) Login(ctx context.Context, username, password, totpCode, recoveryCode, ip string, userAgent *string) (*AuthResult, error) {
	if s.loginAttemptRepo != nil {
		if locked, err := s.isLoginLocked(ctx, username, ip); err != nil {
			return nil, err
		} else if locked {
			return nil, ErrLoginLocked
		}
	}
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		if s.loginAttemptRepo != nil {
			_ = s.recordLoginFailure(ctx, username, ip)
		}
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if s.loginAttemptRepo != nil {
			_ = s.recordLoginFailure(ctx, username, ip)
		}
		return nil, ErrInvalidCredentials
	}

	if s.Is2FARequiredForRole(ctx, user.Role) {
		enabled, _ := s.GetTOTPStatus(ctx, user.ID)
		if !enabled {
			challenge, err := s.generateTwoFAChallenge(user.ID, "setup")
			if err != nil {
				return nil, err
			}
			return nil, &TOTPSetupRequiredError{Challenge: challenge}
		}
	}

	if err := s.ValidateTOTPForUser(ctx, user.ID, totpCode, recoveryCode); err != nil {
		return nil, err
	}

	if s.loginAttemptRepo != nil {
		_ = s.loginAttemptRepo.Reset(ctx, loginAttemptKey(username, ip))
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	ipPtr := &ip
	refreshToken, err := s.generateRefreshToken(ctx, user.ID, ipPtr, userAgent, nil)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: accessToken, RefreshToken: refreshToken, User: *user}, nil
}

func (s *AuthService) Logout(ctx context.Context, claims *AuthClaims, refreshTokenStr string) error {
	entry := &model.TokenBlacklist{
		TokenJTI:  claims.ID,
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt.Time,
	}
	if err := s.userRepo.BlacklistToken(ctx, entry); err != nil {
		return err
	}

	// Revoke refresh token if provided
	if refreshTokenStr != "" {
		hash := sha256.Sum256([]byte(refreshTokenStr))
		tokenHash := hex.EncodeToString(hash[:])
		if err := s.userRepo.RevokeRefreshToken(ctx, tokenHash); err != nil {
			slog.Error("failed to revoke refresh token", "error", err)
		}
	}

	return nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (*AuthClaims, error) {
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	blacklisted, err := s.userRepo.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("check blacklist: %w", err)
	}
	if blacklisted {
		return nil, errors.New("token has been revoked")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("get token user: %w", err)
	}
	if user == nil {
		return nil, errors.New("token user no longer exists")
	}

	claims.Username = user.Username
	claims.Role = user.Role

	return claims, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*AuthResult, error) {
	hash := sha256.Sum256([]byte(refreshTokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	rt, err := s.userRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	if rt == nil {
		return nil, ErrRefreshTokenInvalid
	}

	if rt.RevokedAt != nil {
		// Reuse detected: a revoked token is being used again.
		// Revoke ALL tokens for this user to protect the token family.
		slog.Warn("refresh token reuse detected, revoking all tokens", "user_id", rt.UserID)
		if err := s.userRepo.RevokeAllRefreshTokensForUser(ctx, rt.UserID); err != nil {
			slog.Error("failed to revoke all refresh tokens after reuse", "error", err)
		}
		return nil, ErrRefreshTokenRevoked
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, ErrRefreshTokenInvalid
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil || user.DeletedAt != nil {
		return nil, ErrRefreshTokenInvalid
	}

	if err := s.userRepo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		slog.Error("failed to revoke old refresh token", "error", err)
	}

	if err := s.userRepo.UpdateRefreshTokenLastUsed(ctx, tokenHash); err != nil {
		slog.Error("failed to update refresh token last used", "error", err)
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := s.generateRefreshToken(ctx, user.ID, rt.IPAddress, rt.UserAgent, rt.DeviceName)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: accessToken, RefreshToken: newRefreshToken, User: *user}, nil
}

func (s *AuthService) VerifyTwoFALogin(ctx context.Context, challenge, totpCode, recoveryCode string, ipAddress, userAgent *string) (*AuthResult, error) {
	userID, err := s.verifyTwoFAChallenge(challenge, "oauth")
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if err := s.ValidateTOTPForUser(ctx, user.ID, totpCode, recoveryCode); err != nil {
		return nil, err
	}
	accessToken, err := s.generateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.generateRefreshToken(ctx, user.ID, ipAddress, userAgent, nil)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: accessToken, RefreshToken: refreshToken, User: *user}, nil
}

func (s *AuthService) generateAccessToken(userID int64, username, role string) (string, error) {
	now := time.Now()
	expireMinutes := s.jwtCfg.AccessExpireMinutes
	if expireMinutes <= 0 {
		expireMinutes = 15
	}
	claims := AuthClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireMinutes) * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// GenerateAccessTokenForUser is a public wrapper for generating access tokens.
func (s *AuthService) GenerateAccessTokenForUser(user *model.User) (string, error) {
	if user == nil {
		return "", ErrUserNotFound
	}
	return s.generateAccessToken(user.ID, user.Username, user.Role)
}

func (s *AuthService) generateTwoFAChallenge(userID int64, purpose string) (string, error) {
	now := time.Now()
	claims := TwoFAChallengeClaims{
		UserID:  userID,
		Purpose: purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}

func (s *AuthService) verifyTwoFAChallenge(tokenStr, purpose string) (int64, error) {
	claims := &TwoFAChallengeClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return 0, ErrTOTPChallengeInvalid
	}
	if claims.Purpose != purpose {
		return 0, ErrTOTPChallengeInvalid
	}
	return claims.UserID, nil
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID int64, ipAddress, userAgent, deviceName *string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(tokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	expireDays := s.jwtCfg.RefreshExpireDays
	if expireDays <= 0 {
		expireDays = 7
	}
	expiresAt := time.Now().AddDate(0, 0, expireDays)

	if _, err := s.userRepo.CreateRefreshToken(ctx, userID, tokenHash, expiresAt, ipAddress, userAgent, deviceName); err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return tokenStr, nil
}

// GenerateRefreshTokenForUser is a public wrapper for generating refresh tokens.
func (s *AuthService) GenerateRefreshTokenForUser(ctx context.Context, userID int64, ipAddress, userAgent, deviceName *string) (string, error) {
	return s.generateRefreshToken(ctx, userID, ipAddress, userAgent, deviceName)
}

func (s *AuthService) SendVerificationCode(ctx context.Context, email string) error {
	if s.emailService == nil {
		return fmt.Errorf("email service not configured")
	}

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return ErrEmailTaken
	}

	latest, err := s.userRepo.GetLatestVerification(ctx, email)
	if err != nil {
		return fmt.Errorf("get latest verification: %w", err)
	}
	if latest != nil && time.Since(latest.CreatedAt) < verificationCodeCooldown {
		return ErrCodeTooFrequent
	}

	code, err := generateNumericCode(verificationCodeLength)
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	expiresAt := time.Now().Add(verificationCodeExpiry)
	if _, err := s.userRepo.CreateEmailVerification(ctx, email, code, expiresAt); err != nil {
		return fmt.Errorf("store verification: %w", err)
	}

	subject := "TeamSphere verification code"
	body := fmt.Sprintf("Your verification code is: %s\n\nThis code will expire in 5 minutes. If you did not request it, you can ignore this email.", code)
	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := s.emailService.SendEmail(sendCtx, email, subject, body); err != nil {
		slog.Error("failed to send verification email", "email", email, "error", err)
		return err
	}

	slog.Info("verification email sent", "email", email)
	return nil
}

func (s *AuthService) VerifyEmailCode(ctx context.Context, email, code string) (string, error) {
	v, err := s.userRepo.GetLatestVerification(ctx, email)
	if err != nil {
		return "", fmt.Errorf("get verification: %w", err)
	}
	if v == nil {
		return "", ErrCodeNotFound
	}

	if v.Attempts >= verificationMaxAttempts {
		return "", ErrCodeTooManyRetries
	}

	attempts, err := s.userRepo.IncrementVerificationAttempts(ctx, v.ID)
	if err != nil {
		return "", fmt.Errorf("increment attempts: %w", err)
	}

	if v.Code != code {
		if attempts >= verificationMaxAttempts {
			return "", ErrCodeTooManyRetries
		}
		return "", ErrCodeInvalid
	}

	if err := s.userRepo.MarkVerificationUsed(ctx, v.ID); err != nil {
		return "", fmt.Errorf("mark used: %w", err)
	}

	token := s.generateVerificationToken(email)
	return token, nil
}

func (s *AuthService) SendPasswordResetCode(ctx context.Context, email string) error {
	if s.emailService == nil {
		return fmt.Errorf("email service not configured")
	}
	if !s.IsEmailEnabled(ctx) {
		return ErrEmailDisabled
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("check email: %w", err)
	}
	if user == nil {
		return nil
	}

	latest, err := s.userRepo.GetLatestVerification(ctx, email)
	if err != nil {
		return fmt.Errorf("get latest verification: %w", err)
	}
	if latest != nil && time.Since(latest.CreatedAt) < verificationCodeCooldown {
		return ErrCodeTooFrequent
	}

	code, err := generateNumericCode(verificationCodeLength)
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	expiresAt := time.Now().Add(verificationCodeExpiry)
	if _, err := s.userRepo.CreateEmailVerification(ctx, email, code, expiresAt); err != nil {
		return fmt.Errorf("store verification: %w", err)
	}

	subject := "TeamSphere password reset code"
	body := fmt.Sprintf("You requested a TeamSphere password reset. Your verification code is: %s\n\nThis code will expire in 5 minutes. If you did not request it, you can ignore this email.", code)
	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := s.emailService.SendEmail(sendCtx, email, subject, body); err != nil {
		slog.Error("failed to send password reset email", "email", email, "error", err)
		return err
	}

	slog.Info("password reset email sent", "email", email)
	return nil
}

func (s *AuthService) ResetPasswordByCode(ctx context.Context, email, code, newPassword string) error {
	if !ValidatePassword(newPassword) {
		return ErrWeakPassword
	}
	v, err := s.userRepo.GetLatestVerification(ctx, email)
	if err != nil {
		return fmt.Errorf("get verification: %w", err)
	}
	if v == nil {
		return ErrCodeNotFound
	}

	if v.Attempts >= verificationMaxAttempts {
		return ErrCodeTooManyRetries
	}

	attempts, err := s.userRepo.IncrementVerificationAttempts(ctx, v.ID)
	if err != nil {
		return fmt.Errorf("increment attempts: %w", err)
	}

	if v.Code != code {
		if attempts >= verificationMaxAttempts {
			return ErrCodeTooManyRetries
		}
		return ErrCodeInvalid
	}

	if err := s.userRepo.MarkVerificationUsed(ctx, v.ID); err != nil {
		return fmt.Errorf("mark used: %w", err)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("get user by email: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, string(hash)); err != nil {
		return err
	}
	if err := s.userRepo.RevokeAllRefreshTokensForUser(ctx, user.ID); err != nil {
		slog.Error("failed to revoke refresh tokens after password reset", "user_id", user.ID, "error", err)
	}

	return nil
}

func (s *AuthService) IsEmailRequired(ctx context.Context) bool {
	return s.isEmailRequired(ctx)
}

func (s *AuthService) IsRegisterAllowed(ctx context.Context) bool {
	return s.isRegisterAllowed(ctx)
}

func (s *AuthService) isEmailRequired(ctx context.Context) bool {
	if s.settingsRepo == nil {
		return false
	}
	val, err := s.settingsRepo.Get(ctx, "registration.email_required")
	if err != nil {
		slog.Error("failed to read email_required setting, defaulting to true", "error", err)
		return true
	}
	return val == "true"
}

func (s *AuthService) isRegisterAllowed(ctx context.Context) bool {
	if s.settingsRepo == nil {
		return true
	}
	val, err := s.settingsRepo.Get(ctx, "registration.allow_register")
	if err != nil {
		slog.Error("failed to read allow_register setting, defaulting to false", "error", err)
		return false
	}
	if val == "" {
		return true
	}
	return val == "true"
}

func (s *AuthService) IsEmailEnabled(ctx context.Context) bool {
	if s.settingsRepo == nil {
		return false
	}
	if s.isEmailRequired(ctx) {
		return true
	}
	val, _ := s.settingsRepo.Get(ctx, "email.enabled")
	return val == "true"
}

func (s *AuthService) Get2FAPolicy(ctx context.Context) string {
	if s.settingsRepo == nil {
		return "optional"
	}
	val, _ := s.settingsRepo.Get(ctx, "security.2fa_policy")
	val = strings.ToLower(strings.TrimSpace(val))
	switch val {
	case "off", "optional", "admins", "required":
		return val
	default:
		return "optional"
	}
}

func (s *AuthService) Is2FARequiredForRole(ctx context.Context, role string) bool {
	policy := s.Get2FAPolicy(ctx)
	switch policy {
	case "required":
		return true
	case "admins":
		return role == "admin" || role == "owner" || role == "system_admin"
	default:
		return false
	}
}



func (s *AuthService) generateVerificationToken(email string) string {
	expiry := time.Now().Add(verificationTokenExpiry).Unix()
	data := fmt.Sprintf("%s:%d", email, expiry)
	mac := hmac.New(sha256.New, []byte(s.jwtCfg.Secret))
	mac.Write([]byte(data))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s:%d:%s", email, expiry, sig)
}

func (s *AuthService) validateVerificationToken(email, token string) bool {
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return false
	}
	tokenEmail := parts[0]
	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}
	sig := parts[2]
	if tokenEmail != email {
		return false
	}
	if time.Now().Unix() > expiry {
		return false
	}
	data := fmt.Sprintf("%s:%d", tokenEmail, expiry)
	mac := hmac.New(sha256.New, []byte(s.jwtCfg.Secret))
	mac.Write([]byte(data))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

func (s *AuthService) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	return s.userRepo.RevokeAllRefreshTokensForUser(ctx, userID)
}

func (s *AuthService) ListSessions(ctx context.Context, userID int64) ([]*model.RefreshToken, error) {
	return s.userRepo.ListRefreshTokens(ctx, userID)
}

func (s *AuthService) RevokeSession(ctx context.Context, userID, sessionID int64) error {
	return s.userRepo.RevokeRefreshTokenByID(ctx, userID, sessionID)
}

func (s *AuthService) RevokeOtherSessions(ctx context.Context, userID int64, currentRefreshToken string) error {
	if currentRefreshToken == "" {
		return ErrRefreshTokenInvalid
	}
	hash := sha256.Sum256([]byte(currentRefreshToken))
	tokenHash := hex.EncodeToString(hash[:])
	return s.userRepo.RevokeOtherRefreshTokens(ctx, userID, tokenHash)
}

func (s *AuthService) StartBlacklistCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("token blacklist cleanup stopped")
				return
			case <-ticker.C:
				count, err := s.userRepo.CleanExpiredTokens(ctx)
				if err != nil {
					slog.Error("failed to clean expired tokens", "error", err)
				} else if count > 0 {
					slog.Info("cleaned expired tokens", "count", count)
				}
				vCount, err := s.userRepo.CleanExpiredVerifications(ctx)
				if err != nil {
					slog.Error("failed to clean expired verifications", "error", err)
				} else if vCount > 0 {
					slog.Info("cleaned expired verifications", "count", vCount)
				}
				rtCount, err := s.userRepo.CleanExpiredRefreshTokens(ctx)
				if err != nil {
					slog.Error("failed to clean expired refresh tokens", "error", err)
				} else if rtCount > 0 {
					slog.Info("cleaned expired refresh tokens", "count", rtCount)
				}
			}
		}
	}()
}

func generateNumericCode(length int) (string, error) {
	digits := make([]byte, length)
	for i := range digits {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0') + byte(n.Int64())
	}
	return string(digits), nil
}

const (
	loginMaxFailures  = 5
	loginLockDuration = 15 * time.Minute
)

func loginAttemptKey(username, ip string) string {
	username = strings.ToLower(strings.TrimSpace(username))
	return fmt.Sprintf("%s|%s", username, ip)
}

func (s *AuthService) isLoginLocked(ctx context.Context, username, ip string) (bool, error) {
	if s.loginAttemptRepo == nil {
		return false, nil
	}
	key := loginAttemptKey(username, ip)
	la, err := s.loginAttemptRepo.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if la == nil || la.LockedUntil == nil {
		return false, nil
	}
	return time.Now().Before(*la.LockedUntil), nil
}

func (s *AuthService) recordLoginFailure(ctx context.Context, username, ip string) error {
	key := loginAttemptKey(username, ip)
	la, err := s.loginAttemptRepo.Get(ctx, key)
	if err != nil {
		return err
	}
	attempts := 1
	if la != nil {
		attempts = la.Attempts + 1
	}
	var lockedUntil *time.Time
	if attempts >= loginMaxFailures {
		t := time.Now().Add(loginLockDuration)
		lockedUntil = &t
	}
	return s.loginAttemptRepo.Upsert(ctx, key, attempts, lockedUntil)
}
