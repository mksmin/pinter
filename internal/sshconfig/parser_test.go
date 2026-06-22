package sshconfig

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	input := `
Host *
  User ignored

Host prod
  HostName 10.0.0.1
  User deploy
  Port 2222
  IdentityFile ~/.ssh/prod

Host staging-*
  User nobody

Host dev
  HostName dev.local # comment
`

	hosts, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d: %#v", len(hosts), hosts)
	}
	if hosts[0].Alias != "prod" || hosts[0].Hostname != "10.0.0.1" || hosts[0].Username != "deploy" || hosts[0].Port != 2222 {
		t.Fatalf("unexpected prod host: %#v", hosts[0])
	}
	if hosts[1].Alias != "dev" || hosts[1].Hostname != "dev.local" || hosts[1].Username != "root" || hosts[1].Port != 22 {
		t.Fatalf("unexpected dev host: %#v", hosts[1])
	}
}

func TestParseEqualsSyntax(t *testing.T) {
	hosts, err := Parse(strings.NewReader(`
Host=box
HostName=box.local
User=me
Port=2200
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Alias != "box" || hosts[0].Hostname != "box.local" || hosts[0].Username != "me" || hosts[0].Port != 2200 {
		t.Fatalf("unexpected host: %#v", hosts[0])
	}
}
