# MCP Tools Audit (Gezielte Aufteilung + Parameter-Referenz)

Stand: 2026-04-15
Quelle: `src/backend/internal/mcp/tools.go`, Dispatcher in `src/backend/internal/mcp/server.go`

## Ziel dieses Audits

1. Prüfen, ob Tools weiter aufgeteilt werden sollten, damit KI-Agents gezielter abfragen/aufrufen koennen.
2. Alle Parameter praezise dokumentieren.
3. Zu jedem bool-Parameter die Auswirkungen von `true` und `false` explizit machen.

---

## Zusammenfassung der Split-Empfehlungen

### Bereits gut aufgeteilt (beibehalten)

- Container: `list`, `inspect`, `create`, `start`, `stop`, `restart`, `remove`, `logs`, `exec`, `stats`
- Images: `list`, `pull`, `build`, `tag`, `inspect`, `remove`
- Volumes: `list`, `create`, `inspect`, `remove`
- Networks: `list`, `create`, `inspect`, `connect`, `disconnect`, `remove`
- Compose: `up`, `down`, `ps`, `logs`, `pull`
- System: `info`, `version`, `df`, `prune_*`

### Kandidaten fuer weitere Aufteilung (optional, nicht zwingend)

1. `docker_container_logs`
- Optional split in `docker_container_logs_tail` und `docker_container_logs_since` fuer deterministischere Agent-Aufrufe.

2. `docker_container_exec`
- Optional split in `docker_container_exec_simple` (nur `id`, `command`) und `docker_container_exec_advanced` (`user`, `workdir`).

3. `docker_compose_up`
- Optional split in `docker_compose_up_simple` und `docker_compose_up_build`.

4. `docker_image_build`
- Optional split in `docker_image_build_simple` und `docker_image_build_advanced` (mit `build_args`, `dockerfile`, `no_cache`).

5. `docker_system_prune_all`
- Bereits durch `docker_system_prune_*` sauber ergaenzt. Kein weiterer Split erforderlich.

Empfehlung: Aktuelle Aufteilung ist fuer MCP/Agent-Workflows bereits sehr gut. Weitere Splits nur einfuehren, wenn Telemetrie zeigt, dass Agents regelmaessig falsche Parameterkombinationen waehlen.

---

## Detaillierte Tool-Referenz

## Containers

### `docker_container_list`
- Zweck: Container auflisten.
- Parameter:
  - `all` (bool, optional, default `false`):
    - `false`: nur laufende Container.
    - `true`: auch gestoppte/exited Container.
- Beispiel:
```json
{"all": true}
```

### `docker_container_inspect`
- Zweck: Vollstaendige Detailansicht eines Containers.
- Parameter:
  - `id` (string, required): Container-ID oder Name.
- Beispiel:
```json
{"id":"web-1"}
```

### `docker_container_create`
- Zweck: Container erstellen (ohne Start).
- Parameter:
  - `image` (string, required): z. B. `nginx:latest`.
  - `name` (string, optional): z. B. `my-nginx`.
  - `command` (string, optional): z. B. `sleep infinity`.
  - `env` (string[], optional): `KEY=VALUE`.
  - `ports` (string[], optional): `HOST:CONTAINER`, z. B. `8080:80`.
  - `volumes` (string[], optional): `HOST:CONTAINER`, z. B. `/data:/var/lib/app`.
  - `network` (string, optional): z. B. `bridge`.
  - `restart` (string, optional): `no`, `always`, `unless-stopped`, `on-failure`.
- Beispiel:
```json
{
  "image":"nginx:latest",
  "name":"web-1",
  "ports":["8080:80"],
  "restart":"unless-stopped"
}
```

### `docker_container_start`
- Zweck: Gestoppten Container starten.
- Parameter: `id` (string, required).

### `docker_container_stop`
- Zweck: Laufenden Container stoppen.
- Parameter:
  - `id` (string, required)
  - `timeout` (integer, optional, default `10`): Sekunden bis hartes Kill-Signal.
- Beispiel:
```json
{"id":"web-1","timeout":20}
```

### `docker_container_restart`
- Zweck: Container neu starten.
- Parameter wie `docker_container_stop`.

### `docker_container_remove`
- Zweck: Container loeschen.
- Parameter:
  - `id` (string, required)
  - `force` (bool, optional, default `false`):
    - `false`: nur entfernbare (typisch gestoppte) Container.
    - `true`: entfernt auch laufende Container (forciert).
  - `volumes` (bool, optional, default `false`):
    - `false`: zugeordnete anonyme Volumes bleiben.
    - `true`: anonyme Volumes mit entfernen.
- Beispiel:
```json
{"id":"web-1","force":true,"volumes":true}
```

