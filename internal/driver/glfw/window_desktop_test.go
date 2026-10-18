//go:build !wasm && !test_web_driver

package glfw

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/stretchr/testify/assert"
)

var keyCodeMap = map[glfw.Key]granio.KeyName{
	// non-printable
	glfw.KeyEscape:    granio.KeyEscape,
	glfw.KeyEnter:     granio.KeyReturn,
	glfw.KeyTab:       granio.KeyTab,
	glfw.KeyBackspace: granio.KeyBackspace,
	glfw.KeyInsert:    granio.KeyInsert,
	glfw.KeyDelete:    granio.KeyDelete,
	glfw.KeyRight:     granio.KeyRight,
	glfw.KeyLeft:      granio.KeyLeft,
	glfw.KeyDown:      granio.KeyDown,
	glfw.KeyUp:        granio.KeyUp,
	glfw.KeyPageUp:    granio.KeyPageUp,
	glfw.KeyPageDown:  granio.KeyPageDown,
	glfw.KeyHome:      granio.KeyHome,
	glfw.KeyEnd:       granio.KeyEnd,

	glfw.KeySpace:   granio.KeySpace,
	glfw.KeyKPEnter: granio.KeyEnter,

	// functions
	glfw.KeyF1:  granio.KeyF1,
	glfw.KeyF2:  granio.KeyF2,
	glfw.KeyF3:  granio.KeyF3,
	glfw.KeyF4:  granio.KeyF4,
	glfw.KeyF5:  granio.KeyF5,
	glfw.KeyF6:  granio.KeyF6,
	glfw.KeyF7:  granio.KeyF7,
	glfw.KeyF8:  granio.KeyF8,
	glfw.KeyF9:  granio.KeyF9,
	glfw.KeyF10: granio.KeyF10,
	glfw.KeyF11: granio.KeyF11,
	glfw.KeyF12: granio.KeyF12,

	// numbers - lookup by code to avoid AZERTY using the symbol name instead of number
	glfw.Key0:   granio.Key0,
	glfw.KeyKP0: granio.Key0,
	glfw.Key1:   granio.Key1,
	glfw.KeyKP1: granio.Key1,
	glfw.Key2:   granio.Key2,
	glfw.KeyKP2: granio.Key2,
	glfw.Key3:   granio.Key3,
	glfw.KeyKP3: granio.Key3,
	glfw.Key4:   granio.Key4,
	glfw.KeyKP4: granio.Key4,
	glfw.Key5:   granio.Key5,
	glfw.KeyKP5: granio.Key5,
	glfw.Key6:   granio.Key6,
	glfw.KeyKP6: granio.Key6,
	glfw.Key7:   granio.Key7,
	glfw.KeyKP7: granio.Key7,
	glfw.Key8:   granio.Key8,
	glfw.KeyKP8: granio.Key8,
	glfw.Key9:   granio.Key9,
	glfw.KeyKP9: granio.Key9,

	// desktop
	glfw.KeyLeftShift:    desktop.KeyShiftLeft,
	glfw.KeyRightShift:   desktop.KeyShiftRight,
	glfw.KeyLeftControl:  desktop.KeyControlLeft,
	glfw.KeyRightControl: desktop.KeyControlRight,
	glfw.KeyLeftAlt:      desktop.KeyAltLeft,
	glfw.KeyRightAlt:     desktop.KeyAltRight,
	glfw.KeyLeftSuper:    desktop.KeySuperLeft,
	glfw.KeyRightSuper:   desktop.KeySuperRight,
	glfw.KeyMenu:         desktop.KeyMenu,
	glfw.KeyPrintScreen:  desktop.KeyPrintScreen,
	glfw.KeyCapsLock:     desktop.KeyCapsLock,
}

func TestGlfwKeyToKeyName(t *testing.T) {
	for key, value := range keyCodeMap {
		translated := glfwKeyToKeyName(key)
		assert.Equal(t, value, translated)
	}

	invalid := glfwKeyToKeyName(glfw.Key(-1))
	assert.Equal(t, granio.KeyUnknown, invalid)
}

func TestConvertASCII(t *testing.T) {
	for i := 0; i <= 'Z'-'A'; i++ {
		translated := convertASCII(glfw.KeyA + glfw.Key(i))
		expected := granio.KeyName(rune(granio.KeyA[0] + byte(i)))
		assert.Equal(t, expected, translated)
	}

	invalid := convertASCII(glfw.Key(-1))
	assert.Equal(t, granio.KeyUnknown, invalid)
}

var keyNameMapSpecialCharacters = map[string]granio.KeyName{
	"'": granio.KeyApostrophe,
	",": granio.KeyComma,
	"-": granio.KeyMinus,
	".": granio.KeyPeriod,
	"/": granio.KeySlash,
	"*": granio.KeyAsterisk,
	"`": granio.KeyBackTick,

	";": granio.KeySemicolon,
	"+": granio.KeyPlus,
	"=": granio.KeyEqual,

	"[":  granio.KeyLeftBracket,
	"\\": granio.KeyBackslash,
	"]":  granio.KeyRightBracket,
}

func TestKeyCodeToKeyName(t *testing.T) {
	for key, value := range keyNameMapSpecialCharacters {
		translated := keyCodeToKeyName(key)
		assert.Equal(t, value, translated)
	}

	for i := rune(0); i <= 'z'-'a'; i++ {
		translated := keyCodeToKeyName(string('a' + i))
		expected := granio.KeyName(rune(granio.KeyA[0]) + i)
		assert.Equal(t, expected, translated)
	}

	invalid := keyCodeToKeyName("@")
	assert.Equal(t, granio.KeyUnknown, invalid)

	invalid = keyCodeToKeyName("invalid")
	assert.Equal(t, granio.KeyUnknown, invalid)
}
