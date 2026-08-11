# Changelog

All notable changes to this provider are documented here.

GoReleaser already produces a per-release list of commit subjects. This file exists for
what that list cannot convey: behaviour that changes under an unchanged configuration,
and the reason a fix matters. Entries follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.1.0] - 2026-08-11

First release since `v1.0.13`. It aligns the Alertmanager receiver schema with the Alertmanager
fork Mimir vendors, and fixes the defects found by a full review of the provider. A minor bump
rather than the usual patch, because two changes alter behaviour under an unchanged
configuration — see below.

### Upgrade notes

Review a plan before applying: the provider sends something different for a configuration you
have not edited.

- **SMTP now requires TLS by default.** `email_configs.require_tls` and `global.smtp_require_tls`
  defaulted to `false` and were always emitted, so every generated configuration disabled
  STARTTLS even though Alertmanager itself defaults it to on. They now default to `true`.
  A relay that does not offer STARTTLS needs `require_tls = false` set explicitly, otherwise
  alert email delivery stops at the first apply.
- **`ignore_groups` and `only_groups` no longer delete.** A rule group excluded from management
  but still declared in `content` is now left alone in Mimir instead of being deleted. Removing
  a group from the YAML still deletes it, as before.

A one-time non-empty plan is expected on first use for:

- rule group `interval`, `query_offset` and `evaluation_delay`, and `pushover_configs`
  `retry`/`expire`/`ttl`, whose values are now normalized in state the same way the
  configuration normalizes them;
- receiver credentials now marked sensitive, which changes how they render in plan output.

### Added

- Alertmanager receiver attributes available on every supported Mimir: `webhook_configs.timeout`,
  `pagerduty_configs.source`, `pushover_configs.device` and `ttl`, `global.smtp_tls_config`.
- Alertmanager receiver attributes requiring Mimir 3.1.0 or later, documented as such in each
  description: discord `content`/`username`/`avatar_url`, email `force_implicit_tls` and
  `threading`, pagerduty `timeout`, slack `app_token`/`app_url`/`message_text`/`timeout`,
  pushover `monospace`, telegram `message_thread_id`, global `smtp_force_implicit_tls`/
  `slack_app_token`/`slack_app_url`/`telegram_bot_token`, and `inhibit_rule.name`.

  Mimir rejects a configuration carrying a key it does not know, so an attribute left unset never
  reaches the generated YAML and older Mimir versions are unaffected.
- `mimir_rules` now round-trips `keep_firing_for` on a rule and `query_offset`,
  `evaluation_delay` and `align_evaluation_time_on_interval` on a group. They were absent from
  the Go model, and the resource re-marshals what it parses, so they were stripped from every
  rule group it wrote.
- The alerting rule group data source exposes `keep_firing_for`, `interval` and group labels;
  the recording one exposes group labels.
- `terraform import` works on `mimir_rules`: the importer rebuilds `content` from the namespace
  as the ruler holds it. It previously set only `org_id` and `namespace`, so the Read that
  Terraform runs straight afterwards always failed and the import could never complete.

### Changed

- Generated configurations use the current `time_intervals` key rather than the deprecated
  `mute_time_intervals`. Reads still accept either spelling, so importing a hand-written
  configuration keeps working, and existing users see no Terraform diff.
- Debug mode no longer dumps request and response bodies, and redacts the `Authorization`,
  `Proxy-Authorization` and `X-Amz-Security-Token` headers. The bodies carried receiver
  credentials in cleartext into the Terraform log.
- The proxy URL is logged only when `debug` is enabled.
- The acceptance test matrix covers Mimir 3.1.4 in addition to 2.17.10 and 3.0.6, version-gated
  tests skip rather than abort the test binary when `MIMIR_VERSION` is unset, and `make generate`
  is wired to a `go:generate` directive so the documentation drift check in CI actually runs.

### Fixed

- `mimir_rules` no longer deletes every group it manages when `only_groups` names a group that
  the configuration does not define. The unmatched name is now rejected at plan time; previously
  the managed set silently shrank to nothing and Update deleted the lot, reporting success.
- The provider no longer panics when reading an `sns_configs` receiver stored without a `sigv4`
  key, which a configuration written by mimirtool or by hand routinely is.
- The provider no longer panics when a `time_interval` `times` block sets only one of
  `start_time`/`end_time`, which passes validation.
- `wechat_configs.api_url` reaches Mimir; it was read from a key no schema declares, so it was
  silently dropped and re-planned on every run.
- `http_config.oauth2.tls_config` and `email_configs.tls_config` are sent; both were declared in
  the schema but never read, so the TLS settings were silently discarded.
- A rule group deleted outside Terraform is detected and recreated when `content` is used inline.
- Deleting a rule group that is already gone from Mimir succeeds instead of leaving the resource
  stuck in state.
- Durations render in state the way the configuration normalizes them, removing a permanent diff
  on `pushover_configs` `retry`/`expire`/`ttl` and on rule group `interval`, `query_offset` and
  `evaluation_delay`.
- Empty `sigv4` and slack `confirm` blocks are no longer written into generated configurations
  and state, removing a plan that never converged.
- Trailing slashes are trimmed from `ruler_uri`, `alertmanager_uri` and `distributor_uri`, not
  just `uri`, so request paths no longer contain a double slash.
- An unusable `ca` bundle fails immediately with a clear message instead of surfacing later as an
  unrelated TLS handshake error.
- Request construction and debug dump failures return an error instead of calling `log.Fatal`,
  which terminated the provider plugin process.
- The read-delay settings accept both duration syntaxes the provider has taken (`1d` and `1.5s`)
  and report an unparseable value instead of silently using no delay.
- The alertmanager data source uses its own `org_id` rather than the provider's when computing
  its ID and default `name`.
- The alerting rule group data source no longer fails to read a group whose rules use
  `keep_firing_for`.
- Time range attributes on the alertmanager data source are named `start_time`/`end_time`,
  matching what the provider actually writes. They were declared as `start_minute`/`end_minute`
  and could never be populated.
- Nested object attributes in the generated documentation have titles again
  (`terraform-plugin-docs` bumped to v0.25.0; v0.19.0 rendered them empty).

### Security

- Receiver credentials are marked sensitive so they no longer appear in plan output:
  `discord_configs`, `msteams_configs` and `msteamsv2_configs` `webhook_url`, `pushover_configs`
  `user_key` and `token`, and `slack_configs.api_url`.
- `org_id` is validated on the alertmanager resource and data source and on the distributor stats
  data source. It becomes the `X-Scope-OrgID` tenant and a storage path component, so an
  unchecked value was a cross-tenant primitive.
