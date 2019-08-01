## Strict YAML subset

ytt includes strict YAML subset mode that tries to remove any kind of ambiguity in user's intent when parsing YAML.

Unlike full YAML, strict subset:

- only supports specifying nulls as `` or `null`
- only supports specifying bools as `false` or `true`
- only support basic int and float declarations
  - none of the suffix, octal notations, etc are supported
- all strings that include whitespace or colon to be explicitly quoted
