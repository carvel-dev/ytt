This example shows how to use `@ytt:library` module to programmatically template custom library `app` (found `_ytt_lib/app`). See comments in `config.yml` for more details:

```
# turn _ytt_lib1 -> _ytt_lib
$ mv ./examples/playground/example-ytt-library-module/_ytt_lib{1,}

$ ytt -f ./examples/playground/example-ytt-library-module
```
