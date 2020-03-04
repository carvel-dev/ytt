## Language

Templating language used in `ytt` is a slightly modified version of [Starlark](https://github.com/google/starlark-go/blob/master/doc/spec.md). Following modifications were made:

- requires `end` keyword for block closing
  - hence no longer whitespace sensitive (except new line breaks)
- does not allow use of `pass` keyword

## Reference

### Types

- NoneType: `None` (equivalent to null in other languages)
- Bool: `True` or `False`
- Integer: `1`
- Float: `1.1`
- String: `"string"` [Reference](lang-ref-string.md)
- Dictionary: `{"a": 1, "b": "b"}` [Reference](lang-ref-dict.md)
- List: `[1, 2, {"a":3}]` [Reference](lang-ref-list.md)
- Tuple: `(1, 2, "a")`
- Struct: `struct.make(field1=123, field2="val2")` [Reference](lang-ref-ytt.md#struct)
- TODO Set?
- YAMLFragment: [Reference](lang-ref-yaml-fragment.md)
- Annotation: `@name arg1,arg2,keyword_arg3=123` [Reference](lang-ref-annotation.md)

### Control flow

- [If conditional](lang-ref-if.md)
- [For loop](lang-ref-for.md)
- [Function](lang-ref-def.md)

### Libraries

- [@ytt Library](lang-ref-ytt.md)
  - [@ytt:overlay module](lang-ref-ytt-overlay.md)
  - [@ytt:library module](lang-ref-ytt-library.md)
