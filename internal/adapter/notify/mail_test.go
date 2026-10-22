package notify

import (
	"net/smtp"
	"testing"
)

func TestAutoAuthPicksLoginWhenPlainIsAbsent(t *testing.T) {
	a := &autoAuth{user: "noreply@example.com", password: "pw", host: "smtp.example.com"}
	mech, _, err := a.Start(&smtp.ServerInfo{Name: "smtp.example.com", TLS: true, Auth: []string{"LOGIN", "XOAUTH2"}})
	if err != nil || mech != "LOGIN" {
		t.Fatalf("Start = %q, %v; want LOGIN", mech, err)
	}
	for challenge, want := range map[string]string{"Username:": "noreply@example.com", "Password:": "pw"} {
		got, err := a.Next([]byte(challenge), true)
		if err != nil || string(got) != want {
			t.Fatalf("Next(%q) = %q, %v; want %q", challenge, got, err, want)
		}
	}
	if _, err := a.Next([]byte("Something:"), true); err == nil {
		t.Fatal("an unknown challenge must fail")
	}
}

func TestAutoAuthPrefersPlain(t *testing.T) {
	a := &autoAuth{user: "u", password: "p", host: "smtp.example.com"}
	mech, resp, err := a.Start(&smtp.ServerInfo{Name: "smtp.example.com", TLS: true, Auth: []string{"LOGIN", "PLAIN"}})
	if err != nil || mech != "PLAIN" || string(resp) != "\x00u\x00p" {
		t.Fatalf("Start = %q, %q, %v; want PLAIN", mech, resp, err)
	}
}

func TestAutoAuthRefusesCleartextToRemote(t *testing.T) {
	a := &autoAuth{user: "u", password: "p", host: "smtp.example.com"}
	if _, _, err := a.Start(&smtp.ServerInfo{Name: "smtp.example.com", TLS: false, Auth: []string{"LOGIN"}}); err == nil {
		t.Fatal("credentials must not go over an unencrypted link")
	}
}

func TestMessageIDDomain(t *testing.T) {
	for in, want := range map[string]string{"noreply@sunshinefood.co.id": "sunshinefood.co.id", "SFG <a@b.co>": "b.co", "nobody": "localhost"} {
		if got := messageIDDomain(in); got != want {
			t.Errorf("messageIDDomain(%q) = %q, want %q", in, got, want)
		}
	}
}
