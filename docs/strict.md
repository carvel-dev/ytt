## Strict YAML subset

ytt includes strict YAML subset mode that tries to remove any kind of ambiguity in user's intent when parsing YAML.

Unlike full YAML, strict subset:

- only supports specifying nulls as `` or `null`
- only supports specifying bools as `false` or `true`
- only support basic int and float declarations
  - prefix, suffix, octal notation, etc are not supported
- requires strings with whitespace to be explicitly quoted
- requires strings with colon to be explicitly quoted
- requires strings with triple-dash (document start) to be explicitly quoted

### Example

Non-strict:

```bash
$ echo 'key: yes'|ytt -f-
key: true
```

Strict:

```bash
$ echo 'key: yes'|ytt -f- -s
Error: Unmarshaling YAML template 'stdin.yml': yaml: Strict parsing: Found 'yes' ambigious (could be !!str or !!bool)
```

To fix error, explicitly make it a string:

```bash
$ echo 'key: "yes"'|ytt -f- -s
key: "yes"
```

or via YAML tag `!!str`:

```bash
$ echo 'key: !!str yes'|ytt -f- -s
key: "yes"
```
