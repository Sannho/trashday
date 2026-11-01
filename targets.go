package main

import (
	"os"
	"path/filepath"
	"strings"
)

// tr is a UI text in French and English.
type tr struct{ FR, EN string }

// Target is one checkbox: a set of glob patterns whose matches get emptied (dirs) or deleted (files).
type Target struct {
	Group    tr
	Name     tr
	Desc     tr // optional one-liner under the name; keep it short (one line at 440 px)
	Patterns []string
	Quick    bool // part of "Nettoyage rapide"
}

var (
	gWindows  = tr{"Windows", "Windows"}
	gBrowsers = tr{"Navigateurs (cache uniquement)", "Browsers (cache only)"}
	gApps     = tr{"Applications", "Apps"}
	gDev      = tr{"Développement", "Development"}

	descPackages = tr{"Paquets re-téléchargés si besoin", "Packages re-downloaded when needed"}
)

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

func targets() []Target {
	return []Target{
		{Group: gWindows, Name: tr{"Temp Windows", "Windows temp"}, Quick: true,
			Desc: tr{"Fichiers temporaires du système", "System temporary files"},
			Patterns: []string{
				envPath("SystemRoot", "Temp"),
				envPath("SystemRoot", "SystemTemp"), // system-wide temp since Windows 11 24H2
			}},
		// Not %TEMP%: users and tools sometimes point it at a real folder (D:\, Downloads...).
		{Group: gWindows, Name: tr{"Temp utilisateur courant", "Current user temp"}, Quick: true,
			Desc:     tr{"Fichiers temporaires de votre session", "Your session's temporary files"},
			Patterns: []string{envPath("LOCALAPPDATA", "Temp")}},
		{Group: gWindows, Name: tr{"Temp tous les utilisateurs", "All users temp"}, Quick: true,
			Desc:     tr{"Temp de chaque compte de ce PC", "Temp of every account on this PC"},
			Patterns: []string{drivePath(`Users\*\AppData\Local\Temp`)}},
		{Group: gWindows, Name: tr{"Prefetch", "Prefetch"},
			Desc:     tr{"Recréé par Windows, démarrage plus lent une fois", "Rebuilt by Windows, one slower startup"},
			Patterns: []string{envPath("SystemRoot", "Prefetch")}},
		{Group: gWindows, Name: tr{"Rapports d'erreurs (WER)", "Error reports (WER)"}, Quick: true,
			Desc: tr{"Rapports de plantage envoyés ou en attente", "Crash reports, sent or queued"},
			Patterns: []string{
				envPath("ProgramData", `Microsoft\Windows\WER\ReportArchive`),
				envPath("ProgramData", `Microsoft\Windows\WER\ReportQueue`),
				envPath("LOCALAPPDATA", `Microsoft\Windows\WER\ReportArchive`),
				envPath("LOCALAPPDATA", `Microsoft\Windows\WER\ReportQueue`),
			}},
		{Group: gWindows, Name: tr{"Crash dumps", "Crash dumps"}, Quick: true,
			Desc: tr{"Copies mémoire laissées par les plantages", "Memory dumps left by crashes"},
			Patterns: []string{
				envPath("LOCALAPPDATA", "CrashDumps"),
				envPath("SystemRoot", "Minidump"),
				envPath("SystemRoot", "MEMORY.DMP"),
				envPath("SystemRoot", "LiveKernelReports"),
			}},
		{Group: gWindows, Name: tr{"Logs Windows (CBS / DISM / IIS)", "Windows logs (CBS / DISM / IIS)"},
			Desc: tr{"Journaux de maintenance de Windows", "Windows servicing logs"},
			Patterns: []string{
				envPath("SystemRoot", `Logs\CBS`),
				envPath("SystemRoot", `Logs\DISM`),
				drivePath(`inetpub\logs\LogFiles`),
			}},
		{Group: gWindows, Name: tr{"Cache des applications du Store", "Store apps cache"},
			Desc: tr{"Cache web et temp des applis du Store", "Web cache and temp of Store apps"},
			Patterns: []string{
				envPath("LOCALAPPDATA", `Packages\*\AC\INetCache`),
				envPath("LOCALAPPDATA", `Packages\*\AC\Temp`),
			}},
		{Group: gWindows, Name: tr{"Cache miniatures", "Thumbnail cache"},
			Desc: tr{"Aperçus d'images, recréés à l'affichage", "Image previews, rebuilt when shown"},
			Patterns: []string{
				envPath("LOCALAPPDATA", `Microsoft\Windows\Explorer\thumbcache_*.db`),
				envPath("LOCALAPPDATA", `Microsoft\Windows\Explorer\iconcache_*.db`),
			}},
		// Content.Outlook inside INetCache is never touched (see keepNames in clean.go).
		{Group: gWindows, Name: tr{"Fichiers Internet temporaires", "Temporary Internet files"}, Quick: true,
			Desc:     tr{"Cache WinINet, pièces jointes Outlook gardées", "WinINet cache, Outlook attachments kept"},
			Patterns: []string{envPath("LOCALAPPDATA", `Microsoft\Windows\INetCache`)}},

		{Group: gBrowsers, Name: tr{"Firefox", "Firefox"}, Patterns: []string{
			envPath("LOCALAPPDATA", `Mozilla\Firefox\Profiles\*\cache2`),
			envPath("LOCALAPPDATA", `Mozilla\Firefox\Profiles\*\startupCache`),
			envPath("LOCALAPPDATA", `Mozilla\Firefox\Profiles\*\thumbnails`),
		}},
		{Group: gBrowsers, Name: tr{"Opera / Opera GX", "Opera / Opera GX"}, Patterns: []string{
			envPath("LOCALAPPDATA", `Opera Software\*\Cache`),
			envPath("LOCALAPPDATA", `Opera Software\*\Default\Cache`),
		}},
		{Group: gBrowsers, Name: tr{"Chrome", "Chrome"}, Patterns: chromium("Google", "Chrome", "User Data")},
		{Group: gBrowsers, Name: tr{"Edge", "Edge"}, Patterns: chromium("Microsoft", "Edge", "User Data")},
		{Group: gBrowsers, Name: tr{"Brave", "Brave"}, Patterns: chromium("BraveSoftware", "Brave-Browser", "User Data")},

		{Group: gApps, Name: tr{"Cache Java", "Java cache"}, Patterns: []string{envPath("USERPROFILE", `AppData\LocalLow\Sun\Java\Deployment\cache`)}},
		{Group: gApps, Name: tr{"Cache shaders (DirectX / NVIDIA / AMD / Intel)", "Shader cache (DirectX / NVIDIA / AMD / Intel)"},
			Desc: tr{"Recréé au lancement des jeux", "Rebuilt when games start"},
			Patterns: []string{
				envPath("LOCALAPPDATA", "D3DSCache"),
				envPath("LOCALAPPDATA", `NVIDIA\DXCache`),
				envPath("LOCALAPPDATA", `NVIDIA\GLCache`),
				envPath("LOCALAPPDATA", `AMD\DxCache`),
				envPath("LOCALAPPDATA", `AMD\DxcCache`),
				envPath("LOCALAPPDATA", `Intel\ShaderCache`),
			}},
		{Group: gApps, Name: tr{"Steam (cache web + shaders)", "Steam (web cache + shaders)"}, Patterns: []string{
			envPath("LOCALAPPDATA", `Steam\htmlcache`),
			envPath("ProgramFiles(x86)", `Steam\steamapps\shadercache`),
		}},
		{Group: gApps, Name: tr{"Epic Games Launcher (cache web)", "Epic Games Launcher (web cache)"},
			Patterns: []string{envPath("LOCALAPPDATA", `EpicGamesLauncher\Saved\webcache*`)}},
		{Group: gApps, Name: tr{"Discord", "Discord"}, Patterns: []string{
			envPath("APPDATA", `discord\Cache`),
			envPath("APPDATA", `discord\Code Cache`),
			envPath("APPDATA", `discord\GPUCache`),
		}},
		{Group: gApps, Name: tr{"VS Code", "VS Code"}, Patterns: []string{
			envPath("APPDATA", `Code\Cache`),
			envPath("APPDATA", `Code\CachedData`),
			envPath("APPDATA", `Code\Code Cache`),
			envPath("APPDATA", `Code\GPUCache`),
		}},

		{Group: gDev, Name: tr{"Cache de build Go", "Go build cache"},
			Desc:     tr{"Prochaine compilation plus lente", "Next build will be slower"},
			Patterns: []string{envPath("LOCALAPPDATA", "go-build")}},
		{Group: gDev, Name: tr{"Cache npm", "npm cache"}, Desc: descPackages, Patterns: []string{envPath("LOCALAPPDATA", "npm-cache")}},
		{Group: gDev, Name: tr{"Cache Yarn", "Yarn cache"}, Desc: descPackages, Patterns: []string{envPath("LOCALAPPDATA", `Yarn\Cache`)}},
		{Group: gDev, Name: tr{"Cache pip", "pip cache"}, Desc: descPackages, Patterns: []string{envPath("LOCALAPPDATA", `pip\Cache`)}},
		{Group: gDev, Name: tr{"Cache HTTP NuGet", "NuGet HTTP cache"}, Desc: descPackages, Patterns: []string{envPath("LOCALAPPDATA", `NuGet\v3-cache`)}},
	}
}
