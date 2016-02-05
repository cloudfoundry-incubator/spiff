```
                                        ___ _ __ (_)/ _|/ _|
                                       / __| '_ \| | |_| |_
                                       \__ \ |_) | |  _|  _|
                                       |___/ .__/|_|_| |_|
                                           |_|

```

---

**NOTE**: *Active development on spiff is currently paused, including Pull Requests.  Very severe issues will be addressed, and we will still be actively responding to requests for help via Issues.*

---

spiff is a command line tool and declarative YAML templating system, specially designed for generating BOSH deployment manifests.

Contents:
- [Installation](#installation)
- [Usage](#usage)
- [dynaml Templating Language](#dynaml-templating-language)
	- [(( foo ))](#-foo-)
	- [(( foo.bar.[1].baz ))](#-foobar1baz-)
	- [(( "foo" ))](#-foo--1)
	- [(( foo bar ))](#-foo-bar-)
		- [(( "foo" bar ))](#-foo-bar--1)
		- [(( [1,2] bar ))](#-12-bar-)
	- [(( auto ))](#-auto-)
	- [(( merge ))](#-merge-)
		- [<<: (( merge ))](#--merge-)
			- [merging maps](#merging-maps)
			- [merging lists](#merging-lists)
		- [- <<: (( merge on key ))](#----merge-on-key-)
		- [<<: (( merge replace ))](#--merge-replace-)
			- [merging maps](#merging-maps-1)
			- [merging lists](#merging-lists-1)
		- [<<: (( foo )) ](#--foo-)
			- [merging maps](#merging-maps-2)
			- [merging lists](#merging-lists-2)
		- [<<: (( merge foo ))](#--merge-foo-)
			- [merging maps](#merging-maps-3)
			- [merging lists](#merging-lists-3)
	- [(( a || b ))](#-a--b-)
	- [(( 1 + 2 * foo ))](#-1--2--foo-)
	- [(( "10.10.10.10" - 11 ))](#-10101010---11-)
	- [(( join( ", ", list) ))](#-join---list-)
	- [(( exec( "command", arg1, arg2) ))](#-exec-command-arg1-arg2-)
	- [(( static_ips(0, 1, 3) ))](#-static_ips0-1-3-)
	- [Mappings](#mappings)
		- [(( map[list|elem|->dynaml-expr] ))](#-maplistelem-dynaml-expr-)
		- [(( map[map|key,value|->dynaml-expr] ))](#-maplistidxelem-dynaml-expr-)
	- [Operation Priorities](#operation-priorities)
- [Structural Auto-Merge](#structural-auto-merge)
- [Useful to Know](#useful-to-know)
- [Error Reporting](#error-reporting)


# Installation

Official release executable binaries can be downloaded via [Github releases](https://github.com/cloudfoundry-incubator/spiff/releases) for Darwin and Linux machines (and virtual machines).

Some of spiff's dependencies have changed since the last official release, and spiff will not be updated to keep up with these dependencies.  Working dependencies are vendored in the `Godeps` directory (more information on the `godep` tool is available [here](https://github.com/tools/godep)).  As such, trying to `go get` spiff will likely fail; the only supported way to use spiff is to use an official binary release.

# Usage

### `spiff merge template.yml [stubN.yml ... stub3.yml stub2.yml stub1.yml]`

Merge a bunch of stub files into one template manifest, printing it out.

By default in Spiff, “merge” means that a stub feeds its values in to the
rigid structure of a template. This basic behavior can be tweaked with all
sorts of [dynaml expressions](#dynaml-templating-language) that are detailed
in the following sections.

The following major rules are worth knowing to understand how spiff performs
its merge process.

1. Spiff iterates on stubs from _right to left_ as follows:

   a. First, `stub1.yml` feeds its values into the structure of `stub2.yml`,
      as if `stub2.yml` was a template.

   b. Then the result is treated as a stub and feeds its resolved values into
      the structure of `stub3.yml` as if `stub3.yml` was a template. This step
      is repeated for all intermediate stubs in the right-to-left iteration.

   c. When all stubs are merged as above, the resulting stub feeds its values
      into the structure of `template.yml`.

2. A deep merge of maps is made, with the default intent that the structure of
   the template is meant to be kept. By default, a stub never creates any new
   nodes into a template.

3. Lists are not merged, excepted certain lists of maps.
   See “[Structural Auto-Merge](#structural-auto-merge)” for more details.

4. At each steps of the process, all dynaml expressions must be resolvable.

As a result of this process, values defined by rightmost stubs are meant to
override similar values defined by any stubs on their left. But for this to
happen, those values must be "transmitted" by all intermediate stubs before
they arrive in the final leftmost template.

Example:

```
spiff merge cf-release/templates/cf-deployment.yml my-cloud-stub.yml
```

More complicated examples can be found in the [examples](./examples) subdir.


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


# dynaml Templating Language

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

```yaml
fizz:
  buzz:
    foo: 1
    bar: (( foo ))
  bar: (( foo ))
foo: 3
bar: (( foo ))
```

This example will resolve to:

```yaml
fizz:
  buzz:
    foo: 1
    bar: 1
  bar: 3
foo: 3
bar: 3
```

The following will not resolve because the key name is the same as the value to be merged in:
```yaml
foo: 1

hi:
  foo: (( foo ))
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

```yaml
properties:
  foo: (( something.from.the.stub ))
  something: (( merge ))
```

This will resolve as long as 'something' is resolveable, and as long as it
brings in something like this:

```yaml
from:
  the:
    stub: foo
```

If the path starts with a dot (`.`) the path is always evaluated from the root
of the document.

List entries consisting of a map with `name` field can directly be addressed 
by their name value.

e.g.:

The age of alice in

```yaml
list:
 - name: alice
   age: 25
```

can be referenced by using the path `list.alice.age`, instead of `list[0].age`.


## `(( "foo" ))`

String literal. The only escape character handled currently is '"'.

## `(( foo bar ))`

Concatenation expression used to concatenate a sequence of dynaml expressions.

### `(( "foo" bar ))`

Concatenation (where bar is another dynaml expr). Any sequences of simple values (string, integer and boolean) can be concatenated, given by any dynaml expression.

e.g.

```yaml
domain: example.com
uri: (( "https://" domain ))
```

In this example `uri` will resolve to the value `"https://example.com"`.

### `(( [1,2] bar ))`

Concatenation of lists as expression (where bar is another dynaml expr). Any sequences of lists can be concatenated, given by any dynaml expression.

e.g.

```yaml
other_ips: [ 10.0.0.2, 10.0.0.3 ]
static_ips: (( ["10.0.1.2","10.0.1.3"] other_ips ))
```

In this example `static_ips` will resolve to the value `[ 10.0.1.2, 10.0.1.3, 10.0.0.2, 10.0.0.3 ] `.

If the second expression evaluates to a value other than a list (integer, boolean, string or map), the value is appended to the first list.

e.g.

```yaml
foo: 3
bar: (( [1] 2 foo "alice" ))
```
yields the list `[ 1, 2, 3, "alice" ]` for `bar`.

## `(( auto ))`

Context-sensitive automatic value calculation.

In a resource pool's 'size' attribute, this means calculate based on the total
instances of all jobs that declare themselves to be in the current resource
pool.

e.g.:

```yaml
resource_pools:
  - name: mypool
    size: (( auto ))

jobs:
  - name: myjob
    resource_pool: mypool
    instances: 2
  - name: myotherjob
    resource_pool: mypool
    instances: 3
  - name: yetanotherjob
    resource_pool: otherpool
    instances: 3
```

In this case the resource pool size will resolve to '5'.

## `(( merge ))`

Bring the current path in from the stub files that are being merged in.

e.g.:

```yaml
foo:
  bar:
    baz: (( merge ))
```

Will try to bring in `foo.bar.baz` from the first stub, or the second, etc.,
returning the value from the first stub that provides it.

If the corresponding value is not defined, it will return nil. This then has the
same semantics as reference expressions; a nil merge is an unresolved template.
See `||`.

### `<<: (( merge ))`

Merging of maps or lists with the content of the same element found in some stub.

** Attention **
This form of `merge` has a compatibility propblem. In versions before 1.0.8, this expression
was never parsed, only the existence of the key `<<:` was relevant. Therefore there are often 
usages of `<<: (( merge ))` where `<<: (( merge || nil ))` is meant. The first variant would
require content in at least one stub (as always for the merge operator). Now this expression
is evaluated correctly, but this would break existing manifest template sets, which use the
first variant, but mean the second. Therfore this case is explicitly handled to describe an
optional merge. If really a required merge is meant an additional explicit qualifier has to
be used (`(( merge required ))`).  

#### Merging maps

**values.yml**
```yaml
foo:
  a: 1
  b: 2
```

**template.yml**
```yaml
foo:
  <<: (( merge ))
  b: 3
  c: 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  a: 1
  b: 2
  c: 4
```

#### Merging lists

**values.yml**
```yaml
foo:
  - 1
  - 2
```

**template.yml**
```yaml
foo:
  - 3
  - <<: (( merge ))
  - 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  - 3
  - 1
  - 2
  - 4
```

### `- <<: (( merge on key ))`

`spiff` is able to merge lists of maps with a key field. Those lists are handled like maps with the value of the key field as key. By default the key `name` is used. But with the selector `on` an arbitrary key name can be specified for a list-merge expression.

e.g.:

```yaml
list:
  - <<: (( merge on key ))
  - key: alice
    age: 25
  - key: bob
    age: 24
```

merged with

```yaml
list:
  - key: alice
    age: 20
  - key: peter
    age: 13
```

yields

```yaml
list:
  - key: peter
    age: 13
  - key: alice
    age: 20
  - key: bob
    age: 24
```

If no insertion of new entries is desired (as requested by the insertion merge expression), but only overriding of existent entries, one existing key field can be prefixed with the tag `key:` to indicate a non-standard key name, for example `- key:key: alice`.

### `<<: (( merge replace ))`

Replaces the complete content of an element by the content found in some stub instead of doing a deep merge for the existing content.

#### Merging maps

**values.yml**
```yaml
foo:
  a: 1
  b: 2
```

**template.yml**
```yaml
foo:
  <<: (( merge replace ))
  b: 3
  c: 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  a: 1
  b: 2
```

#### Merging lists

**values.yml**
```yaml
foo:
  - 1
  - 2
```

**template.yml**
```yaml
foo:
  - <<: (( merge replace ))
  - 3
  - 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  - 1
  - 2
```

### `<<: (( foo ))`

Merging of maps and lists found in the same template or stub.

#### Merging maps

```yaml
foo:
  a: 1
  b: 2

bar:
  <<: (( foo )) # any dynaml expression
  b: 3
```

yields:

```yaml
foo:
  a: 1
  b: 2

bar:
  a: 1
  b: 3
```

#### Merging lists

```yaml
bar:
  - 1
  - 2

foo:
  - 3
  - <<: (( bar ))
  - 4
```

yields:

```yaml
bar:
  - 1
  - 2

foo:
  - 3
  - 1
  - 2
  - 4
```

A common use-case for this is merging lists of static ips or ranges into a list of ips. Another possibility is to use a single [concatenation expression](#-12-bar-).

### `<<: (( merge foo ))`

Merging of maps or lists with the content of an arbitrary element found in some stub (Redirecting merge). There will be no further (deep) merge with the element of the same name found in some stub. (Deep merge of lists requires maps with field `name`)

Redirecting merges can be used as direct field value, also. They can be combined with replacing merges like `(( merge replace foo ))`.

#### Merging maps

**values.yml**
```yaml
foo:
  a: 10
  b: 20
  
bar:
  a: 1
  b: 2
```

**template.yml**
```yaml
foo:
  <<: (( merge bar))
  b: 3
  c: 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  a: 1
  b: 2
  c: 4
```

Another way doing a merge with another element in some stub could also be done the traditional way:

**values.yml**
```yaml
foo:
  a: 10
  b: 20
  
bar:
  a: 1
  b: 2
```

**template.yml**
```yaml
bar: 
  <<: (( merge ))
  b: 3
  c: 4
  
foo: (( bar ))
```

But in this scenario the merge still performs the deep merge with the original element name. Therefore 
`spiff merge template.yml values.yml` yields:

```yaml
bar:
  a: 1
  b: 2
  c: 4
foo:
  a: 10
  b: 20
  c: 4
```

#### Merging lists

**values.yml**
```yaml
foo:
  - 10
  - 20

bar:
  - 1
  - 2
```

**template.yml**
```yaml
foo:
  - 3
  - <<: (( merge bar ))
  - 4
```

`spiff merge template.yml values.yml` yields:

```yaml
foo:
  - 3
  - 1
  - 2
  - 4
```

## `(( a || b ))`

Uses a, or b if a cannot be resolved.

e.g.:

```yaml
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

## `(( 1 + 2 * foo ))`

Dynaml expressions can be used to execute arithmetic integer calculations. Supported operations are +, -, *, / and %.

e.g.:

**values.yml**
```yaml
foo: 3
bar: (( 1 + 2 * foo ))
```

`spiff merge values.yml` yields `7` for `bar`. This can be combined with [concatentions](#-foo-bar-) (calculation has higher priority than concatenation in dynaml expressions):

```yaml
foo: 3
bar: (( foo " times 2 yields " 2 * foo ))
```
The result is the string `3 times 2 yields 6`.

## `(( "10.10.10.10" - 11 ))`

Besides arithmetic on integers it is also possible to use addition and subtraction on ip addresses.

e.g.:

```yaml
ip: 10.10.10.10
range: (( ip "-" ip + 247 + 256 * 256 ))
```

yields

```yaml
ip: 10.10.10.10
range: 10.10.10.10-10.11.11.1
```

Additionally there are functions working on CIDRs:

```yaml
cidr: 192.168.0.1/24
range: (( min_ip(cidr) "-" max_ip(cidr) ))
next: (( max_ip(cidr) + 1 ))
```

yields

```yaml
cidr: 192.168.0.1/24
range: 192.168.0.0-192.168.0.255
next: 192.168.1.0
```

## `(( join( ", ", list) ))`

Join entries of lists or direct values to a single string value using a given separator string. The arguments to join can be dynaml expressions evaluating to lists, whose values again are strings or integers, or string or integer values.

e.g.

```yaml
alice: alice
list:
  - foo
  - bar

join: (( join(", ", "bob", list, alice, 10) ))
```

yields the string value `bob, foo, bar, alice, 10` for `join`.

## `(( exec( "command", arg1, arg2) ))`

Execute a command. Arguments can be any dynaml expressions including reference expressions evaluated to lists or maps. Lists or maps are passed as single arguments containing a yaml document with the given fragment.

The result is determined by parsing the standard output of the command. It might be a yaml document or a single multi-line string or integer value. A yaml document must start with the document prefix `---`. If the command fails the expression is handled as undefined.

e.g.

```yaml
arg:
  - a
  - b
list: (( exec( "echo", arg ) ))
string: (( exec( "echo", arg.[0] ) ))

```

yields

```yaml
arg:
- a
- b
list:
- a
- b
string: a
```

Alternatively `exec` can be called with a single list argument completely describing the command line.

The same command will be executed once, only, even if it is used in multiple expressions.

## `(( static_ips(0, 1, 3) ))`

Generate a list of static IPs for a job.

e.g.:

```yaml
jobs:
  - name: myjob
    instances: 2
    networks:
    - name: mynetwork
      static_ips: (( static_ips(0, 3, 4) ))
```

This will create 3 IPs from `mynetwork`s subnet, and return two entries, as
there are only two instances. The two entries will be the 0th and 3rd offsets
from the static IP ranges defined by the network.

For example, given the file **template.yml**:

```yaml
networks: (( merge ))

jobs:
  - name: myjob
    instances: 3
    networks:
    - name: cf1
      static_ips: (( static_ips(0,3,60) ))
```

and file **stub.yml**:

```yaml
networks:
- name: cf1
  subnets:
  - cloud_properties:
      security_groups:
      - cf-0-vpc-c461c7a1
      subnet: subnet-e845bab1
    dns:
    - 10.60.3.2
    gateway: 10.60.3.1
    name: default_unused
    range: 10.60.3.0/24
    reserved:
    - 10.60.3.2 - 10.60.3.9
    static:
    - 10.60.3.10 - 10.60.3.70
  type: manual
```

```
spiff merge template.yml stub.yml
```

returns


```yaml
jobs:
- instances: 3
  name: myjob
  networks:
  - name: cf1
    static_ips:
    - 10.60.3.10
    - 10.60.3.13
    - 10.60.3.70
networks:
- name: cf1
  subnets:
  - cloud_properties:
      security_groups:
      - cf-0-vpc-c461c7a1
      subnet: subnet-e845bab1
    dns:
    - 10.60.3.2
    gateway: 10.60.3.1
    name: default_unused
    range: 10.60.3.0/24
    reserved:
    - 10.60.3.2 - 10.60.3.9
    static:
    - 10.60.3.10 - 10.60.3.70
  type: manual
```
.

If **bye.yml** was instead

```yaml
networks: (( merge ))

jobs:
  - name: myjob
    instances: 2
    networks:
    - name: cf1
      static_ips: (( static_ips(0,3,60) ))
```

```
spiff merge bye.yml hi.yml
```

instead returns

```yaml
jobs:
- instances: 2
  name: myjob
  networks:
  - name: cf1
    static_ips:
    - 10.60.3.10
    - 10.60.3.13
networks:
- name: cf1
  subnets:
  - cloud_properties:
      security_groups:
      - cf-0-vpc-c461c7a1
      subnet: subnet-e845bab1
    dns:
    - 10.60.3.2
    gateway: 10.60.3.1
    name: default_unused
    range: 10.60.3.0/24
    reserved:
    - 10.60.3.2 - 10.60.3.9
    static:
    - 10.60.3.10 - 10.60.3.70
  type: manual
```

## Mappings

### `(( map[list|elem|->dynaml-expr] ))`

Execute a mapping expression on members of a list to produce a new (mapped) list. The first expression (`list`) must resolve to a list. The last expression (`x ":" port`) defines the mapping expression used to map all members of the given list. Inside this expression an arbitrarily declared simple reference name (here `x`) can be used to access the actually processed list element.

e.g.

```yaml
port: 4711
hosts:
  - alice
  - bob
mapped: (( map[hosts|x|->x ":" port] ))
```

yields

```yaml
port: 4711
hosts:
- alice
- bob
mapped:
- alice:4711
- bob:4711
```

This expression can be combined with others, for example:

```yaml
port: 4711
list:
  - alice
  - bob
joined: (( join( ", ", map[list|x|->x ":" port] ) ))

```

which magically provides a comma separated list of ported hosts:

```yaml
port: 4711
list:
  - alice
  - bob
joined: alice:4711, bob:4711
```

### `(( map[list|idx,elem|->dynaml-expr] ))`

In this variant, the first argument `idx` is provided with the index and the
second `elem` with the value for the index.

e.g.

```yaml
list:
  - name: alice
    age: 25
  - name: bob
    age: 24
	
ages: (( map[list|i,p|->i + 1 ". " p.name " is " p.age ] ))
```
 
yields

```yaml
list:
  - name: alice
    age: 25
  - name: bob
    age: 24
	
ages:
- 1. alice is 25
- 2. bob is 24
```

### (( map[map|key,value|->dynaml-expr] ))

Mapping of a map to a list using a mapping expression. The expression may have access to the key and/or the value. If two references are declared, both values are passed to the expression, the first one is provided with the key and the second one with the value for the key. If one reference is declared, only the value is provided.

e.g.

```yaml
ages:
  alice: 25
  bob: 24

keys: (( map[ages|k,v|->k] ))

```

yields

```yaml
ages:
  alice: 25
  bob: 24

keys:
- alice
- bob
```

## Operation Priorities

Dynaml expressions are evaluated obeying certain priority levels. This means operations with a higher priority are evaluated first. For example the expression `1 + 2 * 3` is evaluated in the order `1 + ( 2 * 3 )`. Operations with the same priority are evaluated from left to right (in contrast to version 1.0.7). This means the expression `6 - 3 - 2` is evaluated as `( 6 - 3 ) - 2`.

The following levels are supported (from low priority to high priority)

1. `||`
2. White-space separated sequence as concatenation operation (`foo bar`)
3. `+`, `-`
4. `*`, `/`, `%`
5. Grouping `( )`, constants, references (`foo.bar`) and functions (`merge`, `auto`, `map[]`, `join`, `exec`, `static_ips`, `min_ip`, `max_ip`)

The complete grammar can be found in [dynaml.peg](dynaml/dynaml.peg).

# Structural Auto-Merge

By default `spiff` performs a deep structural merge of its first argument, the template file, with the given stub files. The merge is processed from right to left, providing an intermediate merged stub for every step. This means, that for every step all expressions must be locally resolvable. 

Structural merge means, that besides explicit dynaml `merge` expressions, values will be overridden by values of equivalent nodes found in right-most stub files. In general, flat value lists are not merged. Only lists of maps can be merged by entries in a stub with a matching index. 

There is a special support for the auto-merge of lists containing maps, if the
maps contain a `name` field. Hereby the list is handled like a map with
entries according to the value of the list entries' `name` field. If another
key field than `name` should be used, the key field of one list entry can be
tagged with the prefix `key:` to indicate the indended key name. Such tags
will be removed for the processed output.

In general the resolution of matching nodes in stubs is done using the same rules that apply for the reference expressions [(( foo.bar.[1].baz ))](#-foobar1baz-).

For example, given the file **template.yml**:

```yaml
foo:
  - name: alice
    bar: template
  - name: bob
    bar: template

plip:
  - id: 1
    plop: template
  - id: 2
    plop: template

bar:
  - foo: template

list:
  - a
  - b
```

and file **stub.yml**:

```yaml
foo: 
  - name: bob
    bar: stub

plip:
  - key:id: 1
    plop: stub

bar:
  - foo: stub

list:
  - c
  - d
```

```
spiff merge template.yml stub.yml
```

returns


```yaml
foo:
- bar: template
  name: alice
- bar: stub
  name: bob

plip:
- id: 1
  plop: stub
- id: 2
  plop: template

bar:
- foo: stub

list:
- a
- b
```

Be careful that any `name:` key in the template for the first element of the
`plip` list will defeat the `key:id: 1` selector from the stub. When a `name`
field exist in a list element, then this element can only be targeted by this
name. When the selector is defeated, the resulting value is the one provided
by the template.


## Useful to Know

  There are several scenarios yielding results that do not seem to be obvious. Here are some typical pitfalls.

- _The auto merge never adds nodes to existing structures_

  For example, merging
 
  **template.yml**
  ```yaml
  foo:
    alice: 25
  ```
  with

  **stub.yml**
  ```yaml
  foo:
    alice: 24
    bob: 26
  ```

   yields

  ```yaml
  foo:
    alice: 24
  ```

  Use [<<: (( merge ))](#--merge-) to change this behaviour, or explicitly add desired nodes to be merged:

   **template.yml**
  ```yaml
  foo:
    alice: 25
	bob: (( merge ))
  ```


- _Simple node values are replaced by values or complete structures coming from stubs, structures are deep_ merged.

  For example, merging
 
  **template.yml**
  ```yaml
  foo: (( ["alice"] ))
  ```
  with

  **stub.yml**
  ```yaml
  foo: 
    - peter
    - paul
  ``` 

  yields

  ```yaml
  foo:
    - peter
    - paul 
  ```

  But the template

  ```yaml
   foo: [ (( "alice" )) ] 
  ```

  is merged without any change.

- _Expressions are subject to be overridden as a whole_
  
  A consequence of the behaviour described above is that nodes described by an expession are basically overridden by a complete merged structure, instead of doing a deep merge with the structues resulting from the expression evaluation.

  For example, merging
 
  **template.yml**
  ```yaml
  men:
    - bob: 24
  women:
    - alice: 25
	
  people: (( women men ))
  ```
  with

  **stub.yml**
  ```yaml
  people:
    - alice: 13
  ```
   yields

  ```yaml
  men:
    - bob: 24
  women:
    - alice: 25
	
  people:
    - alice: 24
  ```

  To request an auto-merge of the structure resulting from the expression evaluation, the expression has to be preceeded with the modifier `prefer` (`(( prefer women men ))`). This would yield the desired result:

  ```yaml
  men:
    - bob: 24
  women:
    - alice: 25
	
  people:
    - alice: 24
    - bob: 24
  ```

- _Nested merge expressions use implied redirections_

  `merge` expressions implicity use a redirection implied by an outer redirecting merge. In the following
  example

  ```yaml
  meta:
    <<: (( merge deployments.cf ))
    properties:
      <<: (( merge ))
      alice: 42
  ```
  the merge expression in `meta.properties` is implicity redirected to the path `deployments.cf.properties`
  implied by the outer redirecting `merge`. Therefore merging with

  ```yaml
  deployments:
    cf:
      properties:
        alice: 24
        bob: 42
  ```

  yields

  ```yaml
  meta:
    properties:
      alice: 24
      bob: 42
  ```

# Error Reporting

The evaluation of dynaml expressions may fail because of several reasons:
- it is not parseable
- involved references cannot be satisfied
- arguments to operations are of the wrong type
- operations fail
- there are cyclic dependencies among expressions

If a dynaml expression cannot be resolved to a value, it is reported by the
`spiff merge` operation using the following layout:

```
	(( <failed expression> ))	in <file>	<path to node>	(<referred path>)	<issue>
```
	
e.g.:

```	
	(( min_ip("10") ))	in source.yml	node.a.[0]	()	CIDR argument required
```
	
Cyclic dependencies are detected by iterative evaluation until the document is unchanged after a step.
Nodes involved in a cycle are therefore typically reported without an issue.
