package security

import (
	"path/filepath"
	"testing"

	"cyberstrike-ai/internal/database"

	"go.uber.org/zap"
)

func TestAttachRBACStoreBootstrapsAdminPassword(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "auth-bootstrap.db"), zap.NewNop())
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	manager := NewAuthManager(12)
	setupCode, err := manager.AttachRBACStore(db)
	if err != nil {
		t.Fatalf("AttachRBACStore: %v", err)
	}
	if len(setupCode) != 8 {
		t.Fatalf("expected 8-char setup code on first bootstrap, got %q", setupCode)
	}
	if !manager.NeedsSetup() {
		t.Fatal("expected needsSetup=true after first bootstrap")
	}

	second, err := manager.AttachRBACStore(db)
	if err != nil {
		t.Fatalf("AttachRBACStore second call: %v", err)
	}
	if second != "" {
		t.Fatalf("expected no setup code on second bootstrap, got %q", second)
	}
}

func TestCompleteSetupLifecycle(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "auth-setup.db"), zap.NewNop())
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	manager := NewAuthManager(12)
	setupCode, err := manager.AttachRBACStore(db)
	if err != nil {
		t.Fatalf("AttachRBACStore: %v", err)
	}

	// 错误的设置码必须被拒绝
	if err := manager.CompleteSetup("XXXXXXXX", "goodpassword1"); err == nil {
		t.Fatal("expected wrong setup code to be rejected")
	}
	// 短密码必须被拒绝
	if err := manager.CompleteSetup(setupCode, "short"); err == nil {
		t.Fatal("expected short password to be rejected")
	}

	if err := manager.CompleteSetup(setupCode, "goodpassword1"); err != nil {
		t.Fatalf("CompleteSetup: %v", err)
	}
	if manager.NeedsSetup() {
		t.Fatal("needsSetup should be false after CompleteSetup")
	}
	// 初始化后新密码可登录
	if !manager.CheckUserPassword("admin", "goodpassword1") {
		t.Fatal("new password should authenticate admin after setup")
	}
	// 设置码已作废
	if err := manager.CompleteSetup(setupCode, "anotherpassword1"); err == nil {
		t.Fatal("setup code must be invalid after completed setup")
	}
}