### `docker_container_logs`
- Zweck: Logs eines Containers lesen.
- Parameter:
  - `id` (string, required)
  - `tail` (string, optional, default `100`): letzte N Zeilen.
  - `timestamps` (bool, optional, default `false`):
    - `false`: ohne Zeitstempel.
    - `true`: jede Zeile mit Zeitstempel.
  - `since` (string, optional): Zeitgrenze, z. B. `1h` oder ISO-Zeitpunkt.
- Beispiel:
```json
{"id":"web-1","tail":"200","timestamps":true,"since":"1h"}
```

### `docker_container_exec`
- Zweck: Kommando in laufendem Container ausfuehren.
- Parameter:
  - `id` (string, required)
  - `command` (string, required)
  - `user` (string, optional): z. B. `root`, `1000:1000`.
  - `workdir` (string, optional): z. B. `/app`.
- Beispiel:
```json
{"id":"web-1","command":"ls -la /","user":"root"}
```

### `docker_container_stats`
- Zweck: Laufende Ressourcenmetriken eines Containers.
- Parameter: `id` (string, required).

---

## Images

### `docker_image_list`
- Zweck: Lokale Images auflisten.
- Parameter:
  - `all` (bool, optional):
    - `false`: Standardliste.
    - `true`: inklusive Intermediate-Images.
- Beispiel:
```json
{"all":true}
```

### `docker_image_pull`
- Zweck: Image aus Registry ziehen.
- Parameter:
  - `image` (string, required): z. B. `redis:7`.
  - `platform` (string, optional): z. B. `linux/arm64`.
- Beispiel:
```json
{"image":"redis:7","platform":"linux/amd64"}
```

### `docker_image_build`
- Zweck: Image aus Build-Context bauen.
- Parameter:
  - `context_path` (string, required)
  - `dockerfile` (string, optional, default `Dockerfile`)
  - `tag` (string, optional)
  - `no_cache` (bool, optional, default `false`):
    - `false`: Build-Cache verwenden.
    - `true`: ohne Cache bauen (langsamer, reproduzierbarer in manchen Faellen).
  - `build_args` (string[], optional): `KEY=VALUE`.
- Beispiel:
```json
{
  "context_path":"C:/repo/app",
  "dockerfile":"Dockerfile.prod",
  "tag":"myapp:prod",
  "no_cache":true,
  "build_args":["APP_ENV=prod","COMMIT_SHA=abc123"]
}
```

### `docker_image_tag`
- Zweck: Tag fuer vorhandenes Image setzen.
- Parameter:
  - `source` (string, required)
  - `target` (string, required)
- Beispiel:
```json
{"source":"myapp:latest","target":"registry.example.com/myapp:1.2.3"}
```

### `docker_image_inspect`
- Zweck: Detaillierte Image-Metadaten lesen.
- Parameter: `image` (string, required).

### `docker_image_remove`
- Zweck: Lokales Image entfernen.
- Parameter:
  - `image` (string, required)
  - `force` (bool, optional, default `false`):
    - `false`: nur wenn keine Konflikte.
    - `true`: Entfernen forcieren.
- Beispiel:
```json
{"image":"redis:7","force":true}
```

---

## Volumes

### `docker_volume_list`
- Zweck: Alle Volumes auflisten.
- Parameter: keine.

### `docker_volume_create`
- Zweck: Volume erstellen.
- Parameter:
  - `name` (string, optional)
  - `driver` (string, optional, default `local`)
- Beispiel:
```json
{"name":"pgdata","driver":"local"}
```

### `docker_volume_inspect`
- Zweck: Volume-Details lesen.
- Parameter: `name` (string, required).

### `docker_volume_remove`
- Zweck: Volume entfernen.
- Parameter:
  - `name` (string, required)
  - `force` (bool, optional, default `false`):
    - `false`: nur wenn nicht in Benutzung.
    - `true`: forciertes Entfernen.
- Beispiel:
```json
{"name":"pgdata","force":true}
```

---

## Networks

### `docker_network_list`
- Zweck: Docker-Netzwerke auflisten.
- Parameter: keine.

### `docker_network_create`
- Zweck: Netzwerk erstellen.
- Parameter:
  - `name` (string, required)
  - `driver` (string, optional, default `bridge`)
  - `subnet` (string, optional): CIDR, z. B. `172.28.0.0/16`.
- Beispiel:
```json
{"name":"app-net","driver":"bridge","subnet":"172.28.0.0/16"}
```

### `docker_network_inspect`
- Zweck: Netzwerkdetails lesen.
- Parameter: `name` (string, required).

### `docker_network_connect`
- Zweck: Container an Netzwerk anbinden.
- Parameter:
  - `network` (string, required)
  - `container` (string, required)
- Beispiel:
```json
{"network":"app-net","container":"web-1"}
```

### `docker_network_disconnect`
- Zweck: Container vom Netzwerk trennen.
- Parameter:
  - `network` (string, required)
  - `container` (string, required)
  - `force` (bool, optional, default `false`):
    - `false`: normale Trennung, kann bei Konflikten fehlschlagen.
    - `true`: Trennung erzwingen.
