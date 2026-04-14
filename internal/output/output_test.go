package output

import (
	"bytes"
	"strings"
	"testing"

	"mdnsmap/internal/model"
)

func TestWriteTextDeterministic(t *testing.T) {
	res := model.DiscoveryResult{
		Services: []model.ServiceRecord{
			{InstanceName: "nas", ServiceShort: "http", ServiceType: "_http._tcp.local", Port: 5000, HostName: "nas.local", IPv4: []string{"192.168.1.2"}, TTL: 10, TXT: map[string]string{"path": "/"}, TXTOrder: []string{"path"}},
		},
		PTRAnswers: []string{"_http._tcp.local"},
	}
	var b bytes.Buffer
	if err := WriteText(&b, res); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"services:", "5000/tcp http:", "Name=nas", "answers:", "PTR:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}
