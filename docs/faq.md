# FAQ

## How do I template some text?

[Here is an example](https://get-ytt.io/#example:example-text-template) describing some of the ways text templating can be done.

## Data Values

[Data values doc](ytt-data-values.md)

### How should I check for the existence of a data value?

ytt library's `assert` package is useful for such situations. It can be used like so:

```yaml
password: #@ data.values.env.mysql_password if data.values.env.mysql_password else assert.fail("missing env.mysql_password")
```

or even more compactly,

```yaml
password: #@ data.values.env.mysql_password or assert.fail("missing env.mysql_password")
```

Note that empty strings are falsy in Starlark.

### How can I provide a default for a data value when it may not be defined?
When a value may be null, you can use `or` to specify a default
```yaml
#@ data.values.foo or "bar"
```
Slack link: [2020-05-12](https://kubernetes.slack.com/archives/CH8KCCKA5/p1589307427287400)

### Can I check if a data.values struct (ie. dictionary) has keys that are set?
Yes, you can check the existence of a key by using `hasattr(foo, "bar")`.\
Slack link: [2020-06-29](https://kubernetes.slack.com/archives/CH8KCCKA5/p1593456026390900), [2020-06-26](https://kubernetes.slack.com/archives/CH8KCCKA5/p1593166736308800)

### What's the "best" ways to put secrets into my yamls via ytt?\
Since the main values.yml template effectively acts as a schema for values (we are actually adding explicit schema definition for values instead of relying on the first data values), you cannot inject secrets via a values2.yml file unless the main values.yml template also contains definitions of the values you wish to set. One easy way to do this would be to include
    secrets:
    actual_secret: ""
in your values.yml and then override that with your values2.yml like:\
    secrets:
    actual_secret: "foo"
Additional Resources: [injecting secrets doc](https://github.com/k14s/ytt/blob/develop/docs/injecting-secrets.md)Slack link: [2020-05-14](https://kubernetes.slack.com/archives/CH8KCCKA5/p1589503314304100), [2020-05-23](https://kubernetes.slack.com/archives/CH8KCCKA5/p1590224367405600)

### Is there a way to have a function return a multi-line string that contains text and values from data.values? Is there a pattern for injecting values into multi-line strings?\
Yes, depending on your situation you could [text templating](https://github.com/k14s/ytt/blob/develop/docs/ytt-text-templating.md), or if the string is a json file, you could follow [this example](https://github.com/k14s/ytt/tree/develop/examples/k8s-config-map-files) and describe it as yaml for structure then serialize it to json.\
Slack link: [2020-07-11](https://kubernetes.slack.com/archives/CH8KCCKA5/p1594483959052300), [2020-05-15](https://kubernetes.slack.com/archives/CH8KCCKA5/p1592274814197200)

### Can I generate random strings in ytt?

No. A design goal of ytt is to generate output deterministically. This makes generating random strings out of scope for the tool.

Additional resources: 

-   Documentation on handling secrets, in general:\
    <https://github.com/k14s/ytt/blob/develop/docs/injecting-secrets.md>

-   Kubernetes controller that generates secrets:\
    <https://github.com/k14s/secretgen-controller>

Slack link: , [2020-07-08](https://kubernetes.slack.com/archives/CH8KCCKA5/p1594224108011100?thread_ts=1594224060.011000&cid=CH8KCCKA5)

### How can I load json for use as a data value? Relatedly, why can't I add a new key to my values with the `--data-value` command line argument?

There are two approaches to loading json. An important note is that json is valid yaml, ie. yaml syntax is a superset of json syntax. For the first approach, ytt can naturally parse json by passing it through `--data-value-yaml`. Alternatively, json can be loaded by passing the file as a `--data-value-file` but in addition this requires a values file with the keys. Values passed from the command line must be overrides, not new values.

Additional resources: [json is valid yaml](https://gist.github.com/pivotaljohn/debe4596df5b4158c7c09f6f1841dd47), [loading by file](https://gist.github.com/pivotaljohn/d3468c3239f79fea7e232751757e779a)

Slack link:

### What templating language does ytt use? Why is it using a fork instead of the main branch?

Starlark. Makes ytt non-whitespace-sensitive. 

Additional resources: [Language reference](https://github.com/k14s/ytt/blob/develop/docs/lang.md#language), [Starlark spec](https://github.com/google/starlark-go/blob/master/doc/spec.md)

Slack link:

### Comparison with helm, concern that porting to ytt is a lot of work, and that helm is already the standard

Helm "Release" can be omitted if one is deploying a single instance of their app. Many filters disappear in ytt, specifically those surrounding whitespace as ytt always generates valid yaml. Variables that mirror kubernetes resource properties can be inlined; ytt can always allow overlaying. While helm has some traction, it is not the de facto standard.

Additional resources:

Slack link:

### How can I remove a document subset

Overlays! The `remove` overlay in conjunction with the `subset` overlay matcher

Additional resources: [Overlay remove docs](https://github.com/k14s/ytt/blob/develop/docs/lang-ref-ytt-overlay.md#overlayremove), [overlay subset docs](https://github.com/k14s/ytt/blob/develop/docs/lang-ref-ytt-overlay.md#overlaysubset)

Slack link:

### How do I add items to an existing array?  

Overlays! The default behaviour for arrays is to overwrite as there isn't any default matching criteria (unlike maps which can use their keys). To append, use the `#@overlay/append` directive. Note the directive has to be applied to each individual item that we want to merge in. If you wish to add something more complex, like an array into another array, it may be helpful to use the `expects=` match annotation.

Additional resources: [Overlay docs](https://github.com/k14s/ytt/blob/develop/docs/lang-ref-ytt-overlay.md#overlayappend), [example gist on playground](https://get-ytt.io/#gist:https://gist.github.com/pivotaljohn/8c7f48e183158ce12107f576eeab937c), [replace-list gist](https://get-ytt.io/#gist:https://gist.github.com/pivotaljohn/2b3a9b3367137079195971e1409d539e), [edit-list gist](https://get-ytt.io/#gist:https://gist.github.com/pivotaljohn/217e8232dc080bb764bfd064ffa9c115)

Slack link: [2020-04-18](https://kubernetes.slack.com/archives/CH8KCCKA5/p1587248371373800), [2020-04-18](https://kubernetes.slack.com/archives/CH8KCCKA5/p1587248746378700), [2020-04-27](https://kubernetes.slack.com/archives/CH8KCCKA5/p1588019152046200), [2020-05-08](https://kubernetes.slack.com/archives/CH8KCCKA5/p1588962390214100), 

### How do I rename a key using overlays?

Here is a gist showing how to rename a key without editing the value.

[Rename-key gist](https://get-ytt.io/#gist:https://gist.github.com/gcheadle-vmware/3c41645a80201caaeefa878e84fff958)

[2020-05-25](https://kubernetes.slack.com/archives/CH8KCCKA5/p1590434935446000)

### How do I add/replace values in a dictionary?

To add or replace values in a dictionary you can use [template.replace](https://get-ytt.io/#example:example-replace). You can also use overlays to edit a dictionary, some examples can be found on [this gist playground](https://get-ytt.io/#gist:https://gist.github.com/gcheadle-vmware/af8aeb3120386e58922c816d76f47ab6).

Slack link: [2020-05-02](https://kubernetes.slack.com/archives/CH8KCCKA5/p1588427231127500), [2020-06-29](https://kubernetes.slack.com/archives/CH8KCCKA5/p1593468228397200)

### How can I match a field.name that starts with a string?

overlay/match by=lambda a,_: a["field"]["name"].startswith("string")

Slack link:

### How can I apply an overlay to modify a struct based on the presence of a key?

If you wanted to modify a dictionary from list of dictionaries only if it the `foo:` key is present, then you can add `#@overlay/match by=lambda idx,old,new: "foo" in old, expects="1+"` at the start of list's definition.

Slack link: [2020-04-18](https://kubernetes.slack.com/archives/CH8KCCKA5/p1587249303386100)

### How would I modify one line in a multi-line string?

You can use overlays! [modify-string gist](https://get-ytt.io/#gist:https://gist.github.com/cppforlife/7633c2ed0560e5c8005e05c8448a74d2)

Slack link: [2020-07-08](https://kubernetes.slack.com/archives/CH8KCCKA5/p1594235515016900)

### Is there a way to match a regex pattern in the subset matcher?

Not directly. Instead a custom matcher can be written. See this [playground gist](https://get-ytt.io/#gist:https://gist.github.com/ewrenn8/3409e44252f93497a9b447900f3fb5b7)

Slack link:

### Why am I getting an exception when trying to append?

A common issue when trying to append is to incorrectly set the match missing_ok=True on the key which gets replaced by new key-values. Instead, the match mising_ok=True should be applied to each child (which can be easily done with the match-child-defaults annotation)

Additional resources: [gist from Dmitriy](https://gist.github.com/cppforlife/bf42f2d3d23dacf07affcd4150370cb9)

Slack link:

### Why can't I write standard yaml comments (#)? Why doesn't ytt support the yaml merge operator (<<:)? Why is my anchor reference null despite my anchor's successful template?

These are [know limitations](https://github.com/k14s/ytt/blob/develop/docs/known-limitations.md)

Slack link:

### Can I load multiple functions without having to name each one?

Yes, you can store the functions in a struct and just import the struct and call the functions with the struct.

Storing functions in struct:

`#@ load("@ytt:struct", "struct")

#@ mod = struct.make(func1=func1, func2=func2)

`

Loading and Calling functions in template:

`#@ load("helpers.lib.yml", "mod")

something: #@ mod.func1()

`

Slack link: [2020-04-23](https://kubernetes.slack.com/archives/CH8KCCKA5/p1587660183496700), [2020-05-04](https://kubernetes.slack.com/archives/CH8KCCKA5/p1588615212156300)

* * * * *

