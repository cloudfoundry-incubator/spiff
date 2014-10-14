```
                                        ___ _ __ (_)/ _|/ _|
                                       / __| '_ \| | |_| |_
                                       \__ \ |_) | |  _|  _|
                                       |___/ .__/|_|_| |_|
                                           |_|

```

A declarative YAML templating system tuned for BOSH deployment manifests.

# installation

Download a release from GitHub, or install with HomeBrew on OS X:

```sh
brew tap xoebus/homebrew-cloudfoundry
brew install spiff
spiff
```

# running tests

```
go get github.com/kr/godep
godep go test -v ./...
```

# usage

### `spiff merge template.yml [template2.ymll ...]`

Merge a bunch of template files into one manifest, printing it out.

See 'dynaml templating language' for details of the template file, or
`example.yml` for a complicated example.

Example:

```
spiff merge cf-release/templates/cf-deployment.yml my-cloud-stub.yml
```

### `spiff diff manifest.yml other-manifest.yml`

Show structural differences between two deployment manifests.

Unlike 'bosh diff', this command has semantic knowledge of a deployment
manifest, and is not just text-based. It also doesn't modify either file.

It's tailed for checking differences between one deployment and the next.

Typical flow:

```sh
$ spiff merge template.yml [templates...] > deployment.yml
$ bosh download manifest [deployment] current.yml
$ spiff diff deployment.yml current.yml
$ bosh deployment deployment.yml
$ bosh deploy
```


# dynaml templating language

Spiff uses a declarative, logic-free templating language called 'dynaml'
(dynamic yaml).

Every dynaml node is guaranteed to resolve to a YAML node. It is *not*
string interpolation. This keeps developers from having to think about how
a value will render in the resulting template.

A dynaml node appears in the .yml file as an expression surrounded by two
parentheses. They can be used as the value of a map or an entry in a list.

The following is a complete list of dynaml expressions:


## `(( foo ))`

Look for the nearest 'foo' key (i.e. lexical scoping) in the current
template and bring it in.

e.g.:

```
fizz:
  buzz:
    foo: 1
    bar: (( foo ))
  bar: (( foo ))
foo: 3
bar: (( foo ))
```

This example will resolve to:

```
fizz:
  buzz:
    foo: 1
    bar: 1
  bar: 3
foo: 3
bar: 3
```

## `(( foo.bar.[1].baz ))`

Look for the nearest 'foo' key, and from there follow through to .bar.baz.

A path is a sequence of steps separated by dots. A step is either a word for
maps, or digits surrounded by brackets for list indexing.

If the path cannot be resolved, this evaluates to nil. A reference node at the
top level cannot evaluate to nil; the template will be considered not fully
resolved. If a reference is expected to sometimes not be provided, it should be
used in combination with '||' (see below) to guarantee resolution.

Note that references are always within the template, and order does not matter.
You can refer to another dynamic node and presume it's resolved, and the
reference node will just eventually resolve once the dependent node resolves.

e.g.:

```
properties:
  foo: (( something.from.the.stub ))
  something: (( merge ))
```

This will resolve as long as 'something' is resolveable, and as long as it
brings in something like this:

```
from:
  the:
    stub: foo
```

## `(( "foo" ))`

String literal. The only escape character handled currently is '"'.

## `(( "foo" bar ))`

Concatenation (where bar is another dynaml expr).

e.g.

```
domain: example.com
uri: (( "https://" domain ))
```

In this example `uri` will resolve to the value `"https://example.com"`.

## `(( merge ))`

Bring the current path in from the stub files that are being merged in.

e.g.:

```
foo:
  bar:
    baz: (( merge ))
```

Will try to bring in `foo.bar.baz` from the first stub, or the second, etc.,
returning the value from the first stub that provides it.

If the corresponding value is not defined, it will return nil. This then has the
same semantics as reference expressions; a nil merge is an unresolved template.
See `||`.

## `(( a || b ))`

Uses a, or b if a cannot be resolved.

e.g.:

```
foo:
  bar:
    - name: some
    - name: complicated
    - name: structure

mything:
  complicated_structure: (( merge || foo.bar ))
```

This will try to merge in `mything.complicated_structure`, or, if it cannot be
merged in, use the default specified in `foo.bar`.
