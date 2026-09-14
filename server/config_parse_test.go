//go:build with_utls && with_quic

package main

import (
	"context"
	"testing"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

// Проверяем, что конфиги, сгенерированные панелью (graph.Generate),
// принимаются sing-box 1.14 (парсинг + построение инстанса, без Start).
// Запуск: go test -tags "with_utls,with_quic" ./server/
func parseBuild(t *testing.T, raw string) option.Options {
	t.Helper()
	ctx, cancel := context.WithCancel(include.Context(context.Background()))
	defer cancel()
	opts, err := json.UnmarshalExtendedContext[option.Options](ctx, []byte(raw))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, err := box.New(box.Options{Context: ctx, Options: opts}); err != nil {
		t.Fatalf("box.New: %v", err)
	}
	return opts
}

const nodeARealityCascade = `{
  "inbounds": [
    {
      "type": "vless",
      "tag": "in-a",
      "listen": "::",
      "listen_port": 443,
      "users": [{"name": "user", "uuid": "11111111-1111-1111-1111-111111111111"}],
      "tls": {
        "enabled": true,
        "server_name": "a.example.com",
        "reality": {
          "enabled": true,
          "handshake": {"server": "www.microsoft.com", "server_port": 443},
          "private_key": "pX1etqxgOUimOsJhIT3J0g9Gcij2kcV_yPjTLcC_f9U",
          "short_id": ["abcd1234"]
        }
      }
    }
  ],
  "outbounds": [
    {"type": "vless", "tag": "out-a", "server": "b.example.com", "server_port": 8443,
     "uuid": "22222222-2222-2222-2222-222222222222",
     "tls": {"enabled": true, "server_name": "b.example.com"}},
    {"type": "direct", "tag": "cascadia-direct"}
  ],
  "route": {
    "final": "cascadia-direct",
    "rules": [{"inbound": ["in-a"], "outbound": "out-a"}]
  }
}`

const nodeBSelfSignedTLS = `{
  "inbounds": [
    {"type": "vless", "tag": "in-b", "listen": "::", "listen_port": 8443,
     "users": [{"name": "user", "uuid": "22222222-2222-2222-2222-222222222222"}],
     "tls": {"enabled": true, "server_name": "b.example.com", "insecure": true}}
  ],
  "outbounds": [{"type": "direct", "tag": "out-b"}],
  "route": {"final": "out-b", "rules": [{"inbound": ["in-b"], "outbound": "out-b"}]}
}`

const urltestBalancer = `{
  "inbounds": [
    {"type": "vless", "tag": "in1", "listen": "::", "listen_port": 1000,
     "users": [{"name": "u", "uuid": "11111111-1111-1111-1111-111111111111"}]}
  ],
  "outbounds": [
    {"type": "urltest", "tag": "bal", "outbounds": ["o1"], "url": "https://www.gstatic.com/generate_204", "interval": "3m", "tolerance": 50},
    {"type": "vless", "tag": "o1", "server": "x.example.com", "server_port": 2000, "uuid": "u2",
     "tls": {"enabled": true, "server_name": "x.example.com", "utls": {"enabled": true, "fingerprint": "chrome"},
             "reality": {"enabled": true, "public_key": "jNXHt1yRoQvV45jy5F5GGRWg7tWPdRkSHLcTt0-0wQ8", "short_id": "dead"}}},
    {"type": "direct", "tag": "cascadia-direct"}
  ],
  "route": {"final": "cascadia-direct",
            "rules": [{"inbound": ["in1"], "domain_suffix": [".ru"], "network": ["tcp"], "outbound": "bal"}]}
}`

const hysteria2TUIC = `{
  "inbounds": [
    {"type": "hysteria2", "tag": "in-h", "listen": "::", "listen_port": 3000,
     "up_mbps": 100, "down_mbps": 500, "obfs": {"type": "salamander", "password": "obfs-pw"},
     "users": [{"name": "u", "password": "pw"}], "tls": {"enabled": true, "server_name": "h.example.com", "insecure": true}},
    {"type": "tuic", "tag": "in-t", "listen": "::", "listen_port": 4000, "congestion_control": "bbr",
     "users": [{"name": "u", "uuid": "33333333-3333-3333-3333-333333333333", "password": "tpw"}], "tls": {"enabled": true, "server_name": "t.example.com", "insecure": true}}
  ],
  "outbounds": [
    {"type": "hysteria2", "tag": "o-h", "server": "h2.example.com", "server_port": 5000, "password": "pw2",
     "obfs": {"type": "salamander", "password": "obfs-2"}, "tls": {"enabled": true, "insecure": true}},
    {"type": "tuic", "tag": "o-t", "server": "t.example.com", "server_port": 4000, "uuid": "33333333-3333-3333-3333-333333333333", "password": "tpw",
     "congestion_control": "bbr", "tls": {"enabled": true, "insecure": true}},
    {"type": "direct", "tag": "od1"}
  ],
  "route": {"final": "od1", "rules": [
    {"inbound": ["in-h"], "outbound": "o-h"},
    {"inbound": ["in-h"], "outbound": "od1"},
    {"inbound": ["in-t"], "outbound": "od1"}
  ]}
}`

func TestPanelGeneratedConfigsParse(t *testing.T) {
	for name, cfg := range map[string]string{
		"reality cascade A": nodeARealityCascade,
		"self-signed B":     nodeBSelfSignedTLS,
		"urltest balancer":  urltestBalancer,
		"hysteria2 + tuic":  hysteria2TUIC,
	} {
		t.Run(name, func(t *testing.T) { parseBuild(t, cfg) })
	}
}
