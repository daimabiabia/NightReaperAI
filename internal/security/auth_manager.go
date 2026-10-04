package security

import (
	"database/sql"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"cyberstrike-ai/internal/database"

	"github.com/google/uuid"
)

// Predefined errors for authentication operations.
var (
	ErrInvalidPassword = errors.New("invalid password")
)

// Session represents an authenticated user session.
type Session struct {
	Token            string
	ExpiresAt        time.Time
	UserID           string
	Username         string
	DisplayName      string
	Roles            []string
	Permissions      map[string]bool
	PermissionScopes map[string]string
	Scope            string
}

// AuthManager manages password-based authentication and session lifecycle.
type AuthManager struct {
	sessionDuration time.Duration
	db              *database.DB

	mu       sync.RWMutex
	sessions map[string]Session

	// First-run setup state. setupCode lives only in memory for the current
	// process run and is shown once in the console banner; it gates admin
	// initialization from the web UI while no admin password exists yet.
	needsSetup bool
	setupCode  string
}

// NewAuthManager creates a new AuthManager instance.
func NewAuthManager(sessionDurationHours int) *AuthManager {
	if sessionDurationHours <= 0 {
		sessionDurationHours = 12
	}

	return &AuthManager{
		sessionDuration: time.Duration(sessionDurationHours) * time.Hour,
		sessions:        make(map[string]Session),
	}
}

// AttachRBACStore enables multi-user RBAC authentication. When the built-in
// admin has no usable password yet (fresh DB or empty hash), it bootstraps
// with an empty hash and returns a short setup code for the web-based
// first-run initialization wizard. The pending-setup state persists across
// restarts until CompleteSetup sets a real password.
func (a *AuthManager) AttachRBACStore(db *database.DB) (setupCode string, err error) {
	if db == nil {
		return "", errors.New("database is required for authentication")
	}

	needsAdminPassword, err := db.RBACNeedsAdminPassword()
	if err != nil {
		return "", err
	}

	if err := db.BootstrapRBAC("", PermissionCatalog); err != nil {
		return "", err
	}

	a.mu.Lock()
	a.db = db
	if needsAdminPassword {
		a.needsSetup = true
		code, codeErr := GenerateSetupCode(8)
		if codeErr != nil {
			a.mu.Unlock()
			return "", codeErr
		}
		a.setupCode = code
		setupCode = code
	}
	a.mu.Unlock()
	return setupCode, nil
}

// NeedsSetup reports whether the built-in admin account is awaiting first-run
// initialization (no usable password has been configured yet).
func (a *AuthManager) NeedsSetup() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.needsSetup
}

// CompleteSetup validates the one-time setup code and sets the initial admin
// password from the web UI. After success the setup state is cleared and the
// code becomes invalid.
func (a *AuthManager) CompleteSetup(code, password string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	password = strings.TrimSpace(password)
	if password == "" {
		return errors.New("密码不能为空")
	}
	if len(password) < 8 {
		return fmt.Errorf("密码长度至少需要 8 位")
	}

	a.mu.RLock()
	needsSetup, setupCode, db := a.needsSetup, a.setupCode, a.db
	a.mu.RUnlock()

	if !needsSetup || db == nil {
		return errors.New("管理员账号已完成初始化")
	}
	if code == "" || subtle.ConstantTimeCompare([]byte(code), []byte(setupCode)) != 1 {
		return errors.New("初始化码不正确")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	if err := db.UpdateRBACAdminPassword(hash); err != nil {
		return err
	}

	a.mu.Lock()
	a.needsSetup = false
	a.setupCode = ""
	a.mu.Unlock()
	return nil
}

// Authenticate validates the password and creates a new session.
func (a *AuthManager) Authenticate(username, password string) (string, time.Time, error) {
	session, err := a.authenticateSession(username, password)
	if err != nil {
		return "", time.Time{}, err
	}
	a.mu.Lock()
	a.sessions[session.Token] = session
	a.mu.Unlock()
	return session.Token, session.ExpiresAt, nil
}

func (a *AuthManager) authenticateSession(username, password string) (Session, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(a.sessionDuration)

	a.mu.RLock()
	db := a.db
	a.mu.RUnlock()
	if db == nil {
		return Session{}, errors.New("authentication store is not configured")
	}

	username = strings.TrimSpace(strings.ToLower(username))
	if username == "" {
		username = "admin"
	}
	user, err := db.GetRBACUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			return Session{}, ErrInvalidPassword
		}
		return Session{}, err
	}
	if !user.Enabled || !VerifyPasswordHash(password, user.PasswordHash) {
		return Session{}, ErrInvalidPassword
	}
	access, err := db.ResolveRBACAccess(user.ID)
	if err != nil {
		return Session{}, err
	}
	roleIDs := make([]string, 0, len(access.Roles))
	for _, role := range access.Roles {
		roleIDs = append(roleIDs, role.ID)
	}
	return Session{
		Token:            token,
		ExpiresAt:        expiresAt,
		UserID:           user.ID,
		Username:         user.Username,
		DisplayName:      user.DisplayName,
		Roles:            roleIDs,
		Permissions:      access.Permissions,
		PermissionScopes: access.PermissionScopes,
		Scope:            access.Scope,
	}, nil
}

