package banner

import "testing"

func TestParseTXT(t *testing.T) {
	txt := []string{"model=TS-X64", "fwVer=5.2.9", "flag"}
	m, order := ParseTXT(txt)
	if m["model"] != "TS-X64" || m["flag"] != "" {
		t.Fatalf("unexpected map: %#v", m)
	}
	if len(order) != 3 {
		t.Fatalf("unexpected order: %#v", order)
	}
}
