This example was created to show how to deployment rolling update strategy more dynamically.

[slack thread](https://kubernetes.slack.com/archives/C0NH30761/p1561408219254900)

> can someone please help out with this. I'm trying to set maxUnavailable value dynamically based on the number of requested replicas. For example, if the application is deployed with something like:

```
replicas: 5
```

> -> then i'd want that value for

```spec:
spec:
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
```

> maxUnavailable:  {someFunction that checks if replicas > 3 then use value of replicas - some factor}
> and can it be done inline or need to use helper function for this? Thank you
