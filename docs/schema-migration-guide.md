# Schema Migration Guide

## Arrays
When converting a data values file to a schema, be aware empty arrays are not supported.
A schema array must contain one element which describes the type of the array items contained within.
Multiple array items are also not suitable; these will not be considered default values for the array.

For example, the CF4K8S team will need to update line 5 of their [00-values.yml file](https://github.com/cloudfoundry/cf-for-k8s/blob/875c2bed288d8fafd3a351da03b0b630b27385c3/config/logging/_ytt_lib/cf-k8s-logging/00-values.yml#L5) \
For further reading, see the spec on [Inferring Defaults for Arrays](https://hackmd.io/pODV3wzbT56MbQTxbQOOKQ?view#Inferring-Defaults-for-Arrays)
