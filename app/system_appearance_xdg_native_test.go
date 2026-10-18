//go:build system_appearance_integration && !ci && !wasm && !test_web_driver && !android && !ios && !mobile && linux && !tinygo && !noos && !tamago

package app

import (
	"os"
	"testing"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/dbus/v5/portal"
)

type nativeAppearanceService struct{}

func (nativeAppearanceService) ReadAll([]string) (map[string]map[string]dbus.Variant, *dbus.Error) {
	return map[string]map[string]dbus.Variant{portal.AppearanceNamespace: {"color-scheme": dbus.MakeVariant(uint32(1)), "contrast": dbus.MakeVariant(uint32(1)), "reduced-motion": dbus.MakeVariant(uint32(1))}}, nil
}

func TestNativeXDGAppearanceProviderLifecycle(t *testing.T) {
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		t.Skip("requires isolated dbus-run-session")
	}
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.Export(nativeAppearanceService{}, portal.SettingsPath, portal.SettingsInterface); err != nil {
		t.Fatal(err)
	}
	reply, err := conn.RequestName(portal.SettingsDestination, dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		t.Fatalf("isolated portal unavailable: %v %v", reply, err)
	}
	s := &settings{}
	s.startSystemAppearance()
	s.startSystemAppearance()
	defer s.stopSystemAppearance()
	// Startup returns only after the initial snapshot or an explicit fallback.
	// A later successful read must not hide a first-frame race in this test.
	if !s.SystemAppearance().Available {
		t.Fatal("initial portal snapshot missing before Show:", s.SystemAppearance())
	}
	snapshot := s.SystemAppearance()
	if !snapshot.ReducedMotion.Available || !snapshot.ReducedMotion.Value || !snapshot.HighContrast.Value || s.ShowAnimations() {
		t.Fatal(snapshot)
	}
	s.stopSystemAppearance()
	s.stopSystemAppearance()
	if s.SystemAppearance().Available {
		t.Fatal("closed provider still available")
	}
	if !conn.Connected() {
		t.Fatal("provider closed another connection")
	}
}
