// Package dialogs shows the OS's own file dialogs and alerts, attached to the current window.
package dialogs

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/spartan-native/host-wails/internal/contract"
	"github.com/spartan-native/host-wails/internal/debug"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Short names for the contract's dialog types.
type (
	FileFilter        = contract.FileFilter
	OpenFileOptions   = contract.OpenFileOptions
	SaveFileOptions   = contract.SaveFileOptions
	PickFolderOptions = contract.PickFolderOptions
	PickedFile        = contract.PickedFile
	MessageButton     = contract.MessageButton
	MessageOptions    = contract.MessageOptions
)

// filterPattern turns ["png", ".jpg"] into Wails' "*.png;*.jpg". "*" anywhere means every file.
func filterPattern(extensions []string) string {
	parts := make([]string, 0, len(extensions))
	for _, e := range extensions {
		e = strings.TrimPrefix(strings.TrimSpace(e), ".")
		if e == "" {
			continue
		}
		if e == "*" {
			return "*"
		}
		parts = append(parts, "*."+e)
	}
	return strings.Join(parts, ";")
}

// pickedFiles turns paths into PickedFiles, skipping the empty path a cancelled dialog returns.
func pickedFiles(paths ...string) []PickedFile {
	out := make([]PickedFile, 0, len(paths))
	for _, p := range paths {
		if p != "" {
			out = append(out, PickedFile{Name: filepath.Base(p), Path: p})
		}
	}
	return out
}

func firstPicked(path string) *PickedFile {
	if files := pickedFiles(path); len(files) > 0 {
		return &files[0]
	}
	return nil
}

// messageButtons fills in the default: one OK button that is both default and cancel.
func messageButtons(o MessageOptions) MessageOptions {
	if len(o.Buttons) == 0 {
		o.Buttons = []MessageButton{{ID: "ok", Label: "OK"}}
		if o.DefaultID == "" {
			o.DefaultID = "ok"
		}
		if o.CancelID == "" {
			o.CancelID = "ok"
		}
	}
	return o
}

func currentWindow(app *application.App) application.Window {
	if w := app.Window.Current(); w != nil {
		return w
	}
	if all := app.Window.GetAll(); len(all) > 0 {
		return all[0]
	}
	return nil
}

func openDialog(app *application.App, title, button, directory string, filters []FileFilter) *application.OpenFileDialogStruct {
	d := app.Dialog.OpenFile().CanCreateDirectories(true)
	if title != "" {
		d.SetTitle(title)
	}
	if button != "" {
		d.SetButtonText(button)
	}
	if directory != "" {
		d.SetDirectory(directory)
	}
	for _, f := range filters {
		if p := filterPattern(f.Extensions); p != "" {
			d.AddFilter(f.Name, p)
		}
	}
	if w := currentWindow(app); w != nil {
		d.AttachToWindow(w)
	}
	return d
}

// OpenFile shows the OS's open dialog. A cancelled dialog returns no files.
func OpenFile(app *application.App, o OpenFileOptions) ([]PickedFile, error) {
	debug.Logf("OpenFile(%+v)", o)
	d := openDialog(app, o.Title, o.ButtonLabel, o.Directory, o.Filters).CanChooseFiles(true)
	if o.Multiple {
		paths, err := d.PromptForMultipleSelection()
		return pickedFiles(paths...), err
	}
	path, err := d.PromptForSingleSelection()
	return pickedFiles(path), err
}

// SaveFile shows the OS's save dialog. A cancelled dialog returns nil.
func SaveFile(app *application.App, o SaveFileOptions) (*PickedFile, error) {
	debug.Logf("SaveFile(%+v)", o)
	d := app.Dialog.SaveFile().CanCreateDirectories(true)
	if o.Title != "" {
		d.SetMessage(o.Title) // the save panel has no title; macOS shows the message instead
	}
	if o.ButtonLabel != "" {
		d.SetButtonText(o.ButtonLabel)
	}
	if o.Directory != "" {
		d.SetDirectory(o.Directory)
	}
	if o.FileName != "" {
		d.SetFilename(o.FileName)
	}
	for _, f := range o.Filters {
		if p := filterPattern(f.Extensions); p != "" {
			d.AddFilter(f.Name, p)
		}
	}
	if w := currentWindow(app); w != nil {
		d.AttachToWindow(w)
	}
	path, err := d.PromptForSingleSelection()
	return firstPicked(path), err
}

// PickFolder shows the OS's folder dialog. A cancelled dialog returns nil.
func PickFolder(app *application.App, o PickFolderOptions) (*PickedFile, error) {
	debug.Logf("PickFolder(%+v)", o)
	path, err := openDialog(app, o.Title, o.ButtonLabel, o.Directory, nil).
		CanChooseFiles(false).
		CanChooseDirectories(true).
		PromptForSingleSelection()
	return firstPicked(path), err
}

// Message shows the OS's alert and returns the id of the pressed button. If the alert is
// closed without one (GTK lets the user do that), the call ends when the frontend cancels it.
func Message(ctx context.Context, app *application.App, o MessageOptions) (*string, error) {
	o = messageButtons(o)
	debug.Logf("Message(%+v)", o)

	var d *application.MessageDialog
	switch o.Kind {
	case "question":
		d = app.Dialog.Question()
	case "warning":
		d = app.Dialog.Warning()
	case "error":
		d = app.Dialog.Error()
	default:
		d = app.Dialog.Info()
	}
	if o.Title != "" {
		d.SetTitle(o.Title)
	}
	d.SetMessage(o.Message)

	pressed := make(chan string, 1)
	for _, b := range o.Buttons {
		id := b.ID
		button := d.AddButton(b.Label).OnClick(func() {
			select {
			case pressed <- id:
			default:
			}
		})
		if id == o.DefaultID {
			button.SetAsDefault()
		}
		if id == o.CancelID {
			button.SetAsCancel()
		}
	}
	if w := currentWindow(app); w != nil {
		d.AttachToWindow(w)
	}
	d.Show()

	select {
	case id := <-pressed:
		return &id, nil
	case <-ctx.Done():
		if o.CancelID == "" {
			return nil, nil
		}
		return &o.CancelID, nil
	}
}
