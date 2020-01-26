## File marks

ytt allows to control certain metadata about files via `--file-mark` flag.

- `--file-mark` (optional) Format: `<path>:<mark>=<value>`. Can be specified multiple times. Example: `--file-mark generated.go.txt:exclusive-for-output=true`.

Available marks:

- `path`: Changes relative path. Example: `generated.go.txt:path=gen.go.txt`.
- `exclude`: Exclude file from any kind of processing. Values: `true`. Example `config.yml:exclude=true`.
- `type`: Change type of file. By default type is determined based on file extension. Values: `yaml-template`, `yaml-plain`, `text-template`, `text-plain`, `starlark`, `data`. Example `config.yml:type=data`.
- `for-output`: Mark file to be used as part of output. Values: `true`. Example `config.lib.yml:for-output=true`.
- `exclusive-for-output`: Mark file to be used _exclusively_ as part of output. If there is at least one file marked this way, only these files will be used in output. Values: `true`. Example `config.lib.yml:exclusive-for-output=true`.

Path can be following:

- exact path (use `--files-inspect` to see paths as seen by ytt)
- path with `*` to match files in a directory
- path with `**/*` to match files and directories recursively

### Example

```bash
ytt -f . \
  --file-mark 'alt-example**/*:type=data' \
  --file-mark 'example**/*:type=data' \
  --file-mark 'generated.go.txt:exclusive-for-output=true' \
  --output-directory ../../tmp/
```
