//go:build debug

package routines

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

var (
	lib           syscall.Handle
	renderProcPtr uintptr
)

func libExt() string {
	switch runtime.GOOS {
	case "windows":
		return ".dll"
	case "linux", "android":
		return ".so"
	case "darwin":
		return ".dylib"
	default:
		return ""
	}
}

func initGui() (err error) {
	path, ok := os.LookupEnv("FREDBOARD_LIBGUI")
	if !ok {
		path, err = filepath.Abs(fmt.Sprintf("./bin/libgui%s", libExt()))
		if err != nil {
			return fmt.Errorf("failed to resolve path to library: %w", err)
		}
	}

	lib, err = syscall.LoadLibrary(path)
	if err != nil {
		return fmt.Errorf("failed to load library: %w", err)
	}

	renderProcPtr, err = syscall.GetProcAddress(lib, "Render")
	if err != nil {
		return fmt.Errorf("failed to resolve address for process Render: %w", err)
	}

	return nil
}

func renderGui() error {
	_, _, _ = syscall.SyscallN(renderProcPtr)

	return nil
}

func deinitGui() (err error) {
	err = syscall.FreeLibrary(lib)
	if err != nil {
		return fmt.Errorf("failed to free library: %w", err)
	}

	return nil
}
