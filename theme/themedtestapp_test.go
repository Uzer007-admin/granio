package theme_test

import (
	"net/url"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

type themedApp struct {
	primaryColor string
	theme        granio.Theme
	variant      granio.ThemeVariant
}

func (*themedApp) CloudProvider() granio.CloudProvider {
	return nil
}

func (*themedApp) BuildType() granio.BuildType {
	return granio.BuildStandard
}

func (*themedApp) NewWindow(_ string) granio.Window {
	return nil
}

func (*themedApp) OpenURL(_ *url.URL) error {
	return nil
}

func (*themedApp) Icon() granio.Resource {
	return nil
}

func (*themedApp) SetIcon(granio.Resource) {
}

func (*themedApp) Run() {
}

func (*themedApp) Quit() {
}

func (*themedApp) Driver() granio.Driver {
	return nil
}

func (*themedApp) UniqueID() string {
	return ""
}

func (*themedApp) SendNotification(_ *granio.Notification) {
}

func (*themedApp) ScheduleNotification(_ *granio.Notification, _ time.Time) (*granio.ScheduledNotification, error) {
	return nil, nil
}

func (*themedApp) CancelScheduledNotification(_ string) error {
	return nil
}

func (a *themedApp) Settings() granio.Settings {
	return a
}

func (*themedApp) Storage() granio.Storage {
	return nil
}

func (*themedApp) Preferences() granio.Preferences {
	return nil
}

func (*themedApp) SecretPreferences([]byte) (granio.Preferences, error) {
	return nil, nil
}

func (*themedApp) Lifecycle() granio.Lifecycle {
	return nil
}

func (*themedApp) Metadata() granio.AppMetadata {
	return granio.AppMetadata{}
}

func (a *themedApp) PrimaryColor() string {
	if a.primaryColor != "" {
		return a.primaryColor
	}

	return theme.ColorBlue
}

func (a *themedApp) Theme() granio.Theme {
	return a.theme
}

func (a *themedApp) SetTheme(t granio.Theme) {
	a.theme = t
}

func (a *themedApp) ThemeVariant() granio.ThemeVariant {
	return a.variant // The null value is theme.VariantDark
}

func (*themedApp) SetCloudProvider(granio.CloudProvider) {
}

func (*themedApp) Scale() float32 {
	return 1.0
}

func (*themedApp) ShowAnimations() bool {
	return true
}

func (*themedApp) AddChangeListener(chan granio.Settings) {
}

func (*themedApp) AddListener(func(granio.Settings)) {
}

func (*themedApp) Cache() granio.Cache {
	return nil
}

func (*themedApp) Clipboard() granio.Clipboard {
	return nil
}
