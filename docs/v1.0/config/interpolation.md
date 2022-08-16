---
layout: default
parent: Configure
title: Interpolation
nav_order: 6
---

# Interpolation

Blip interpolates environment variables and monitor variables in the [config file](config-file) and [plan files](../plans/file).

Environment variable
: `${FOO}`

Environment variable with default value
: `${FOO:-default}`

Monitor variable
: `%{monitor.VAR}`

{: .note }
**NOTE**: `${}` and `%{}` are always required.

Environment variable interpolation is a simple implementation of the shell standard.
In Blip, only the two cases shown above are supported, and `default` must be a literal value (it cannot be another `${}`).

Monitor variables are scoped to (only work within) a single monitor.
For example:

```yaml
monitors:
  - hostname: db.local
    username: metrics
    tags:
      hostname: %{monitor.hostname}
```

The result is `monitors.tags.hostname = "db.local"` because `%{monitor.hostname}` refers to the local `monitors.hostname` variable.
Blip is remarkably flexible, so this works the other way, too:

```yaml
monitors:
  - hostname: %{monitor.tags.hostname}
    username: metrics
    tags:
      hostname: db.local
```

The result is `monitors.hostname = "db.local"` because `%{monitor.tags.hostname}` refers to the local `monitors.tags.hostname` variable.

{: .note }
Singular "monitor" in `%{monitor.VAR}`, not plural, to emphasize that the reference is only to the single monitor in which it appears

You can use both in a single value, like:

```yaml
tls:
  ca: "${SECRETS_DIR}/%{monitor.hostname}"

monitors:
  - hostname: db1
  - hostname: db2
```

Top-level `tls.ca` specifies a monitor default that applies to all monitors that don't explicitly set the variable.
If `SECRETS_DIR = /secrets`, the result is:

```yaml
monitors:
  - hostname: db1
    tls:
      ca: /secrets/db1
  - hostname: db2
    tls:
      ca: /secrets/db2
```
