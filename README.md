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
	- [(( [ 1, 2, 3 ] ))](#--1-2-3--)
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
	- [(( a > 1 ? foo :bar ))](#-a--1--foo-bar-)
	- [(( 5 -or 6 ))](#-5--or-6-)
	- [Functions](#functions)
		- [(( format( "%s %d", alice, 25) ))](#-format-s-d-alice-25-)
		- [(( join( ", ", list) ))](#-join---list-)
		- [(( split( ",", string) ))](#-split--string-)
		- [(( trim(string) ))](#-trimstring-)
		- [(( length(list) ))](#-lengthlist-)
		- [(( defined(foobar) ))](#-definedfoobar-)
		- [(( exec( "command", arg1, arg2) ))](#-exec-command-arg1-arg2-)
		- [(( eval( foo "." bar ) ))](#-eval-foo--bar--)
		- [(( env( "HOME" ) ))](#-env-HOME--)
		- [(( read("file.yml") ))](#-readfileyml-)
		- [(( static_ips(0, 1, 3) ))](#-static_ips0-1-3-)
	- [(( lambda |x|->x ":" port ))](#-lambda-x-x--port-)
	- [Mappings](#mappings)
		- [(( map[list|elem|->dynaml-expr] ))](#-maplistelem-dynaml-expr-)
		- [(( map[list|idx,elem|->dynaml-expr] ))](#-maplistidxelem-dynaml-expr-)
		- [(( map[map|key,value|->dynaml-expr] ))](#-mapmapkeyvalue-dynaml-expr-)
	- [Templates](#templates)
		- [<<: (( &template ))](#--template-)
		- [(( *foo.bar ))](#-foobar-)
	- [Operation Priorities](#operation-priorities)
- [Structural Auto-Merge](#structural-auto-merge)
- [Bringing it all together](#bringing-it-all-together)
- [Useful to Know](#useful-to-know)
- [Error Reporting](#error-reporting)


# Installation

Official release executable binaries can be downloaded via [Github releases](https://github.com/cloudfoundry-incubator/spiff/releases) for Darwin and Linux machines (and virtual machines).

Some of spiff's dependencies have changed since the last official release, and spiff will not be updated to keep up with these dependencies.  Working dependencies are vendored in the `Godeps` directory (more information on the `godep` tool is available [here](https://github.com/tools/godep)).  As such, trying to `go get` spiff will likely fail; the only supported way to use spiff is to use an official binary release.

# Usage

### `spiff merge template.yml [template2.ymll ...]`

Merge a bunch of template files into one manifest, printing it out.

See 'dynaml templating language' for details of the template file, or examples/ subdir for more complicated examples.

Example:

```
spiff merge cf-release/templates/cf-deployment.yml my-cloud-stub.yml
```

### `spiff diff manifest.yml other-manifest.yml`

Show structural differences between two deployment manifests.

Unlike basic diffing tools and even `bosh diff`, this command has semantic 
knowledge of a deployment manifest, and is not just text-based. For example,
if two manifests are the same except they have some jobs listed in different
orders, `spiff diff` will detect this, since job order matters in a manifest.
On the other hand, if two manifests differ only in the order of their
resource pools, for instance, then it will yield and empty diff since 
resource pool order doesn't actually matter for a deployment.

Also unlike `bosh diff`, this command doesn't modify either file.

It's tailored for checking differences between one deployment and the next.

Typical flow:

```sh
$ spiff merge template.yml [templates...] > upgrade.yml
$ bosh download manifest [deployment] current.yml
$ spiff diff upgrade.yml current.yml
$ bosh deployment upgrade.yml
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

## `(( [ 1, 2, 3 ] ))`

List literal. The list elements might again be expressions. There is a special list literal `[1 .. -1]`, that can be used to resolve an increasing or descreasing number range to a list. 

e.g.:

```yaml
list: (( [ 1 .. -1 ] ))
```

yields

```yaml
list:
  - 1
  - 0
  - -1
```

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

Additionally there are functions working on IPv4 CIDRs:

```yaml
cidr: 192.168.0.1/24
range: (( min_ip(cidr) "-" max_ip(cidr) ))
next: (( max_ip(cidr) + 1 ))
num: (( min_ip(cidr) "+" num_ip(cidr) "=" min_ip(cidr) + num_ip(cidr) ))
```

yields

```yaml
cidr: 192.168.0.1/24
range: 192.168.0.0-192.168.0.255
next: 192.168.1.0
num: 192.168.0.0+256=192.168.1.0
```

## `(( a > 1 ? foo :bar ))`

Dynaml supports the comparison operators `<`, `<=`, `==`, `!=`, `>=` and `>`. The comparison operators work on
integer values. The check for equality also works on lists and maps. The result is always a boolean value.

Additionally there is the ternary conditional operator `?:`, that can be used to evaluate expressions depending on a condition. The first operand is used as condition. The expression is evaluated to the second operand, if the condition is true, and to the third one, otherwise.

e.g.:

```yaml
foo: alice
bar: bob
age: 24
name: (( age > 24 ? foo :bar ))
```

yields the value `bob` for the property `name`.

**Remark**

The use of the symbol `:` may collide with the yaml syntax, if the complete expression is not a quoated string value.

The operators `-or` and `-and` can be used to combine comparison operators to compose more complex conditions.

**Remark:**

The more traditional operator symbol `||` (and `&&`) cannot be used here, because the operator `||` already exists in dynam with a different semantic, that does not hold for logical operations. The expression `false || true` evaluates to `false`, because it yields the first operand, if it is defined, regardless of its value. To be as compatible as possible this cannot be changed and the bare symbols `or` and `and` cannot be be used, because this would invalidate the concatenation of references with such names. 

## `(( 5 -or 6 ))`

If both sides of an `-or` or `-and` operator evaluate to integer values, a bit-wise operation is executed and the result is again an integer. Therefore the expression `5 -or 6` evaluates to `7`.

## Functions

Dynaml supports a set of predefined functions. A function is generally called like

```yaml
result: (( functionname(arg, arg, ...) ))
```

Additional functions may be defined as part of the yaml document using [lambda expressions](#-lambda-x-x--port-). The function name then is either a grouped expression or the path to the node hosting the lambda expression.
 
### `(( format( "%s %d", alice, 25) ))`

Format a string based on arguments given by dynaml expressions. There is a second flavor of this function: `error` formats an error message and sets the evaluation to failed.
  

### `(( join( ", ", list) ))`

Join entries of lists or direct values to a single string value using a given separator string. The arguments to join can be dynaml expressions evaluating to lists, whose values again are strings or integers, or string or integer values.

e.g.:

```yaml
alice: alice
list:
  - foo
  - bar

join: (( join(", ", "bob", list, alice, 10) ))
```

yields the string value `bob, foo, bar, alice, 10` for `join`.

### `(( split( ",", string) ))`

Split a string for a dedicated separator. The result is a list.

e.g.:

```yaml
list: (( split("," "alice, bob") ))
```

yields:

```yaml
list:
  - alice
  - ' bob'
```

### `(( trim(string) ))`

Trim a string or all elements of a list of strings. There is an optional second string argument. It can be used to specify a set of characters that will be cut. The default cut set consists of a space and a tab character.

e.g.:

```yaml
list: (( trim(split("," "alice, bob")) ))
```

yields:

```yaml
list:
  - alice
  - bob
```

### `(( length(list) ))`

Determine the length of a list, a map or a string value.

e.g.:

```yaml
list:
  - alice
  - bob
length: (( length(list) ))
```

yields:

```yaml
list:
  - alice
  - bob
length: 2
```

### `(( defined(foobar) ))`

The function `defined` checks whether an expression can successfully be evaluated. It yields the boolean value `true`, if the expression can be evaluated, and `false` otherwise.

e.g.:

```yaml
zero: 0
div_ok: (( defined(1 / zero ) ))
zero_def: (( defined( zero ) ))
null_def: (( defined( null ) ))
```

evaluates to

```yaml
zero: 0
div_ok: false
zero_def: true
null_def: false
```

This function can be used in combination of the [conditional operator](#-a--1--foo-bar-) to evaluate expressions depending on the resolvability of another expression.

### `(( exec( "command", arg1, arg2) ))`

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

### `(( eval( foo "." bar ) ))`

Evaluate the evaluation result of a string expression again as dynaml expression. This can, for example, be used to realize indirections.

e.g.: the expression in

```yaml
alice:
  bob: married

foo: alice
bar: bob

status: (( eval( foo "." bar ) ))
```

calculates the path to a field, which is then evaluated again to yield the value of this composed field:

```yaml
alice:
  bob: married

foo: alice
bar: bob

status: married
```

### `(( env( "HOME" ) ))`

Read the value of an environment variable whose name is given as dynaml expression. If the environment variable is not set the evaluation fails.

In a second flavor the function `env` accepts multiple arguments and/or list arguments, which are joined to a single list. Every entry in this list is used as name of an environment variable and the result of the function is a map of the given given variables as yaml element. Hereby non-existent environment variables are omitted.

### `(( read("file.yml") ))` 

Read a file and return its content. There is support for two content types: `yaml` files and `text` files.
If the file suffix is `.yml`, by default the yaml type is used. An optional second parameter can be used
to explicitly specifiy the desired return type: `yaml` or `text`.

#### yaml documents
A yaml document will be parsed and the tree is returned. The  elements of the tree can be accessed by regular dynaml expressions.

Additionally the yaml file may again contain dynaml expressions. All included dynaml expressions will be evaluated in the context of the reading expression. This means that the same file included at different places in a yaml document may result in different sub trees, depending on the used dynaml expressions. 

#### text documents
A text document will be returned as single string.

### `(( static_ips(0, 1, 3) ))`

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

For example, given the file **bye.yml**:

```yaml
networks: (( merge ))

jobs:
  - name: myjob
    instances: 3
    networks:
    - name: cf1
      static_ips: (( static_ips(0,3,60) ))
```

and file **hi.yml**:

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
spiff merge bye.yml hi.yml
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

## `(( lambda |x|->x ":" port ))`

Lambda expressions can be used to define additional anonymous functions. They can be assigned to yaml nodes as values and referenced with path expressions to call the function with approriate arguments in other dynaml expressions. For the final document they are mapped to string values.

There are two forms of lambda expressions. While

```yaml
lvalue: (( lambda |x|->x ":" port ))
```

yields a function taking one argument by directly taking the elements from the dynaml expression,

```yaml
string: "|x|->x \":\" port"
lvalue: (( lambda string ))
```

evaluates an expression to a function or a string. If the expression is evaluated to a string it parses the function from the string.

Since the evaluation result of a lambda expression is a regular value, it can also be passed as argument to function calls and merged as value along stub processing.

A complete example could look like this:

```yaml
lvalue: (( lambda |x,y|->x + y ))
mod: (( lambda|x,y,m|->(lambda m)(x, y) + 3 ))
value: (( .mod(1,2, lvalue) ))
```

yields

```yaml
lvalue: lambda |x,y|->x + y
mod: lambda|x,y,m|->(lambda m)(x, y) + 3
value: 6
```

A lambda expression might refer to absolute or relative nodes of the actual template. Relative references are evaluated in the context of the function call. Therefore

```yaml
lvalue: (( lambda |x,y|->x + y + offset ))
offset: 0
values:
  offset: 3
  value: (( .lvalue(1,2) ))
```

yields `6` for `values.value`.

Besides the specified parameters, there is an implicit name (`_`), that can be used to refer to the function itself. It can be used to define self recursive function. Together with the logical and conditional operators a fibunacci function can be defined:

```yaml
fibonacci: (( lambda |x|-> x <= 0 ? 0 :x == 1 ? 1 :_(x - 2) + _( x - 1 ) ))
value: (( .fibonacci(5) ))
```

yields the value `8` for the `value` property.

Inner lambda expressions remember the local binding of outer lambda expressions. This can be used to return functions based an arguments of the outer function.

e.g.:

```yaml
mult: (( lambda |x|-> lambda |y|-> x * y ))
mult2: (( .mult(2) ))
value: (( .mult2(3) ))
```

yields `6` for property `value`.

If a lambda function is called with less arguments than expected, the result is a new function taking the missing arguments (currying).

e.g.:

```yaml
mult: (( lambda |x,y|-> x * y ))
mult2: (( .mult(2) ))
value: (( .mult2(3) ))
```

If a complete expression is a lambda expression the keyword `lambda` can be omitted.

## Mappings

Mappings are used to produce a new list from the entries of a _list_ or _map_ containing the entries processed by a dynaml expression. The expression is given by a [lambda function](#-lambda-x-x--port-). There are two basic forms of the mapping function: It can be inlined as in`(( map[list|x|->x ":" port] ))`, or it can be determined by a regular dynaml expression evaluating to a lambda function as in `(( map[list|mapping.expression))` (here the mapping is taken from the property `mapping.expression`, which should hold an approriate lambda function).


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

### `(( map[map|key,value|->dynaml-expr] ))`

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

## Templates

A map can be tagged by a dynaml expression to be used as template. Dynaml expressions in a template are not evaluated at its definition location in the document, but can be inserted at other locations using dynaml.
At every usage location it is evaluated separately.

### `<<: (( &template ))`

The dynaml expression `&template` can be used to tag a map node as template:

i.g.:

```yaml
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))
```

The template will be the value of the node `foo.bar`. As such it can be overwritten as a whole by settings in a stub during the merge process. Dynaml expressions in the template are not evaluated.

### `(( *foo.bar ))`

The dynaml expression `*<refernce expression>` can be used to evaluate a template somewhere in the yaml document.
Dynaml expressions in the template are evaluated in the context of this expression.

e.g.:

```yaml
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))


use:
  subst: (( *foo.bar ))
  verb: loves

verb: hates
```

evaluates to

```yaml
foo:
  bar:
    <<: (( &template ))
    alice: alice
    bob: (( verb " " alice ))
	
use:
  subst:
    alice: alice
    bob: loves alice
  verb: loves

verb: hates
```


## Operation Priorities

Dynaml expressions are evaluated obeying certain priority levels. This means operations with a higher priority are evaluated first. For example the expression `1 + 2 * 3` is evaluated in the order `1 + ( 2 * 3 )`. Operations with the same priority are evaluated from left to right (in contrast to version 1.0.7). This means the expression `6 - 3 - 2` is evaluated as `( 6 - 3 ) - 2`.

The following levels are supported (from low priority to high priority)

1. `||`
2. White-space separated sequence as concatenation operation (`foo bar`)
3. `-or`, `-and`
4. `==`, `!=`, `<=`, `<`, `>`, `>=`
5. `+`, `-`
6. `*`, `/`, `%`
7. Grouping `( )`, `!`, constants, references (`foo.bar`), `merge`, `auto`, `lambda`, `map[]`, and [functions](#functions)

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

## Bringing it all together

Merging the following files in the given order

**deployment.yml**
```yaml
networks: (( merge ))
```

**cf.yml**
```yaml
utils: (( merge )) 
network: (( merge ))
meta: (( merge ))

networks:
  - name: cf1
    <<: (( utils.defNet(network.base.z1,meta.deployment_no,30) ))
  - name: cf2
    <<: (( utils.defNet(network.base.z2,meta.deployment_no,30) ))
```

**infrastructure.yml**
```yaml
network:
  size: 16
  block_size: 256
  base:
    z1: 10.0.0.0
    z2: 10.1.0.0
```

**rules.yml**
```yaml
utils:
  defNet: (( |b,n,s|->(*.utils.network).net ))
  network:
    <<: (( &template ))
    start: (( b + n * .network.block_size ))
    first: (( start + ( n == 0 ? 2 :0 ) ))
    lower: (( n == 0 ? [] :b " - " start - 1 ))
    upper: (( start + .network.block_size " - " max_ip(net.subnets.[0].range) ))
    net:
      subnets:
      - range: (( b "/" .network.size ))
        reserved: (( [] lower upper ))
        static:
          - (( first " - " first + s - 1 ))
```

**instance.yml**
```yaml
meta:
  deployment_no: 1
  
```

will yield a network setting for a dedicated deployment

```yaml
networks:
- name: cf1
  subnets:
  - range: 10.0.0.0/16
    reserved:
    - 10.0.0.0 - 10.0.0.255
    - 10.0.2.0 - 10.0.255.255
    static:
    - 10.0.1.0 - 10.0.1.29
- name: cf2
  subnets:
  - range: 10.1.0.0/16
    reserved:
    - 10.1.0.0 - 10.1.0.255
    - 10.1.2.0 - 10.1.255.255
    static:
    - 10.1.1.0 - 10.1.1.29
```

Using the same config for another deployment of the same type just requires the replacement of the `instance.yml`.
Using a different `instance.yml`

```yaml
meta:
  deployment_no: 0
  
```

will yield a network setting for a second deployment providing the appropriate settings for a unique other IP block.

```yaml
networks:
- name: cf1
  subnets:
  - range: 10.0.0.0/16
    reserved:
    - 10.0.1.0 - 10.0.255.255
    static:
    - 10.0.0.2 - 10.0.0.31
- name: cf2
  subnets:
  - range: 10.1.0.0/16
    reserved:
    - 10.1.1.0 - 10.1.255.255
    static:
    - 10.1.0.2 - 10.1.0.31
```

If you move to another infrastructure you might want to change the basic IP layout. You can do it just by adapting the `infrastructure.yml`

```yaml
network:
  size: 17
  block_size: 128
  base:
    z1: 10.0.0.0
    z2: 10.0.128.0
```

Without any change to your other settings you'll get

```yaml
networks:
- name: cf1
  subnets:
  - range: 10.0.0.0/17
    reserved:
    - 10.0.0.128 - 10.0.127.255
    static:
    - 10.0.0.2 - 10.0.0.31
- name: cf2
  subnets:
  - range: 10.0.128.0/17
    reserved:
    - 10.0.128.128 - 10.0.255.255
    static:
    - 10.0.128.2 - 10.0.128.31
```

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

- _Functions and mappings can freely be nested_

  e.g.:

  ```yaml
  pot: (( lambda |x,y|-> y == 0 ? 1 :(|m|->m * m)(_(x, y / 2)) * ( 1 + ( y % 2 ) * ( x - 1 ) ) ))
  seq: (( lambda |b,l|->map[l|x|-> .pot(b,x)] ))
  values: (( .seq(2,[ 0..4 ]) ))
  ```

  yields the list `[ 1,2,4,8,16 ]` for the property `values`.

- _Functions can be used to parameterize templates_

  The combination of functions with templates can be use to provide functions yielding complex structures.
  The parameters of a function are part of the scope used to resolve reference expressions in a template used in the function body.

  e.g.:

  ```yaml
  relation:
    template:
      <<: (( &template ))
      bob: (( x " " y ))
    relate: (( |x,y|->*relation.template ))

  banda: (( relation.relate("loves","alice") ))
  ```

  evaluates to

  ```yaml
  relation:
    relate: lambda|x,y|->*(relation.template)
    template:
      <<: (( &template ))
      bob: (( x " " y ))
	
	banda:
      bob: loves alice
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
