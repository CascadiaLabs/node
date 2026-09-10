package main

// defaultConfigJSON — стартовый каскадный конфиг, если файла персистентности ещё нет.
// Формат — нативный sing-box JSON (vless-in -> next-hop-out).
const defaultConfigJSON = `{
  "log": {
    "level": "info"
  },
  "inbounds": [
    {
      "type": "vless",
      "tag": "vless-in",
      "listen": "0.0.0.0",
      "listen_port": 443,
      "users": [
        {
          "name": "user-1",
          "uuid": "b831381d-6324-4d53-ad4f-8cda48b30811"
        }
      ]
    }
  ],
  "outbounds": [
    {
      "type": "vless",
      "tag": "next-hop-out",
      "server": "192.168.1.50",
      "server_port": 443,
      "uuid": "c942492e-7435-5e64-be5f-9deb59c41922"
    },
    {
      "type": "direct",
      "tag": "direct"
    }
  ],
  "route": {
    "rules": [
      {
        "inbound": ["vless-in"],
        "action": "route",
        "outbound": "next-hop-out"
      }
    ]
  }
}
`
