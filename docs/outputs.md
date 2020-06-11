## Outputs

ytt supports three different output destinations:

- stdout, which is default
  - All YAML documents are combined into one document set. Non-YAML files are not printed anywhere.
- output files, controlled via `--output-files` flag (v0.28.0+)
  - Output files will be added to given directory, preserving file names.
    - Example: `ytt -f config.yml --output-files tmp/`.
- output directory, controlled via `--dangerous-emptied-output-directory` flag
  - Given directory will be _emptied out_ beforehand and output files will be added preserving file names.
    - Example: `ytt -f config.yml --dangerous-emptied-output-files tmp/ytt/`.

If you want to control which files are included in the output use `--file-mark 'something.yml:exclusive-for-output=true'` flag to mark one or more files.
