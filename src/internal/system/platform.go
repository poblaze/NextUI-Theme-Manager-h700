// src/internal/system/platform.go
// Platform detection and per-platform file locations (tg5040 / h700)

package system

import (
	"os"
	"path/filepath"
	"sync"
)

// Supported NextUI platform tags
const (
	PlatformTG5040 = "tg5040" // TrimUI Brick / Smart Pro
	PlatformH700   = "h700"   // Anbernic RG XX family (Allwinner H700)

	defaultSDCardPath = "/mnt/SDCARD"
)

// Platform describes the device the theme manager is running on
type Platform struct {
	Name       string // NextUI platform tag, e.g. "tg5040" or "h700"
	Device     string // NextUI device name, e.g. "brick" or "rg40xxh"
	SDCardPath string // SD card root, normally /mnt/SDCARD
}

var (
	current     *Platform
	currentOnce sync.Once
)

// Current returns the detected platform. Detection runs once per process.
func Current() *Platform {
	currentOnce.Do(func() {
		current = detectPlatform()
	})
	return current
}

// detectPlatform reads the values the NextUI launcher exports to paks
// (PLATFORM, DEVICE, SDCARD_PATH) and falls back to sensible defaults.
func detectPlatform() *Platform {
	sdcard := os.Getenv("SDCARD_PATH")
	if sdcard == "" {
		sdcard = defaultSDCardPath
	}

	name := os.Getenv("PLATFORM")
	if !isKnownPlatform(name) {
		name = platformFromPakDir()
	}

	device := os.Getenv("DEVICE")
	if device == "" && name == PlatformTG5040 {
		device = "brick" // behaviour before multi-platform support
	}

	return &Platform{Name: name, Device: device, SDCardPath: sdcard}
}

// platformFromPakDir guesses the platform from where the pak is installed:
// Tools/<platform>/Theme Manager.pak. An SD card can hold .system folders for
// several platforms, so the install folder is a more reliable signal.
func platformFromPakDir() string {
	if cwd, err := os.Getwd(); err == nil {
		if tag := filepath.Base(filepath.Dir(cwd)); isKnownPlatform(tag) {
			return tag
		}
	}
	return PlatformTG5040
}

func isKnownPlatform(name string) bool {
	return name == PlatformTG5040 || name == PlatformH700
}

// IsH700 reports whether this is an Allwinner H700 device
func (p *Platform) IsH700() bool { return p.Name == PlatformH700 }

// SystemPath is the platform's .system folder, e.g. /mnt/SDCARD/.system/h700
func (p *Platform) SystemPath() string {
	if v := os.Getenv("SYSTEM_PATH"); v != "" {
		return v
	}
	return filepath.Join(p.SDCardPath, ".system", p.Name)
}

// ToolsPath is the platform's Tools folder, e.g. /mnt/SDCARD/Tools/h700
func (p *Platform) ToolsPath() string {
	return filepath.Join(p.SDCardPath, "Tools", p.Name)
}

// ToolsIconName is the file name of the Tools folder icon in Tools/.media
func (p *Platform) ToolsIconName() string {
	return p.Name + ".png"
}

// ToolsIconPath is the full path of the Tools folder icon
func (p *Platform) ToolsIconPath() string {
	return filepath.Join(p.SDCardPath, "Tools", ".media", p.ToolsIconName())
}

// ResPath is the shared resources folder that holds the UI fonts
func (p *Platform) ResPath() string {
	return filepath.Join(p.SDCardPath, ".system", "res")
}

// FontPath returns the full path of a font file in the resources folder.
// font1.ttf is NextUI's "Next" font and font2.ttf the "OG" font on every platform.
func (p *Platform) FontPath(file string) string {
	return filepath.Join(p.ResPath(), file)
}

// SharedUserdataPath is the settings folder shared by all platforms
func (p *Platform) SharedUserdataPath() string {
	if v := os.Getenv("SHARED_USERDATA_PATH"); v != "" {
		return v
	}
	return filepath.Join(p.SDCardPath, ".userdata", "shared")
}

// MinuiSettingsPath is NextUI's settings file, which holds the accent colours
func (p *Platform) MinuiSettingsPath() string {
	return filepath.Join(p.SharedUserdataPath(), "minuisettings.txt")
}

// SupportsLEDs reports whether the theme manager's LED features apply here.
// They read and write the TrimUI Brick LED layout (F1/F2 keys, top bar,
// L/R triggers). H700 LED files use a different layout ([Left]/[Right]), so
// LED features are turned off there until that layout is supported.
func (p *Platform) SupportsLEDs() bool {
	return p.Name == PlatformTG5040
}

// LEDSettingsPath is the LED settings file the theme manager reads and writes.
// Only meaningful when SupportsLEDs is true.
func (p *Platform) LEDSettingsPath() string {
	return filepath.Join(p.SharedUserdataPath(), "ledsettings_brick.txt")
}

// LibraryPath is the LD_LIBRARY_PATH the NextUI launcher uses on this platform
func (p *Platform) LibraryPath() string {
	lib := filepath.Join(p.SystemPath(), "lib")
	if p.IsH700() {
		return lib + ":/usr/lib:/usr/lib/aarch64-linux-gnu:/lib/aarch64-linux-gnu"
	}
	return lib + ":/usr/trimui/lib"
}

// HelperPath returns the path of a bundled UI helper (minui-list or
// minui-presenter) for this platform. Release paks keep one set per platform in
// bin/<platform>/; a helper next to the binary is used as a fallback.
func (p *Platform) HelperPath(pakDir, name string) string {
	perPlatform := filepath.Join(pakDir, "bin", p.Name, name)
	if _, err := os.Stat(perPlatform); err == nil {
		return perPlatform
	}
	return filepath.Join(pakDir, name)
}
