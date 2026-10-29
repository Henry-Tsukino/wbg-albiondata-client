### Настройка клиента

Настройки, в частности закрытие консоли клиента, можно найти в системном трее Windows в правом нижнем углу экрана:

<img width="444" height="208" alt="Скриншот системного трея в среде Windows 11" src="https://github.com/user-attachments/assets/b1f7559d-f653-4447-8a7a-6c3044181f24" />

* **Hide Console** — Спрятать консоль
* **Start on Windows** — Запуск приложения при старте Windows
* **Quit** — Выключить программу

> [!NOTE]
> **Рекомендуется включить**, чтоб клиент не приходилось включать при каждом запуске игры

> [!CAUTION]
> В случае, **если ваше устройсво крайне слабое** и работа в фоне, даже такого маленького, приложения вам критична для производительности, **стоит оставить данную настройку выключенной**
## Установка / Installation

**Windows:** скачайте `installer-WBG-tool.exe` и запустите. Установщик сам поставит Npcap (нужен для захвата трафика).
**Windows:** download `installer-WBG-tool.exe` and run it. It installs Npcap (required for packet capture).

**Linux:** скачайте `GNU-Linux.tar.gz`, распакуйте и запустите `bash ./install.sh` (подробности в `инструкция.md` внутри архива).
**Linux:** download `GNU-Linux.tar.gz`, extract it and run `bash ./install.sh`.
Или одной командой / or in one command:
```
bash -c 'd=$(mktemp -d) && { curl -fsSL https://github.com/Henry-Tsukino/wbg-albiondata-client/releases/latest/download/GNU-Linux.tar.gz || wget -qO- https://github.com/Henry-Tsukino/wbg-albiondata-client/releases/latest/download/GNU-Linux.tar.gz; } | tar -xz -C "$d" && bash "$d/install.sh"; rm -rf "$d"'
```

**macOS:** `WBG-albion-data-client-amd64-mac.zip`

## Обновление / Updating

Уже установленный клиент обновляется сам (файлы `update-*.gz` нужны только для автообновления, скачивать их вручную не надо).
An installed client updates itself automatically (the `update-*.gz` files are for the auto-updater only).

---
