// Command trashday opens the TrashDay window: a WebView2 page (package ui) that
// asks Go to measure or clean the targets of package cleaner.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	webview2 "github.com/jchv/go-webview2"

	"github.com/Sannho/trashday/internal/cleaner"
	"github.com/Sannho/trashday/ui"
)

type targetInfo struct {
	ID    string `json:"id"`
	Group string `json:"group"`
	Quick bool   `json:"quick"`
}

func main() {
	list := cleaner.Targets()

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "TrashDay",
			Width:  440,
			Height: 760,
			IconId: 2, // icon group resource ID assigned by rsrc (manifest is 1)
			Center: true,
		},
	})
	if w == nil {
		log.Fatal("WebView2 is not available: install the Microsoft Edge WebView2 runtime.")
	}
	defer w.Destroy()

	// call runs JS on the UI thread with JSON-encoded arguments.
	call := func(fn string, args ...any) {
		b, _ := json.Marshal(args)
		js := fmt.Sprintf("%s(...%s)", fn, b)
		w.Dispatch(func() { w.Eval(js) })
	}

	w.Bind("goI18n", ui.Texts)

	w.Bind("goTargets", func() []targetInfo {
		out := make([]targetInfo, len(list))
		for i, t := range list {
			out[i] = targetInfo{t.ID, t.Group, t.Quick}
		}
		return out
	})

	// Bindings run on the UI thread, so cleaning happens in a goroutine and
	// reports back through onProgress / onDone. With dry set, nothing is deleted:
	// Freed is then the size that would be freed.
	w.Bind("goClean", func(ids []int, dry bool) {
		go func() {
			var total cleaner.Result
			for _, id := range ids {
				if id < 0 || id >= len(list) {
					continue
				}
				call("onProgress", id, nil)
				r := cleaner.Clean(list[id], dry)
				total.Add(r)
				call("onProgress", id, r)
			}
			call("onDone", total)
		}()
	})

	w.SetHtml(ui.Page())
	w.Run()
}
