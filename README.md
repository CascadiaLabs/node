# Cascadia Node

Нода каскадной сети Cascadia: контейнер sing-box, которым управляет панель
Cascadia по gRPC (TLS с пиннингом сертификата).

## Почему Cascadia Node

- **Крошечная.** Один контейнер со статичным бинарником sing-box. Реальный замер
  на VPS 512MB RAM / 5GB диска / 1 vCPU (10% нагрузки):

  ```
  NAME   CPU %   MEM USAGE
  node   0.6%    6.297MiB
  ```

  ~7MiB RAM, образ на диске — 18.1MB, CPU простаивает и лишь изредка
  всплескивает до 0.6%. Ноду можно ставить даже на сервер, где уже занято
  почти всё.
- **sing-box.** Нативные конфиги: vless/reality, vmess, trojan, shadowsocks,
  hysteria2, tuic. Без Xray-совместимости и обвязок.
- **Управляется только панелью.** Интерфейса на ноде нет — конфиг пушится по
  gRPC из [Cascadia Panel](https://github.com/CascadiaLabs/panel), что и позволяет
  строить каскадные сети визуально, на графе.

## Install

```
sudo bash -c "$(curl -sL https://github.com/CascadiaLabs/install/raw/main/node.sh)" @ install
```

Поднимает контейнер на :6237 (gRPC поверх TLS). API-токен печатается в конце
установки — вставьте его вместе с сертификатом (`cert.pem`) при регистрации ноды
в панели.

## Релизы и версионирование

Имя git-тега = имя Docker-тега (`1a`, `1b` — альфа, бета релизы, `1`, `2`, `3` —
стабильные). Пуш тега публикует образ той же версии в GHCR
(`ghcr.io/cascadialabs/node:1a`); `:latest` — `main`. Версия ноды и sing-box
печатаются при старте: `docker logs node 2>&1 | head -1`.