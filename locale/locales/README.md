# API translations (gettext)

API error and success messages are translated using **gettext**. This directory is embedded at build time; the app loads only `.mo` files at runtime.

| Format | Role |
|--------|------|
| **.po** | Source: edit these to change or add languages (e.g. `ar.po`). |
| **.mo** | Compiled from `.po` with `msgfmt`; used at runtime for faster loading. Commit `.mo` so the app doesn’t need gettext at build time. |

**Build a .mo from a .po:**
```bash
msgfmt -o en.mo en.po
```

**Message format:** Each entry has a `msgid` (the key used in code, e.g. `notAuthenticated`) and a `msgstr` (the text shown to the user). Keys match the arguments to `locale.TR(r, "key")` in the API.

**Adding a language:** Duplicate `en.po` as e.g. `ar.po`, translate the `msgstr` values (leave `msgid` as-is), then run `msgfmt -o ar.mo ar.po` and add both `ar.po` and `ar.mo` here.
