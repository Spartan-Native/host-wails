package spartannative

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFilterPattern(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"png", ".jpg"}, "*.png;*.jpg"},
		{[]string{" pdf "}, "*.pdf"},
		{[]string{"png", "*"}, "*"},
		{[]string{"", "."}, ""},
		{nil, ""},
	}
	for _, c := range cases {
		if got := filterPattern(c.in); got != c.want {
			t.Errorf("filterPattern(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPickedFilesSkipsCancelledPaths(t *testing.T) {
	got := pickedFiles("", "/home/me/Report.pdf")
	want := []PickedFile{{Name: "Report.pdf", Path: "/home/me/Report.pdf"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pickedFiles = %+v, want %+v", got, want)
	}
	if firstPicked("") != nil {
		t.Error("firstPicked(\"\") must be nil (a cancelled dialog)")
	}
	// A cancelled open dialog reaches the frontend as [], not null.
	if b, _ := json.Marshal(pickedFiles()); string(b) != "[]" {
		t.Errorf("no files marshals to %s, want []", b)
	}
}

func TestMessageButtonsDefaultsToOK(t *testing.T) {
	got := messageButtons(MessageOptions{Message: "Done"})
	if len(got.Buttons) != 1 || got.Buttons[0].ID != "ok" || got.DefaultID != "ok" || got.CancelID != "ok" {
		t.Errorf("messageButtons = %+v, want one OK button that is default and cancel", got)
	}
	custom := MessageOptions{Message: "Save?", Buttons: []MessageButton{{ID: "save", Label: "Save"}}, DefaultID: "save"}
	if got := messageButtons(custom); !reflect.DeepEqual(got, custom) {
		t.Errorf("messageButtons changed explicit buttons: %+v", got)
	}
}

func TestDialogOptionsMatchTheTSContract(t *testing.T) {
	var o OpenFileOptions
	in := `{"title":"Open","buttonLabel":"Import","directory":"/tmp","multiple":true,"filters":[{"name":"Images","extensions":["png"]}]}`
	if err := json.Unmarshal([]byte(in), &o); err != nil {
		t.Fatal(err)
	}
	want := OpenFileOptions{Title: "Open", ButtonLabel: "Import", Directory: "/tmp", Multiple: true, Filters: []FileFilter{{Name: "Images", Extensions: []string{"png"}}}}
	if !reflect.DeepEqual(o, want) {
		t.Errorf("OpenFileOptions = %+v, want %+v", o, want)
	}
	var m MessageOptions
	if err := json.Unmarshal([]byte(`{"kind":"warning","message":"x","defaultId":"a","cancelId":"b"}`), &m); err != nil {
		t.Fatal(err)
	}
	if m.Kind != "warning" || m.DefaultID != "a" || m.CancelID != "b" {
		t.Errorf("MessageOptions = %+v", m)
	}
}
