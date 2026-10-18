"""Exercise production gamma-ramp bridges and WGL failure handling in C fixtures."""

import argparse
import os
from pathlib import Path
import re
import subprocess
import tempfile


ROOT = Path(__file__).resolve().parent.parent


def compile_and_run(directory, name, source, args, include=None, extra_source=None):
    file = directory / (name + ".c")
    executable = directory / (name + (".exe" if os.name == "nt" else ""))
    file.write_text(source, encoding="utf-8")
    command = [args.cc, "-std=c11", "-Wall", "-Wextra", "-Werror", str(file), "-o", str(executable)]
    if include:
        command.extend(["-I", str(include)])
    if extra_source:
        command.append(str(extra_source))
    if args.sanitize:
        command.extend(["-fsanitize=address,undefined", "-fno-omit-frame-pointer"])
    subprocess.run(command, check=True)
    subprocess.run([str(executable)], check=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cc", default="gcc")
    parser.add_argument("--sanitize", action="store_true")
    parser.add_argument("--temp-root", type=Path)
    args = parser.parse_args()
    temp_root = (args.temp_root or Path(tempfile.gettempdir())).resolve()
    with tempfile.TemporaryDirectory(prefix="native-bridges-", dir=temp_root) as temporary:
        directory = Path(temporary).resolve()
        if not directory.is_relative_to(temp_root):
            raise ValueError("Temporary directory escaped its root")
        for version in ("v3.3", "v3.4", "v3.5"):
            wrapper = ROOT / "deps/glfw" / version / "glfw"
            code = (wrapper / "monitor.c").read_text(encoding="utf-8")
            body = re.search(r"void glfwSetGammaRampBridge\([^}]+\}", code).group()
            source = """
#define GLFW_INCLUDE_NONE
#include <GLFW/glfw3.h>
#include <assert.h>
static unsigned short red[256], green[256], blue[256];
static int calls;
void glfwSetGammaRamp(GLFWmonitor *monitor, const GLFWgammaramp *ramp) {
  assert(monitor == NULL);
  assert(ramp->size == 256);
  assert(ramp->red == red && ramp->green == green && ramp->blue == blue);
  for (unsigned int i = 0; i < ramp->size; i++) {
    assert(ramp->red[i] == i && ramp->green[i] == i + 256 && ramp->blue[i] == 65535 - i);
  }
  calls++;
}
""" + body + """
int main(void) {
  for (unsigned int i = 0; i < 256; i++) {
    red[i] = i; green[i] = i + 256; blue[i] = 65535 - i;
  }
  glfwSetGammaRampBridge(NULL, red, green, blue, 256);
  assert(calls == 1);
  return 0;
}
"""
            compile_and_run(directory, version.replace(".", "-"), source, args, wrapper / "glfw/include")
            source = """
#define GLFW_INCLUDE_NONE
#include "native_array.h"
#include <assert.h>
static float axes[] = {-1, 0, 0.5f, 1};
static unsigned char buttons[] = {0, 1, 255}, hats[] = {0, 2};
static int monitor_token;
static GLFWmonitor *monitors[] = {(GLFWmonitor *) &monitor_token};
static GLFWvidmode modes[] = {{1920, 1080, 8, 8, 8, 60}};
const float *glfwGetJoystickAxes(int joystick, int *count) { assert(joystick == 7); *count = 4; return axes; }
const unsigned char *glfwGetJoystickButtons(int joystick, int *count) { assert(joystick == 7); *count = 3; return buttons; }
const unsigned char *glfwGetJoystickHats(int joystick, int *count) { assert(joystick == 7); *count = 2; return hats; }
int glfwGetGamepadState(int joystick, GLFWgamepadstate *state) {
  assert(joystick == 7);
  for (int i = 0; i < 15; i++) state->buttons[i] = i % 2;
  for (int i = 0; i < 6; i++) state->axes[i] = i * 0.125f;
  return GLFW_TRUE;
}
GLFWmonitor **glfwGetMonitors(int *count) { *count = 1; return monitors; }
const GLFWvidmode *glfwGetVideoModes(GLFWmonitor *monitor, int *count) { assert(monitor == monitors[0]); *count = 1; return modes; }
void glfwGetWindowSize(GLFWwindow *window, int *width, int *height) { assert(window == NULL); *width = 800; *height = 600; }
void glfwGetFramebufferSize(GLFWwindow *window, int *width, int *height) { assert(window == NULL); *width = 1200; *height = 900; }
void glfwGetWindowContentScale(GLFWwindow *window, float *x, float *y) { assert(window == NULL); *x = 1.25f; *y = 1.25f; }
int glfwGetWindowAttrib(GLFWwindow *window, int attribute) { assert(window == NULL); return attribute == GLFW_TRANSPARENT_FRAMEBUFFER || attribute == GLFW_VISIBLE; }
int main(void) {
  GLFWgoSurfaceGeometry geometry = glfwGetSurfaceGeometryBridge(NULL);
  assert(geometry.width == 800 && geometry.height == 600 && geometry.framebufferWidth == 1200 && geometry.framebufferHeight == 900);
  assert(geometry.scaleX == 1.25f && geometry.scaleY == 1.25f);
  GLFWgoSurfaceMetrics surface = glfwGetSurfaceMetricsBridge(NULL);
  assert(surface.width == 800 && surface.height == 600 && surface.framebufferWidth == 1200 && surface.framebufferHeight == 900);
  assert(surface.scaleX == 1.25f && surface.scaleY == 1.25f && surface.transparent && surface.visible && !surface.iconified);
  GLFWgoAxes a = glfwGetJoystickAxesBridge(7);
  GLFWgoBytes b = glfwGetJoystickButtonsBridge(7), h = glfwGetJoystickHatsBridge(7);
  GLFWgoGamepad g = glfwGetGamepadStateBridge(7);
  GLFWgoMonitors m = glfwGetMonitorsBridge();
  GLFWgoVideoModes v = glfwGetVideoModesBridge(monitors[0]);
  assert(a.data == axes && a.count == 4 && a.data[2] == 0.5f);
  assert(b.data == buttons && b.count == 3 && b.data[2] == 255);
  assert(h.data == hats && h.count == 2 && h.data[1] == 2);
  assert(g.present == GLFW_TRUE && g.state.buttons[13] == 1 && g.state.axes[5] == 0.625f);
  assert(m.count == 1 && m.data[0] == monitors[0]);
  assert(v.count == 1 && v.data[0].width == 1920 && v.data[0].refreshRate == 60);
  return 0;
}
"""
            compile_and_run(directory, version.replace(".", "-") + "-arrays", source, args, wrapper, wrapper / "native_array.c")
        loader_bodies = set()
        for path in (ROOT / "deps/gl").rglob("procaddr.go"):
            code = path.read_text(encoding="utf-8")
            body = re.search(r"static HMODULE ogl32dll = NULL;.*?(?=\s*#elif defined\(TAG_DARWIN\))", code, re.S)
            if body:
                loader_bodies.add(body.group().strip())
        if len(loader_bodies) != 1:
            raise ValueError("Generated WGL loaders differ or were not found")
        source = """
#include <assert.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>
typedef void *HMODULE;
typedef const char *LPCSTR;
static int token, fallback_token, module_token, fail_load, fallbacks;
static void *wgl_result;
static void *wglGetProcAddress(LPCSTR name) { assert(strcmp(name, "glExample") == 0); return wgl_result; }
static HMODULE LoadLibraryA(const char *name) { assert(strcmp(name, "opengl32.dll") == 0); return fail_load ? NULL : &module_token; }
static void *GetProcAddress(HMODULE module, LPCSTR name) {
  assert(module == &module_token && strcmp(name, "glExample") == 0);
  fallbacks++; return &fallback_token;
}
""" + loader_bodies.pop() + """
int main(void) {
  intptr_t invalid[] = {0, 1, 2, 3, -1};
  for (unsigned int i = 0; i < sizeof(invalid)/sizeof(invalid[0]); i++) {
    ogl32dll = NULL; fail_load = 0; fallbacks = 0; wgl_result = (void *) invalid[i];
    assert(GlowGetProcAddress("glExample") == &fallback_token && fallbacks == 1);
    ogl32dll = NULL; fail_load = 1; fallbacks = 0;
    assert(GlowGetProcAddress("glExample") == NULL && fallbacks == 0);
  }
  fail_load = 0; fallbacks = 0; wgl_result = &token;
  assert(GlowGetProcAddress("glExample") == &token && fallbacks == 0);
  return 0;
}
"""
        compile_and_run(directory, "wgl-loader", source, args)
    print("PASS: three gamma bridges; 18 array/state and three surface bridges; WGL loader: NULL, 1, 2, 3, -1, valid pointer and DLL failure")


if __name__ == "__main__":
    main()
