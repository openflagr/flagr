# Swagger / go-swagger customization

`make gen` runs `swagger generate server` (see root `Makefile`).

## JSON library (not a go-swagger flag)

go-swagger **v0.34.1** does not choose sonic/jsoniter/easyjson at generate time. Relevant knobs:

| Mechanism | What it does |
|-----------|----------------|
| `--struct-tags=json` (default) | Struct field tags only; models still use `encoding/json` + `swag/jsonutils` for `MarshalBinary` |
| `--existing-models` | Point at external model package (still std JSON unless you hand-write codecs) |
| `--template-dir` + `--allow-template-override` | Override templates (Flagr uses this for `configure_flagr.go`) |

Flagr HTTP JSON is wired in **`swagger/templates/server/configureapi.gotmpl`** → `swagger_gen/restapi/configure_flagr.go`, calling **`pkg/jsoncodec`** and runtime env **`FLAGR_EVAL_JSON_CODEC`** (`std` | `sonic`).

## Preserving `configure_flagr.go`

Historically the Makefile copied `/tmp/configure_flagr.go` after generate. With the template override above, regeneration should emit jsoncodec wiring directly. After `make swagger`, diff `configure_flagr.go` and drop the `/tmp` copy steps if the template stays in sync.

## Regenerate

```bash
make swagger   # or make gen
```