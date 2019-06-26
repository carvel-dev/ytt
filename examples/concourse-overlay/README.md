This example shows how to remove portion of Concourse pipeline with a ytt overlay. Submitted as a [question](https://kubernetes.slack.com/archives/CH8KCCKA5/p1561562790249000) to slack.

Explanation of `remove-test.yml`:

- `#@overlay/match by=overlay.all` indicates to find all other documents (which may be not desirable, so you may have to add some more unique info to your pipeline to identify if you have multiple pipelines)

- `#@overlay/match by="name"` tells ytt to find job that matches `name: unit`

- `#@overlay/match by=overlay.subset({"task":"test"})` tells ytt to find array item that has `{"task":"test"}` in it
cant use `#@overlay/match by="task"` for that last one because no all array items within `plan` have task key

- `-` is an empty array item since we don't care what's inside
