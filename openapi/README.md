# API documentation

API endpoints are documented in **OpenAPI 3.0** with **JSON Schema** used for request and response bodies.

- **Spec:** [openapi.yaml](./openapi.yaml)

## Format

- **Paths** are relative to the server base path `/api/v1`.
- **Request/response bodies** are defined under `components/schemas` using JSON Schema (OpenAPI Schema Object).
- Reusable types (e.g. `User`, `ErrorResponse`) are referenced with `$ref`.

## Adding or updating endpoints

1. Add or edit the path under `paths` in `openapi.yaml`.
2. Define or reuse a schema in `components/schemas` for request body and responses.
3. Add path parameters under `components/parameters` if shared.
4. Keep schema property names and types in sync with the Go structs in `api/` (e.g. `json` tags).

## Viewing the spec

- Use an OpenAPI viewer (e.g. [Swagger Editor](https://editor.swagger.io), or VS Code extension) to render the spec.
- Some tools can generate clients or validate requests/responses from this file.
