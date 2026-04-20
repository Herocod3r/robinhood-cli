# Output schema

Every command emits a stable envelope:

```json
{
  "schema": "robinhood-cli/v1",
  "command": "<name>",
  "generated_at": "<RFC3339 UTC>",
  "data": "<command-specific>",
  "meta": { "count": 0, "profile": "<profile>" },
  "error": null
}
```

On error, `data` is `null` and `error` is non-null with shape:

```json
{
  "code": "<code>",
  "message": "<human message>",
  "hint": "<actionable hint>",
  "retryable": false
}
```

See the CLI repo doc `docs/JSON_SCHEMA.md` for every command's `data` shape.

To get the schema for a specific command at runtime:

```bash
rh schema <cmd> --json
```

This is authoritative — it is generated from the same structs the CLI uses.
