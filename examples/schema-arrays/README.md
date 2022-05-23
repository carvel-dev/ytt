This example was created to show how to use [arrays](https://carvel.dev/ytt/docs/latest/lang-ref-ytt-schema/#inferring-defaults-for-arrays) with ytt Schema.

In this example, a schema file, a data values file, and a template file are provided. 

The default value for an array is an empty array. Run this command to see the empty array result:
```bash
ytt -f config.yml -f schema.yml
```

To override the default values from the schema file, provide data values via a [data values file](http://carvel.dev/ytt/docs/develop/ytt-data-values/#configuring-data-values-via-command-line-flags). The values must be of the same type that is declared in `schema.yml`.

Run this command to see that when an array item is provided in a data values file, the defaults for the item are filled in:
```bash
ytt -f config.yml -f schema.yml -f values.yml
```