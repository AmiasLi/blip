---
layout: config
title: "Config File"
parent: Configure
---

# Config File

Blip configuration is specified in a single [config file](blip#config-file) that has three sections:

```yaml
---
# Server config

# Monitor defaults

# Monitors
monitors: []
```
<br>

_Server config_
: Server config is top-level sections that configure the Blip binary (`blip`), which is called the _server_ because it runs the monitors and has an external [API](../api).
Server config is optional.

_Monitor defaults_
: Monitor defaults are top-level sections that configure default values for each monitor: hostname, username, and so on.
Monitor defaults are commonly configured to avoid repeating the same config for each monitor.

_Monitors_
: Monitors are a list of MySQL instances to monitor.
Each monitor in the list inherits config from the monitor defaults.
For example, if a monitor does not explicitly set the MySQL hostname, it inherits the value from monitor defaults if set.

<br>
The simplest possible Blip config for monitoring a single MySQL instance looks like:

```yaml
monitors:
  - hostname: db.local
    username: blip
    password: "foo"
```

That uses the [default plan](../plans/defaults) to collect metrics.
You will likely write your own custom plan, which is configured like:

```yaml
plans:
  files: my-plan.yaml
monitors:
  - hostname: db.local
    username: blip
    password: "bar"
```

In this case, the monitor inherits the only plan.
See [Plans / Loading](../plans/loading) for details on how plans are loaded and shared.

# Conventions

To reference sections, subsections, and specific user-configurable variables within those sections:

`SECTION`
: Any top-level section. For example: `api`, `plans`, and `monitors`.

`SECTION[.SUB...].VAR`
: A user-configurable variable in the `SECTION` with optional subsections (`SUB`). For example: `mysql.username` (section = `mysql`, variable = `username`), or `monitors.heartbeat.freq` (section = `monitors`, subsection = `heartbeat`, variable = `freq`).

Note the difference between `heartbeat` and `monitors.heartbeat`: the former is the top-level [monitor default](#monitor-defaults); the latter is a [monitor config](#monitors) that inherts the monitor defaults, if any.

In the Blip documentation outside this page, config file references begin with `config.`.
For example, `config.api` refers to the [`api`](#api) server config section of the Blip config file.

`disable` and `disable-auto-...` are used to disable features.
There are no "enable" prefixes or variables.
Instead, some features of either off or auto by default.
If off by default, the feature is enabled by specifying a variable noted in the docs.
For example, `heartbeat` is off by default and enabled when `heartbeat.freq` is specified.
If auto by default, the feature is disabled by specifying `disable-auto-FEATURE: true`, where `FEATURE` is the feature name.
For example, `aws.disable-auto-region: true` to disable auto-detecting the AWS region.

Blip uses `lowercase-kebab-case` for all sections and variable names.

---

# Server Config

{: .config-section-title}
## api

The `api` section configures the [Blip API](../api/).

```yaml
api:
  bind: "127.0.0.1:7522"
  disable: false
```

### `bind`

{: .var-table }
|**Type**|string|
|**Valid values**|`addr:port`, `:port`|
|**Default value**|`127.0.0.1:7522`|

The `bind` variable sets the interface address and port that the API listens on.

### `disable`

{: .var-table }
|**Type**|bool|
|**Valid values**|`true` or `false`|
|**Default value**|`false`|

The `disable` variable disables the Blip API.

{: .config-section-title}
## monitor-loader

The `monitor-loader` section configures how Blip finds and loads MySQL instances.

```yaml
monitor-loader:
  aws:
    regions: []
  files: []
  local:
    disable-auto: false
    disable-auto-root: false
  stop-loss: ""
```

### aws

The `aws` subsection of the `monitor-loader` section configure built-in support for loading Amazon RDS instances.
By default, this feature is disabled.
To enable, specify `regions`.

### `regions`

{: .var-table }
|**Type**|list of strings|
|**Valid values**|AWS region names or "auto" to auto-detect|
|**Default value**||

The `regions` variable sets which AWS regions to query for RDS instances.
If `auto` is specified, Blip queries [EC2 IMDS](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html), which only works if Blip is running on an EC2 instance with an [EC2 instance profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html) that allows [rds:DescribeDBInstances](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DescribeDBInstances.html).

### `files`

{: .var-table }
|**Type**|list of strings|
|**Valid values**|file names|
|**Default value**||

The `files` variable specifies YAML files to load monitors from.
Each file must have a `monitors` section, like:

```yaml
---
monitors:
  - hostname: db.local
```

File paths are relative to the current working directory of `blip`.

### local

The `local` subsection has only two variables:

`disable-auto: true`
`disable-auto-root: true`

### `stop-loss`

{: .var-table }
|**Type**|string|
|**Valid values**|&bull;&nbsp;"N%" (percentage) where N is an integer between 0 and 100 (exclusive)<br>&bull;&nbsp;"N" where N is an integer greater than 0|
|**Default value**||

The `stop-loss` variable enables the [stop-lost feature](../monitors/loading#stop-loss).

---

# Monitor Defaults

Monitor defaults are top-level sections that set default values for monitors that do not set an explicit value.
Monitor defaults are useful when you have several MySQL instances to monitor and the configuration only differs by basic connection parameters, like hostname or socket.

For example, imagine that you have 10 monitors all with the same username and password.
Instead of setting `username` and `password` in all 10 monitors, you can set these variables once in the top-level `mysql` section:

```yaml
mysql:
  username: "defaultUser"
  password: "defaultPass"
monitors:
  - hostname: db1
  # ...
  - hostname: db10
```

The default `username` and `password` are applied to the 10 monitors because none of them explicitly set these variables.
If a monitor explicitly sets one of the variables, then its explicit value is used instead of the default value.

{: .note }
Monitor defaults are convenient, but explicit monitor configuration is more clear, so use monitor defaults sparingly.
The intended use case is for variables that _must_ be consistent for all monitors.
For example, if Blip monitors Amazon RDS instances in region `us-east-1`, then setting monitor default `aws.region: "us-east-1"` makes sense.

{: .config-section-title}
## aws

The `aws` section configures Amazon RDS for MySQL.

```yaml
aws:
  disable-auto-region: false
  disable-auto-tls: false
  iam-auth: false
  password-secret: ""
  region: ""
```

### `disable-auto-region`

{: .var-table }
|**Type**|string|
|**Valid values**|`true` or `false`|
|**Default value**|`false`|

The `disable-auto-region` variable enables/disables automatic detection of the AWS region.

### `disable-auto-tls`

{: .var-table }
|**Type**|string|
|**Valid values**|`true` or `false`|
|**Default value**|`false`|

The `disable-auto-tls` variables disables automatic use of the Amazon RDS certificate authority (CA).
By default, Blip uses the Amazon RDS CA-2019 certificate, which is built-in (you don't need to configure anything).
See [Cloud / AWS / TLS](../cloud/aws#tls) for details.

### `iam-auth`

{: .var-table }
|**Type**|bool|
|**Valid values**|`true` or `false`|
|**Default value**|`false`|

The `iam-auth` variable enables [IAM database authentication](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth).
When enabled, an IAM authentication token is generated by Blip and used as the password.

See [Cloud / AWS / IAM Authentication](../cloud/aws#iam-authentication) for details.

### `password-secret`

{: .var-table }
|**Type**|string|
|**Valid values**|AWS Secrets Manager ARN|
|**Default value**||

The `password-secret` variables sets the AWS Secrets Manager ARN that contains the MySQL user password.

### `region`

{: .var-table }
|**Type**|string|
|**Valid values**|"auto" or any valid AWS Region|
|**Default value**|"auth"|

The `region` variable sets the AWS region used to create an AWS configuration (which includes the AWS credentials).

See [Cloud / AWS / IAM Authentication](../cloud/aws#region) for details.

{: .config-section-title}
## exporter

The `exporter` section configure Blip to [emulate Prometheus `mysqld_exporter`](../prometheus).

```yaml
exporter:
  flags:
    web.listen-address: "127.0.0.1:9104"
    web.telemetry-path: "/metrics"
  mode: ""
```

### `flags`

{: .var-table }
|**Type**|key-value map (string: string)|
|**Valid values**|(see list below)|
|**Default value**|(see list below)|

The `flag` variable is a key-value map of strings for certain Prometheus mysqld_exporter flags:

* `web.listen-address` (default: `127.0.0.1:9104`)
* `web.telemetry-path` (default: `/metrics`)

### `mode`

{: .var-table }
|**Type**|string|
|**Valid values**|`dual` or `legacy`|
|**Default value**||

The `mode` variables enables [Prometheus emulation](../prometheus).
When set to `dual`, Blip runs normally _and_ emulates Prometheus.
When set to `legacy`, Blip runs _only_ emulates Prometheus.
The feature is disabled by default.

### `plan`

{: .var-table }
|**Type**|string|
|**Valid values**|Plan name|
|**Default value**|`default-exporter`|

The `plan` variables specifies which plan to load.
The plan must have only 1 level.
See [Prometheus emulation](../prometheus#plan) for details.

{: .config-section-title}
## heartbeat

The `heartbeat` section configures the [Blip heartbeat](../heartbeat).

```yaml
heartbeat:
  freq: ""
  role: ""
  source-id: ""
  table: blip.heartbeat
```

### `freq`

{: .var-table }
|**Type**|string|
|**Valid values**|[Go duration string](https://pkg.go.dev/time#ParseDuration) greater than zero|
|**Default value**||

The `freq` variable enables [Blip heartbeats](../hearbeat) at the specified frequency.
A frequency of "1s" or "2s" is suggested because heartbeat frequency does _not_ determine replication lag accuracy or reporting.
See [Heartbeat > Accuracy](../heartbeat#accuracy) for details.

See [`repl.lag` metric collector](../metrics/domains#repllag) for reporting replication lag.

To disable heartbeat, remove `freq` or set to an empty string (zero is not a valid value).

### `role`

{: .var-table }
|**Type**|string|
|**Valid values**|User-defined|
|**Default value**||

The `role` variable sets the role that the monitor reports in the [heartbeat table](../heartbeat#table).
For example, this might be a region like "us-east-1".

See [Heartbeat > Topology](../heartbeat#replication-topology) to learn how `role` and `source-id` are used.

### `source-id`

{: .var-table }
|**Type**|string|
|**Valid values**|User-defined|
|**Default value**|`%{monitor.id}`|

The `source-id` variable sets the source ID that the monitor reports in the [heartbeat table](../heartbeat#table).
This overrides the default value, which is often necessary in the cloud where MySQL instances do not have user-defined hostnames, especially with respect to replication.

See [Heartbeat > Topology](../heartbeat#replication-topology) to learn how `role` and `source-id` are used.

### `table`

{: .var-table }
|**Type**|string|
|**Valid values**|valid MySQL table name|
|**Default value**|`blip.heartbeat`|

The `table` variable sets the Blip heartbeat table (where heartbeat are written).
The default database is `blip` if the table name is not database-qualified like `db.heartbeat`.

The table must already exist; Blip does not create the table.
See [Heartbeat > Table](../heartbeat#table) for details.

{: .config-section-title}
## mysql

The `mysql` section configures how to connect to MySQL.

```yaml
mysql:
  hostname: ""
  mycnf: ""
  password: ""
  password-file: ""
  socket: ""
  timeout-connect: "10s"
  username: "blip"
```

As monitor defaults, this section is specified as shown above: top-level with the variables specified under `mysql:`.
For each monitor in the [`monitors`](#monitors) section, these variables are top-level (omit the `mysql:` header).
For example:

```yaml
monitors:
  - hostname: ""
    mycnf: ""
    # Other mysql section variables
```

_These are the only variables the become top-level in [`monitors`](#monitors)._

### `hostname`

{: .var-table }
|**Type**|string|
|**Valid values**|`hostname` or `hostname:port`|
|**Default value**||

The `hostname` variable sets the MySQL hostname.

### `mycnf`

{: .var-table }
|**Type**|string|
|**Valid values**|my.cnf file name|
|**Default value**||

The `mycnf` variable sets a my.cnf file to read.

Blip reads the `[client]` section of the my.cnf file:

|my.cnf File|Blip Variable|
|------|----|
|host|[`hostname`](#hostname)|
|password|[`password`](#password)|
|port|Appended to [`hostname`](#hostname)|
|socket|[`socket`](#socket)|
|ssl-ca|[`tls.ca`](#ca)|
|ssl-cert|[`tls.cert`](#cert)|
|ssl-key|[`tls.key`](#key)|
|ssl-mode|(Special handling)|
|user|[`username`](#username)|

### `username`

{: .var-table }
|**Type**|string|
|**Valid values**||
|**Default value**||

The `username` variable sets the MySQL username.

### `password`

{: .var-table }
|**Type**|string|
|**Valid values**||
|**Default value**||

The `password` variable sets the MySQL password.

### `password-file`

{: .var-table }
|**Type**|string|
|**Valid values**||
|**Default value**||

The `password-file` variable sets a file from which Blip reads the MySQL password.

### `socket`

{: .var-table }
|**Type**|string|
|**Valid values**||
|**Default value**||

The `socket` variable sets the MySQL socket.

### `timeout-connect`

{: .var-table }
|**Type**|string|
|**Valid values**|[Go duration string](https://pkg.go.dev/time#ParseDuration)|
|**Default value**||

The `timeout-connect` variable sets the connection timeout.

{: .config-section-title}
## plans

The `plans` section configures the source of [plans](../plans/).

```yaml
plans:
  files:
    - plan1.yaml
    - plan2.yaml
  table: "blip.plans"
  monitor: {}
  change:
    # See below
```

### change

The `change` subsection of the `plan` section configures [plan changing](../plans/changing) based on the state of MySQL.

```yaml
plans:
  change:
    offline:
      after: ""
      plan: ""
    standby:
      after: ""
      plan: ""
    read-only:
      after: ""
      plan: ""
    active:
      after: ""
      plan: ""
```

Each of the four sections&mdash;`offline`, `standby`, `read-only`, and `active`&mdash;have the same two variables:

#### `after`

{: .var-table }
|**Type**|string|
|**Valid values**|[Go duration string](https://pkg.go.dev/time#ParseDuration)|
|**Default value**||

The `after` variable sets how long before the state takes effect.

#### `plan`

{: .var-table }
|**Type**|string|
|**Valid values**|plan name|
|**Default value**||

The `plan` variable sets the plan to load when the state takes effect.

### `disable-default-plans`

{: .var-table }
|**Type**|string|
|**Valid values**|`true` or `false`|
|**Default value**|`false`|

The `disable-default-plans` variable enables/disables [default plans](../plans/loading#default).

### `files`

{: .var-table }
|**Type**|list of strings|
|**Valid values**|file names|
|**Default value**||

The `files` variable is a list of file names from which to load plans.
File paths are relative to the current working directory of `blip`.
See [Plans / Loading](../plans/loading) for details.

### `monitor`

{: .var-table }
|**Type**|dictionary|
|**Valid values**|[Monitor](#monitors)|
|**Default value**||

The `monitor` variable configures the MySQL instance from which the [`table`](#table-1) is loaded.

### `table`

{: .var-table }
|**Type**|string|
|**Valid values**|valid MySQL table name|
|**Default value**||

The `table` variable is the MySQL table name from which plans are loaded.
See [Plans / Table](../plans/table).

{: .config-section-title}
## sinks

The `sinks` section configures [built-in metric sinks](../sinks/) and [custom metrics sinks](../develop/sinks).
This section is a map of maps:

```yaml
sinks:
  sink-foo:
    # Options for sink-foo:
    key1: value1
  sink-bar:
    # Options for sink-bar:
    key1: value1
```

Keys are sink names.
Values for each are a key-value map (of strings) as options for the named sink.

The key-value options for each sink are sink-specific and passed directly to the sink.
The sink validates the options.
For built-in sinks, see [Sinks](../sinks/) for each one's options.
For custom sinks, the options are whatever you program the custom sink to accept.

{: .note }
Blip does not distinguish between built-in and custom sinks.
The built-in sinks are merely sink plugins automatically registered on startup;
they implement the same interface as custom sinks.

If no sinks are specified, Blip use the sink defined by package variable `sink.Default`, which is [log](../sinks/log).

{: .config-section-title}
##  tags

The `tags` section sets user-defined key-value pairs (as strings) that are passed to each sink.
For example (using [interpolation](interpolation)):

```yaml
tags:
  env: ${ENVIRONMENT:-dev}
  dc: ${DATACENTER:-local}
  hostname: %{monitor.hostname}
```

Blip calls these "tags", but each sink might have a different term for the same concept.
For example, with SignalFx these are called "dimensions".
But the concept is the same: metadata (usually string key-value pairs) attached to metrics that describe or annotate the metrics for grouping, aggregation, or filtering when display in graphs/charts.

The [default sinks](../sinks) automatically send all tags with metrics.
For example, the [`signalfx` sink](../sinks/signalfx) sends all tags as SignalFx dimensions.

{: .config-section-title}
## tls

The `tls` section configures TLS certificates.

```yaml
tls:
  ca: ""
  cert: ""
  key: ""
  disable: false
  skip-verify: false
```

You can specify only `tls.ca`, or `tls.cert` and `tls.key`, or all three; any other combination is invalid.

{: .note}
By default, Blip does not use TLS for MySQL connections _except_ when using AWS; see section [`aws`](#aws) or [AWS](../cloud/aws).

### `ca`

{: .var-table }
|**Type**|string|
|**Valid values**|file name|
|**Default value**||

The `ca` variables sets the certificate authority file.

### `cert`

{: .var-table }
|**Type**|string|
|**Valid values**|file name|
|**Default value**||

The `cert` variables sets the public certificate file.

### `key`

{: .var-table }
|**Type**|string|
|**Valid values**|file name|
|**Default value**||

The `key` variables sets the private key file.

### `disable`

{: .var-table }
|**Type**|bool|
|**Valid values**|`true` or `false`|
|**Default value**|`false`|

The `disable` variable disables TLS even if configured.
`ssl-mode=DISABLED` in a [`mycnf`](#mycnf) file also disables TLS.

### `skip-verify`

{: .var-table }
|**Type**|bool|
|**Valid values**|`true` or `false`|
|**Default value**|`false`|

Do not verify the server address (MySQL hostname).

# Monitors

The `monitors` section is a list of MySQL instances to monitor.
Each instance is a YAML dictionary containing any of the [monitor default sections](#monitor-defaults) with one exception: `mysql` variables are top-level in a monitor.
The example below shows two different MySQL instances to monitor.

```yaml
monitors:

  - hostname: db1.local
    username: metrics
    password-file: "/secret/db-password"
    heartbeat:
      freq: 1s

  - mycnf: "/secret/my.cnf"
    exporter:
      mode: legacy
```

The first MySQL instance is configured in lines 3-7.
(Note the single, leading hyphen on line 3 that denotes an item in a YAML list.)
The first three variables&mdash;`hostname`, `username`, and `password-file`&mdash;are [`mysql`](#mysql) variables but in a monitor they are top-level.
But all other sections, like [`heartbeat`](#heartbeat) and its variable `freq`, are exactly the same in a monitor.

The second MySQL instance is configured in lines 9-11.
Variable `mycnf` belongs to section `mysql`, but again: in a monitor, [`mysql`](#mysql) variables are top-level.
Section [`exporter`](#exporter) is exactly the same in a monitor.

<b>Refer to [Monitor Defaults](#monitor-defaults) for configuring MySQL instances, and remember: [`mysql`](#mysql) variables are top-level in a monitor (omit `mysql:` and include the variables directly).<b>

Monitors have three variables that only appear in monitors: `id`, `meta`, and `plan`.

### `id`

{: .var-table }
|**Type**|string|
|**Valid values**|any string|
|**Default value**|(automatic)|

The `id` variable uniquely identifies the MySQL instance in Blip.

Every monitor has a unique ID that, by default, Blip sets automatically.
You can set monitor IDs manually, but it's better to let Blip set them automatically to avoid duplicates (which causes a fatal error).

Blip uses monitor IDs to track and report each MySQL instance in its own output and API.

Blip does _not_ use monitor IDs to identify MySQL instances for reporting metrics, but you can use them if you want.
For example:
```yaml
monitors:
  - id: db1
    hostname: db1.local
    tags:
      monitorId: %{monitor.id}
```
Since tags are passed to sinks (which report metrics), all sinks will receive the monitor ID.
(Sinks receive the monitor ID at the code-level too, so technically this example is not necessary.)

Monitor IDs are not guaranteed to be stable&mdash;they might change between Blip versions.
Therefore, do not rely on them outside of Blip for truly stable, unique MySQL instance identification.

### `meta`

{: .var-table }
|**Type**|key-value map (string: string)|
|**Valid values**|any strings|
|**Default value**||

The `meta` variable is a map of key-value strings for user-defined monitor metadata.

No part of Blip uses or requires monitor metadata.
Unlike [`tags`](#tags) and [_metric_ metadata](../metrics/reporting#meta), Blip does not copy or send monitor metadata.
This makes monitor metadata useful for advanced or automated configurations because it allows you to add custom configuration and reference it with [interpolation](interpolation).

Monitor metadata is optional.
When useful, the Blip documentation will shown to use it.

### `plan`

{: .var-table }
|**Type**|string|
|**Valid values**|any string|
|**Default value**||

The `plan` variable selects the [shared plan](../plans/loading#shared) for the monitor to use if [`change`](#change) is not configured.
The default (no value) selects a plan according to [plan precedence](../plans/loading#precedence).
