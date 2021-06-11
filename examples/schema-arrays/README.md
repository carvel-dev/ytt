This example was created to show how to use [arrays](https://carvel.dev/ytt/docs/latest/lang-ref-ytt-schema/#inferring-defaults-for-arrays) with ytt Schema.

In this example, a schema file, a data values file, and a template file are provided. 

The default value for an array is an empty array. Run this command to see the empty array result:
```bash
ytt -f config.yml -f schema.yml --enable-experiment-schema
```

To override the default values from the schema file, provide data values via a [data values file](https://carvel.dev/ytt/docs/latest/ytt-data-values/#declaring-and-using-data-values). The values must be of the same type that is declared in `schema.yml`. 

Run this command to see that when an array item is provided in a data values file, the defaults for the item are filled in:
```bash
ytt -f config.yml -f schema.yml -f values.yml --enable-experiment-schema
```