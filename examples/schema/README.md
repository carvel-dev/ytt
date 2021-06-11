This example was created to show a simplified introduction to using ytt Schema.

In following examples, a schema file and a template file are provided. 

To see the default values from the schema file in the output, run this command:
```bash
ytt -f config.yml -f schema.yml --enable-experiment-schema
```

To override the default values from the schema file, provide data values via [`--data-value-yaml`](https://carvel.dev/ytt/docs/latest/ytt-data-values/#overriding-data-values-via-command-line-flags) flags, or by adding a [data values file](https://carvel.dev/ytt/docs/latest/ytt-data-values/#declaring-and-using-data-values). The values must be of the same type that is declared in `schema.yml`. 
```bash
ytt -f config.yml -f schema.yml --data-value-yaml string=str --enable-experiment-schema
```