package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/teamsphere/server/internal/config"
)

const (
	totpDigits = 6
	totpPeriod = 30
	totpIssuer = "TeamSphere"
)

type TOTPSetupResult struct {
	Secret     string `json:"secret"`
	OtpAuthURL string `json:"otpauth_url"`
}

func (s *AuthService) GetTOTPStatus(ctx context.Context, userID int64) (bool, error) {
	if s.totpRepo == nil {
		return false, nil
	}
	secret, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	if secret == nil {
		return false, nil
	}
	return secret.Enabled, nil
}

func (s *AuthService) StartTOTPSetup(ctx context.Context, userID int64) (*TOTPSetupResult, error) {
	if s.totpRepo == nil {
		return nil, fmt.Errorf("totp repository not configured")
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	secret := generateTOTPSecret()
	enc := secret
	if s.encryptionKey != "" {
		if encrypted, err := config.Encrypt(s.encryptionKey, secret); err == nil {
			enc = encrypted
		}
	}
	if _, err := s.totpRepo.Upsert(ctx, userID, enc, false); err != nil {
		return nil, err
	}

	account := user.Username
	label := url.QueryEscape(fmt.Sprintf("%s:%s", totpIssuer, account))
	issuer := url.QueryEscape(totpIssuer)
	otpauth := fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s&digits=%d&period=%d", label, secret, issuer, totpDigits, totpPeriod)
	return &TOTPSetupResult{Secret: secret, OtpAuthURL: otpauth}, nil
}

// StartTOTPSetupWithChallenge allows setup using a short-lived setup challenge token.
func (s *AuthService) StartTOTPSetupWithChallenge(ctx context.Context, challenge string) (*TOTPSetupResult, int64, error) {
	userID, err := s.verifyTwoFAChallenge(challenge, "setup")
	if err != nil {
		return nil, 0, ErrTOTPChallengeInvalid
	}
	result, err := s.StartTOTPSetup(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return result, userID, nil
}

func (s *AuthService) EnableTOTP(ctx context.Context, userID int64, code string) ([]string, error) {
	if s.totpRepo == nil {
		return nil, fmt.Errorf("totp repository not configured")
	}
	secret, err := s.getDecryptedTOTPSecret(ctx, userID)
	if err != nil {
		return nil, err
	}
	if secret == "" {
		return nil, ErrTOTPNotSetup
	}
	if !verifyTOTP(secret, code, time.Now()) {
		return nil, ErrTOTPInvalid
	}
	if err := s.totpRepo.SetEnabled(ctx, userID, true); err != nil {
		return nil, err
	}
	_ = s.totpRepo.UpdateLastUsed(ctx, userID)
	codes, err := s.generateRecoveryCodes(ctx, userID)
	if err != nil {
		return nil, err
	}
	return codes, nil
}

// EnableTOTPWithChallenge enables 2FA using a short-lived setup challenge token.
func (s *AuthService) EnableTOTPWithChallenge(ctx context.Context, challenge, code string) ([]string, int64, error) {
	userID, err := s.verifyTwoFAChallenge(challenge, "setup")
	if err != nil {
		return nil, 0, ErrTOTPChallengeInvalid
	}
	codes, err := s.EnableTOTP(ctx, userID, code)
	if err != nil {
		return nil, 0, err
	}
	return codes, userID, nil
}

func (s *AuthService) DisableTOTP(ctx context.Context, userID int64, code string) error {
	if s.totpRepo == nil {
		return fmt.Errorf("totp repository not configured")
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if s.Is2FARequiredForRole(ctx, user.Role) {
		return ErrTOTPPolicyRequired
	}
	secret, err := s.getDecryptedTOTPSecret(ctx, userID)
	if err != nil {
		return err
	}
	if secret == "" {
		return ErrTOTPNotSetup
	}
	if !verifyTOTP(secret, code, time.Now()) {
		return ErrTOTPInvalid
	}
	return s.totpRepo.Delete(ctx, userID)
}

func (s *AuthService) ValidateTOTPForUser(ctx context.Context, userID int64, code string, recoveryCode string) error {
	if s.totpRepo == nil {
		return nil
	}
	secret, err := s.getDecryptedTOTPSecret(ctx, userID)
	if err != nil {
		return err
	}
	if secret == "" {
		return nil
	}
	enabled, _ := s.GetTOTPStatus(ctx, userID)
	if !enabled {
		return nil
	}
	if recoveryCode != "" {
		if ok, err := s.consumeRecoveryCode(ctx, userID, recoveryCode); err != nil {
			return err
		} else if ok {
			return nil
		}
	}
	if code == "" {
		return ErrTOTPRequired
	}
	if !verifyTOTP(secret, code, time.Now()) {
		return ErrTOTPInvalid
	}
	_ = s.totpRepo.UpdateLastUsed(ctx, userID)
	return nil
}

func (s *AuthService) RegenRecoveryCodes(ctx context.Context, userID int64, code string) ([]string, error) {
	if s.totpRepo == nil || s.recoveryRepo == nil {
		return nil, fmt.Errorf("recovery codes not configured")
	}
	secret, err := s.getDecryptedTOTPSecret(ctx, userID)
	if err != nil {
		return nil, err
	}
	if secret == "" {
		return nil, ErrTOTPNotSetup
	}
	if !verifyTOTP(secret, code, time.Now()) {
		return nil, ErrTOTPInvalid
	}
	return s.generateRecoveryCodes(ctx, userID)
}

func (s *AuthService) RecoveryCodesStatus(ctx context.Context, userID int64) (int, error) {
	if s.recoveryRepo == nil {
		return 0, nil
	}
	return s.recoveryRepo.CountAvailable(ctx, userID)
}

func (s *AuthService) generateRecoveryCodes(ctx context.Context, userID int64) ([]string, error) {
	if s.recoveryRepo == nil {
		return nil, fmt.Errorf("recovery repo not configured")
	}
	codes := make([]string, 0, 10)
	hashes := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		code := generateRecoveryCode()
		codes = append(codes, code)
		hashes = append(hashes, hashRecoveryCode(code))
	}
	if err := s.recoveryRepo.ReplaceCodes(ctx, userID, hashes); err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *AuthService) consumeRecoveryCode(ctx context.Context, userID int64, code string) (bool, error) {
	if s.recoveryRepo == nil {
		return false, nil
	}
	hash := hashRecoveryCode(code)
	return s.recoveryRepo.ConsumeCode(ctx, userID, hash)
}

func (s *AuthService) getDecryptedTOTPSecret(ctx context.Context, userID int64) (string, error) {
	secret, err := s.totpRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", nil
	}
	raw := secret.SecretEnc
	if s.encryptionKey != "" && raw != "" {
		if decrypted, err := config.Decrypt(s.encryptionKey, raw); err == nil {
			raw = decrypted
		}
	}
	return raw, nil
}

func generateTOTPSecret() string {
	b := make([]byte, 20)
	_, _ = rand.Read(b)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
}

func generateRecoveryCode() string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b)[:10])
}

func hashRecoveryCode(code string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(strings.ToUpper(code))))
	return hex.EncodeToString(h[:])
}

func verifyTOTP(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}
	for i := -1; i <= 1; i++ {
		t := now.Add(time.Duration(i*totpPeriod) * time.Second)
		if totpAt(secret, t) == code {
			return true
		}
	}
	return false
}

func totpAt(secret string, t time.Time) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}
	counter := uint64(t.Unix() / totpPeriod)
	var buf [8]byte
	for i := 7; i >= 0; i-- {
		buf[i] = byte(counter & 0xff)
		counter >>= 8
	}
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	bin := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)
	mod := 1
	for i := 0; i < totpDigits; i++ {
		mod *= 10
	}
	code := bin % mod
	return fmt.Sprintf("%0*d", totpDigits, code)
}
