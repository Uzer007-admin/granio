package lang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Uzer007-admin/granio/v2"
)

func TestAddTranslations(t *testing.T) {
	err := AddTranslations(granio.NewStaticResource("en.json", []byte(`{
		"Test": "Match"
	}`)))
	if assert.NoError(t, err) {
		setupLang("en")
		assert.Equal(t, "Match", L("Test"))
	}

	err = AddTranslationsForLocale([]byte(`{
		"Test2": "Match2"
	}`), "fr")
	if assert.NoError(t, err) {
		setupLang("fr")
		assert.Equal(t, "Match2", L("Test2"))
	}
}

func TestLocalize_Default(t *testing.T) {
	fallback := closestSupportedLocale([]string{"xx"})
	assert.Equal(t, granio.Locale("en"), fallback[0:2])
}

func TestLocalize_Fallback(t *testing.T) {
	assert.Equal(t, "Missing", L("Missing"))
}

func TestLocalize_Struct(t *testing.T) {
	type data struct {
		Str string
	}

	assert.Equal(t, "Hello World!", L("Hello {{.Str}}!", data{Str: "World"}))
}

func TestLocalize_Map(t *testing.T) {
	assert.Equal(t, "Hello World!", L("Hello {{.Str}}!", map[string]string{"Str": "World"}))
}

func TestLocalizePlural_Fallback(t *testing.T) {
	require.NoError(t, AddTranslations(granio.NewStaticResource("en.json", []byte(`{
		"Apple": {
			"one": "Apple",
			"other": "Apples"
		}
	}`))))

	setupLang("en")
	assert.Equal(t, "Missing", N("Missing", 1))
	assert.Equal(t, "Apple", N("Apple", 1))
	assert.Equal(t, "Apples", N("Apple", 2))
}

func TestLocalizeKey_Fallback(t *testing.T) {
	require.NoError(t, AddTranslations(granio.NewStaticResource("en.json", []byte(`{
		"appleID": "Apple Matched"
	}`))))

	setupLang("en")
	assert.Equal(t, "Apple", X("appleIDMissing", "Apple"))
	assert.Equal(t, "Apple Matched", X("appleID", "Apple"))
}

func TestLocalizePluralKey_Fallback(t *testing.T) {
	require.NoError(t, AddTranslations(granio.NewStaticResource("en.json", []byte(`{
		"appleID": {
			"one": "Apple",
			"other": "Apples"
		}
	}`))))

	setupLang("en")
	assert.Equal(t, "Missing", XN("appleIDMissing", "Missing", 1))
	assert.Equal(t, "Apple", XN("appleID", "Apple", 1))
	assert.Equal(t, "Apples", XN("appleID", "Apple", 2))
}

func TestLoadBuiltInTranslations(t *testing.T) {
	assert.NoError(t, AddTranslationsFS(translations, "translations"))
	setupLang("en")
	assert.Equal(t, "OK", X("OK", "FAIL"))
}

func TestUpdateLocalizer_RetriesAfterRuntimeIsReady(t *testing.T) {
	origReady := runtimeReady
	defer func() {
		runtimeReady = origReady
		updateLocalizer()
	}()

	// Simulate Android package init, where the JVM context is not registered.
	runtimeReady = func() bool { return false }
	updateLocalizer()
	require.True(t, localePending.Load())
	require.NotNil(t, localizer.Load())

	assert.Equal(t, "Missing", L("Missing"))
	require.True(t, localePending.Load())

	runtimeReady = origReady
	assert.Equal(t, "Missing", L("Missing"))
	assert.False(t, localePending.Load())
}

func TestSetLocale(t *testing.T) {
	orig := localeOverride.Load()
	t.Cleanup(func() {
		localeOverride.Store(orig)
		updateLocalizer()
	})
	SetLocale("")
	systemTranslation := L("Cancel")

	SetLocale("en")
	assert.Equal(t, "Cancel", L("Cancel"))
	assert.Equal(t, "(Select one)", L("(Select one)"))
	SetLocale("ru")
	assert.Equal(t, "Отмена", L("Cancel"))
	assert.Equal(t, "(Выберите вариант)", L("(Select one)"))

	SetLocale("en")
	require.NoError(t, AddTranslationsForLocale([]byte(`{"locale_override_test": "Override retained"}`), "en"))
	assert.Equal(t, "Cancel", L("Cancel"))
	assert.Equal(t, "Override retained", L("locale_override_test"))

	SetLocale("und")
	assert.Equal(t, "Cancel", L("Cancel"))
	SetLocale("")
	assert.Nil(t, localeOverride.Load())
	assert.Equal(t, systemTranslation, L("Cancel"))
}
