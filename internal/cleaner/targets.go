package cleaner

import (
	"os"
	"path/filepath"
	"strings"
)

// Target is one checkbox: a set of glob patterns whose matches get emptied (dirs) or deleted (files).
// Its name and description live in ui/i18n/<lang>.json under "targets"."<ID>", its group under "groups".
type Target struct {
	ID       string // stable key: translations and the saved selection use it
	Group    string
	Patterns []string
	Quick    bool // part of "Quick clean"
}

// envPath joins an environment variable with sub-paths. It returns "" (target skipped)
// unless the variable is an absolute path that is not a drive root and has no glob
// characters, so a bad value can never widen what gets deleted.
func envPath(name string, parts ...string) string {
	base := filepath.Clean(os.Getenv(name))
	if !filepath.IsAbs(base) || filepath.Dir(base) == base || strings.ContainsAny(base, "*?[") {
		return ""
	}
	return filepath.Join(append([]string{base}, parts...)...)
}

// drivePath joins %SystemDrive% (a bare "C:", which envPath would reject) with sub-paths.
func drivePath(parts ...string) string {
	d := os.Getenv("SystemDrive")
	if len(d) != 2 || d[1] != ':' {
		return ""
	}
	return filepath.Join(append([]string{d + `\`}, parts...)...)
}

// chromium lists the cache folders of every profile of a Chromium-based browser,
// plus the shared shader caches and crash reports at the User Data level.
func chromium(userData ...string) []string {
	var out []string
	for _, cache := range []string{"Cache", "Code Cache", "GPUCache"} {
		out = append(out, envPath("LOCALAPPDATA", append(userData, "*", cache)...))
	}
	for _, cache := range []string{"GrShaderCache", "ShaderCache", `Crashpad\reports`} {
		out = append(out, envPath("LOCALAPPDATA", append(userData, cache)...))
	}
	return out
}

func Targets() []Target {
	return []Target{
		{ID: "windows-temp", Group: "windows", Quick: true,
			Patterns: []string{
				envPath("SystemRoot", "Temp"),
				envPath("SystemRoot", "SystemTemp"), // system-wide temp since Windows 11 24H2
			}},
		// Not %TEMP%: users and tools sometimes point it at a real folder (D:\, Downloads...).
		// The current user's Temp is listed too for profiles outside %SystemDrive%\Users;
		// Clean counts it once when both patterns match it.
		{ID: "users-temp", Group: "windows", Quick: true,
			Patterns: []string{
				envPath("LOCALAPPDATA", "Temp"),
				drivePath(`Users\*\AppData\Local\Temp`),
			}},
		{ID: "prefetch", Group: "windows",
			Patterns: []string{envPath("SystemRoot", "Prefetch")}},
		{ID: "error-reports", Group: "windows", Quick: true,
			Patterns: []string{
				envPath("ProgramData", `Microsoft\Windows\WER\ReportArchive`),
				envPath("ProgramData", `Microsoft\Windows\WER\ReportQueue`),
				envPath("LOCALAPPDATA", `Microsoft\Windows\WER\ReportArchive`),
				envPath("LOCALAPPDATA", `Microsoft\Windows\WER\ReportQueue`),
			}},
		{ID: "crash-dumps", Group: "windows", Quick: true,
			Patterns: []string{
				envPath("LOCALAPPDATA", "CrashDumps"),
				envPath("SystemRoot", "Minidump"),
				envPath("SystemRoot", "MEMORY.DMP"),
				envPath("SystemRoot", "LiveKernelReports"),
			}},
		{ID: "windows-logs", Group: "windows",
			Patterns: []string{
				envPath("SystemRoot", `Logs\CBS`),
				envPath("SystemRoot", `Logs\DISM`),
				drivePath(`inetpub\logs\LogFiles`),
			}},
		{ID: "store-apps", Group: "windows",
			Patterns: []string{
				envPath("LOCALAPPDATA", `Packages\*\AC\INetCache`),
				envPath("LOCALAPPDATA", `Packages\*\AC\Temp`),
			}},
		{ID: "thumbnails", Group: "windows",
			Patterns: []string{
				envPath("LOCALAPPDATA", `Microsoft\Windows\Explorer\thumbcache_*.db`),
				envPath("LOCALAPPDATA", `Microsoft\Windows\Explorer\iconcache_*.db`),
			}},
		// Content.Outlook inside INetCache is never touched (see keepNames in clean.go).
		{ID: "internet-files", Group: "windows", Quick: true,
			Patterns: []string{envPath("LOCALAPPDATA", `Microsoft\Windows\INetCache`)}},

		{ID: "firefox", Group: "browsers", Patterns: []string{
			envPath("LOCALAPPDATA", `Mozilla\Firefox\Profiles\*\cache2`),
			envPath("LOCALAPPDATA", `Mozilla\Firefox\Profiles\*\startupCache`),
			envPath("LOCALAPPDATA", `Mozilla\Firefox\Profiles\*\thumbnails`),
		}},
		{ID: "opera", Group: "browsers", Patterns: []string{
			envPath("LOCALAPPDATA", `Opera Software\*\Cache`),
			envPath("LOCALAPPDATA", `Opera Software\*\Default\Cache`),
		}},
		{ID: "chrome", Group: "browsers", Patterns: chromium("Google", "Chrome", "User Data")},
		{ID: "edge", Group: "browsers", Patterns: chromium("Microsoft", "Edge", "User Data")},
		{ID: "brave", Group: "browsers", Patterns: chromium("BraveSoftware", "Brave-Browser", "User Data")},

		{ID: "java", Group: "apps", Patterns: []string{envPath("USERPROFILE", `AppData\LocalLow\Sun\Java\Deployment\cache`)}},
		{ID: "shaders", Group: "apps",
			Patterns: []string{
				envPath("LOCALAPPDATA", "D3DSCache"),
				envPath("LOCALAPPDATA", `NVIDIA\DXCache`),
				envPath("LOCALAPPDATA", `NVIDIA\GLCache`),
				envPath("LOCALAPPDATA", `AMD\DxCache`),
				envPath("LOCALAPPDATA", `AMD\DxcCache`),
				envPath("LOCALAPPDATA", `Intel\ShaderCache`),
			}},
		{ID: "steam", Group: "apps", Patterns: []string{
			envPath("LOCALAPPDATA", `Steam\htmlcache`),
			envPath("ProgramFiles(x86)", `Steam\steamapps\shadercache`),
		}},
		{ID: "epic", Group: "apps",
			Patterns: []string{envPath("LOCALAPPDATA", `EpicGamesLauncher\Saved\webcache*`)}},
		{ID: "discord", Group: "apps", Patterns: []string{
			envPath("APPDATA", `discord\Cache`),
			envPath("APPDATA", `discord\Code Cache`),
			envPath("APPDATA", `discord\GPUCache`),
		}},
		{ID: "vscode", Group: "apps", Patterns: []string{
			envPath("APPDATA", `Code\Cache`),
			envPath("APPDATA", `Code\CachedData`),
			envPath("APPDATA", `Code\Code Cache`),
			envPath("APPDATA", `Code\GPUCache`),
		}},

		{ID: "go-build", Group: "dev",
			Patterns: []string{envPath("LOCALAPPDATA", "go-build")}},
		{ID: "npm", Group: "dev", Patterns: []string{envPath("LOCALAPPDATA", "npm-cache")}},
		{ID: "yarn", Group: "dev", Patterns: []string{envPath("LOCALAPPDATA", `Yarn\Cache`)}},
		{ID: "pip", Group: "dev", Patterns: []string{envPath("LOCALAPPDATA", `pip\Cache`)}},
		{ID: "nuget", Group: "dev", Patterns: []string{envPath("LOCALAPPDATA", `NuGet\v3-cache`)}},
	}
}
