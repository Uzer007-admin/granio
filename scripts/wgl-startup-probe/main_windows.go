//go:build windows && cgo

package main

/*
#include <windows.h>
#include <versionhelpers.h>
#include <stdint.h>
typedef int (WINAPI *GetSwapInterval)(void);
static int querySwapInterval(void* proc) {
 uintptr_t address=(uintptr_t)proc;
 if(address<=3 || address==UINTPTR_MAX)return -1;
 return ((GetSwapInterval)proc)();
}
static double counterMilliseconds(void) {
 LARGE_INTEGER counter,frequency;
 QueryPerformanceCounter(&counter);
 QueryPerformanceFrequency(&frequency);
 return (double)counter.QuadPart*1000.0/frequency.QuadPart;
}
static int isWin8(void) { return IsWindows8OrGreater(); }
static int isWin81(void) { return IsWindows8Point1OrGreater(); }
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	gl "github.com/Uzer007-admin/gl/v2.1/material"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

func main() {
	runtime.LockOSThread()
	scenario := "noapi"
	if len(os.Args) > 1 {
		scenario = os.Args[1]
	}
	if scenario != "noapi" && scenario != "wgl" {
		fmt.Fprintln(os.Stderr, "Используйте noapi или wgl")
		os.Exit(2)
	}
	started := time.Now()
	stage := func(label string, begin time.Time) {
		json.NewEncoder(os.Stdout).Encode(map[string]any{"stage": label, "milliseconds": float64(time.Since(begin)) / float64(time.Millisecond)})
	}
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	stage("glfw_init", started)
	defer glfw.Terminate()
	glfw.WindowHint(glfw.Visible, glfw.False)
	if scenario == "noapi" {
		glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	} else {
		glfw.WindowHint(glfw.ContextVersionMajor, 2)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
	}
	before := time.Now()
	window, err := glfw.CreateWindow(320, 240, "Проверка WGL Granio", nil, nil)
	if err != nil {
		panic(err)
	}
	stage("create_window", before)
	defer window.Destroy()
	if scenario == "noapi" {
		return
	}
	before = time.Now()
	window.MakeContextCurrent()
	stage("make_current", before)
	api, err := gl.LoadWithProcAddrFunc(glfw.GetProcAddress)
	if err != nil {
		panic(err)
	}
	proc := glfw.GetProcAddress("wglGetSwapIntervalEXT")
	initial := int(C.querySwapInterval(proc))
	json.NewEncoder(os.Stdout).Encode(map[string]any{"initial_swap_interval": initial, "is_win8": C.isWin8() != 0, "is_win81": C.isWin81() != 0})
	before = time.Now()
	window.Show()
	stage("show", before)
	for _, interval := range []int{-1, 0, 1} {
		if interval >= 0 {
			glfw.SwapInterval(interval)
		}
		actual := int(C.querySwapInterval(proc))
		swaps := make([]float64, 0, 12)
		for range 12 {
			glfw.PollEvents()
			api.ClearColor(0, 0.4, 0, 1)
			api.Clear(0x4000)
			beforeCounter := float64(C.counterMilliseconds())
			window.SwapBuffers()
			swaps = append(swaps, float64(C.counterMilliseconds())-beforeCounter)
			time.Sleep(16 * time.Millisecond)
		}
		json.NewEncoder(os.Stdout).Encode(map[string]any{"requested_swap_interval": interval, "actual_swap_interval": actual, "swap_ms": swaps})
	}
	glfw.DetachCurrentContext()
	fmt.Fprintln(os.Stderr, "completed", time.Since(started))
}
