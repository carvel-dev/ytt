### For loop

```yaml
array:
#@ for i in range(0,3):
- #@ i
- #@ i+1
#@ end
```

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
