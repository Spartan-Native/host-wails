package appearance

import "testing"

func TestResolveSkin(t *testing.T) {
	v15, v26 := "15.6.1", "26.1.0"
	b10, b11 := 19045, 22631
	kde, gnome := "kde", "gnome"
	cases := []struct {
		os      string
		version *string
		build   *int
		desktop *string
		want    string
	}{
		{"windows", nil, &b11, nil, "win"},
		{"windows", nil, &b10, nil, "w10"},
		{"macos", &v26, nil, nil, "mac"},
		{"macos", &v15, nil, nil, "macc"},
		{"linux", nil, nil, &gnome, "gtk"},
		{"linux", nil, nil, &kde, "qt"},
		{"plan9", nil, nil, nil, "web"},
	}
	for _, c := range cases {
		if got := resolveSkin(c.os, c.version, c.build, c.desktop); got != c.want {
			t.Errorf("resolveSkin(%s) = %s, want %s", c.os, got, c.want)
		}
	}
}

func TestHelpers(t *testing.T) {
	if got := rgbHex(0, 0.478, 1); got != "#007aff" {
		t.Errorf("rgbHex = %s", got)
	}
	for in, want := range map[string]string{"ubuntu:GNOME": "gnome", "KDE": "kde", "X-Cinnamon": "cinnamon", "XFCE": "xfce", "sway": "other"} {
		if got := linuxDesktop(in); got != want {
			t.Errorf("linuxDesktop(%q) = %s, want %s", in, got, want)
		}
	}
	if roundScale(1.249) != 1.25 || roundScale(0) != 1 {
		t.Error("roundScale")
	}
}