func (s Session) ScopeFor(permission string) string {
	if scope := strings.TrimSpace(s.PermissionScopes[strings.TrimSpace(permission)]); scope != "" {
		return scope
	}
	return strings.TrimSpace(s.Scope)
}

// ValidateToken checks whether the provided token is still valid.
func (a *AuthManager) ValidateToken(token string) (Session, bool) {
	if strings.TrimSpace(token) == "" {
		return Session{}, false
	}

	a.mu.RLock()
	session, ok := a.sessions[token]
	a.mu.RUnlock()
	if !ok {
		return Session{}, false
	}

	if time.Now().After(session.ExpiresAt) {
		a.mu.Lock()
		delete(a.sessions, token)
		a.mu.Unlock()
		return Session{}, false
	}

	return session, true
}

// CheckPassword verifies whether the provided password matches the current password.
func (a *AuthManager) CheckPassword(password string) bool {
	return a.CheckUserPassword("admin", password)
}

// CheckUserPassword verifies whether the provided password matches a user.
func (a *AuthManager) CheckUserPassword(username, password string) bool {
	a.mu.RLock()
	db := a.db
	a.mu.RUnlock()
	if db == nil {
		return false
	}
	user, err := db.GetRBACUserByUsername(username)
	if err != nil {
		return false
	}
	return VerifyPasswordHash(password, user.PasswordHash)
}

func (a *AuthManager) UpdateUserPassword(userID, password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return errors.New("auth password must be configured")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	a.mu.RLock()
	db := a.db
	a.mu.RUnlock()
	if db == nil {
		return errors.New("authentication store is not configured")
	}
	if err := db.UpdateRBACUserPassword(userID, hash); err != nil {
		return err
	}
	a.mu.Lock()
	for token, session := range a.sessions {
		if session.UserID == userID {
			delete(a.sessions, token)
		}
	}
	a.mu.Unlock()
	return nil
}

// RevokeToken invalidates the specified token.
func (a *AuthManager) RevokeToken(token string) {
	if strings.TrimSpace(token) == "" {
		return
	}

	a.mu.Lock()
	delete(a.sessions, token)
	a.mu.Unlock()
}

func (a *AuthManager) RevokeUserSessions(userID string) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return
	}
	a.mu.Lock()
	for token, session := range a.sessions {
		if session.UserID == userID {
			delete(a.sessions, token)
		}
	}
	a.mu.Unlock()
}

func (a *AuthManager) RevokeAllSessions() {
	a.mu.Lock()
	a.sessions = make(map[string]Session)
	a.mu.Unlock()
}

// SessionDurationHours returns the configured session duration in hours.
func (a *AuthManager) SessionDurationHours() int {
	return int(a.sessionDuration / time.Hour)
}

func allPermissions() map[string]bool {
	out := make(map[string]bool, len(PermissionCatalog))
	for key := range PermissionCatalog {
		out[key] = true
	}
	return out
}
