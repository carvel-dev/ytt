## Outputs

ytt supports two different output destinations:

- stdout, which is default
- output directory, controlled via `--output-directory`

When destination is stdout, all YAML documents are combined into one document set. Non-yaml files are not printed anywhere.

When destination is an output directory, ytt will _empty out_ directory beforehand and write out result files preserving file names.

If you want to control which files are included in the output use `--file-mark 'something.yml:exclusive-for-output=true'` flag to mark one or more files.
