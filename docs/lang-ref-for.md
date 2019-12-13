### For loop

Refer to [Starlark for loop specification](https://github.com/google/starlark-go/blob/master/doc/spec.md#for-loops) for details.

- iterating with values

```yaml
array:
#@ for i in range(0,3):
- #@ i
- #@ i+1
#@ end
```

- iterating with index

```yaml
array:
#@ arr = [1,5,{"key":"val"}]
#@ for i in range(len(arr)):
- val: #@ arr[i]
  index: #@ i
#@ end 
```

- use of `continue/break`

```yaml
array:
#@ for i in range(0,3):
#@   if i == 1:
#@     continue
#@   end
- #@ i
- #@ i+1
#@ end
```