- Beispiel:
```json
{"network":"app-net","container":"web-1","force":true}
```

### `docker_network_remove`
- Zweck: Netzwerk loeschen.
- Parameter: `name` (string, required).

---

## Compose

### `docker_compose_up`
- Zweck: Compose-Services starten.
- Parameter:
  - `project_dir` (string, required)
  - `services` (string[], optional): leer/omitted = alle Services.
  - `detach` (bool, optional, default `true`):
    - `true`: Hintergrundmodus.
    - `false`: foreground/blockierend.
  - `build` (bool, optional, default `false`):
    - `false`: vorhandene Images verwenden.
    - `true`: vor Start neu bauen.
  - `force_recreate` (bool, optional, default `false`):
    - `false`: nur wenn erforderlich neu erstellen.
    - `true`: Container immer neu erstellen.
- Beispiel:
```json
{
  "project_dir":"C:/repo/stack",
  "services":["api","worker"],
  "detach":true,
  "build":true,
  "force_recreate":false
}
```

### `docker_compose_down`
- Zweck: Compose-Stack stoppen/abbauen.
- Parameter:
  - `project_dir` (string, required)
  - `volumes` (bool, optional, default `false`):
    - `false`: Volumes behalten.
    - `true`: benannte/anonyme Volumes entfernen.
  - `remove_orphans` (bool, optional, default `false`):
    - `false`: Orphans belassen.
    - `true`: Orphan-Container entfernen.
- Beispiel:
```json
{"project_dir":"C:/repo/stack","volumes":true,"remove_orphans":true}
```

### `docker_compose_ps`
- Zweck: Compose-Containerstatus auflisten.
- Parameter: `project_dir` (string, required).

### `docker_compose_logs`
- Zweck: Compose-Service-Logs lesen.
- Parameter:
  - `project_dir` (string, required)
  - `services` (string[], optional): leer/omitted = alle.
  - `tail` (string, optional, default `100`)
- Beispiel:
```json
{"project_dir":"C:/repo/stack","services":["api"],"tail":"300"}
```

### `docker_compose_pull`
- Zweck: Compose-Images pullen.
- Parameter:
  - `project_dir` (string, required)
  - `services` (string[], optional): leer/omitted = alle.
- Beispiel:
```json
{"project_dir":"C:/repo/stack","services":["api","worker"]}
```

---

## System

### `docker_system_info`
- Zweck: Host-/Daemon-Metadaten (CPU, RAM, OS, Driver, Counts).
- Parameter: keine.

### `docker_system_version`
- Zweck: Client-/Server-Versionen fuer Kompatibilitaetspruefung.
- Parameter: keine.

### `docker_system_df`
- Zweck: Disk-Usage-UEbersicht inkl. reclaimable.
- Parameter: keine.

### `docker_system_prune_all`
- Zweck: Alle ungenutzten Ressourcen bereinigen.
- Parameter:
  - `filters` (string[], optional): Docker-Prune-Filterausdruecke.
- Bool-Parameter: keine.
- Beispiel:
```json
{"filters":["until=24h","label!=keep"]}
```

### `docker_system_prune_containers`
- Zweck: Nur ungenutzte Containerressourcen bereinigen.
- Parameter:
  - `filters` (string[], optional)
- Beispiel:
```json
{"filters":["until=168h","label=env=dev"]}
```

### `docker_system_prune_images`
- Zweck: Nur ungenutzte Images bereinigen.
- Parameter:
  - `filters` (string[], optional)
- Beispiel:
```json
{"filters":["until=720h","label!=retain"]}
```

### `docker_system_prune_networks`
- Zweck: Nur ungenutzte Netzwerke bereinigen.
- Parameter:
  - `filters` (string[], optional)
- Beispiel:
```json
{"filters":["until=48h","label=scope=temp"]}
```

### `docker_system_prune_build_cache`
- Zweck: Nur Build-Cache bereinigen.
- Parameter:
  - `filters` (string[], optional)
- Beispiel:
```json
{"filters":["until=72h"]}
```

### `docker_system_prune_volumes`
- Zweck: Nur ungenutzte Volumes bereinigen.
- Parameter:
  - `filters` (string[], optional)
- Beispiel:
```json
{"filters":["label=cleanup=true","label!=keep"]}
```

---

## Empfehlungen fuer naechste Iteration

1. `tools.go` Descriptions noch einheitlicher machen:
- Ueberall gleiche Struktur: "What it does" + "When to use" + "Parameter constraints" + "Example".

2. Optional: `dry_run` fuer alle `docker_system_prune_*` Tools.
- Heute nicht vorhanden.
- Wuerde Agenten erlauben, erst Impact abzuschaetzen, dann aufzuraeumen.

3. Optional: `docker_system_prune_plan` Tool einfuehren.
- Liefert nur Kandidaten + erwartete freizugebende Groesse.
- Keine Destruktion, gut fuer sichere Agent-Workflows.
