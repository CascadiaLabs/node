package main

// defaultConfigJSON — стартовый базовый конфиг без внешних и внутренних соединений.
// Формат — нативный sing-box JSON.
const defaultConfigJSON = `{
  "log": {
    "level": "info"
  },
  "inbounds": [],
  "outbounds": [],
  "route": {
    "rules": []
  }
}
`
