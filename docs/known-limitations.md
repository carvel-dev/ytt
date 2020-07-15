## Known Limitations

- All YAML comments are expected to be qualified (i.e. must start with `#@ ` or `#! `) by default. Current reason for this is ytt is trying to be cautious and disallow non-qualified comments (`# `) since users may unintentionally write templating directives but forget to use `@` after `#`. This default behaviour may change in future.

- YAML merge operator (`<<:`) is not currently supported.

- YAML anchors and templating directive for the same YAML node are not supported.

    ```yaml
    first: &content #@ 123
    second: *content
    ```
    `second` key-value pair will _not_ contain 123 since YAML anchors are resolved before ytt evaluates templating directives.
