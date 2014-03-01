<pre>
                                       _  __  __
                             ___ _ __ (_)/ _|/ _|
                            / __| '_ \| | |_| |_
                            \__ \ |_) | |  _|  _|
                            |___/ .__/|_|_| |_|
                                |_|
                                
             A declarative templating system for BOSH deployment manifests.
</pre>


# pre-reqs

```
  # you must have bzr installed
  $ which bzr
  /usr/local/bin/bzr
  
  # if you do not have it, get it with something like
  $ brew install bzr
```

## installation

```
# set up $GOPATH if not already
  export GOPATH=~/go
  export PATH=~/go/bin:$PATH

  # install
  go get -v github.com/cloudfoundry-incubator/spiff

  # to update a previously installed version of spiff
  go get -u github.com/cloudfoundry-incubator/spiff

  # run
  spiff
```

## development

```
  # install dependencies
  go get github.com/xoebus/gocart
  gocart install

  # run tests
  go test -v ./...

  # or, with ginkgo:
  go install github.com/onsi/ginkgo/ginkgo
  ginkgo -r
```

## spiff merge template.yml [template2.yml template3.yml ...] > manifest.yml

  Merge a bunch of template files into one manifest, printing it out.

  See 'dynaml templating language' for details of the template file, or
  example.yml for a complicated example.

  Example:

```
    spiff merge cf-release/templates/cf-deployment.yml my-cloud-stub.yml
```

## spiff diff manifest.yml other-manifest.yml

  Show structural differences between two deployment manifests.

  Unlike 'bosh diff', this command has semantic knowledge of a deployment
  manifest, and is not just text-based. It also doesn't modify either file.

  It's tailed for checking differences between one deployment and the next.

  Typical flow:
    1. spiff merge template.yml [stubs...] > deployment.yml
      1a. bosh download manifest [deployment] current.yml
      1b. spiff diff deployment.yml current.yml
    2. bosh deployment deployment.yml
    3. bosh deploy


## dynaml templating language

Spiff uses a declarative, logic-free templating language called 'dynaml'
(dynamic yaml).

Every dynaml node is guaranteed to resolve to a YAML node. It is *not*
string interpolation. This keeps developers from having to think about how
a value will render in the resulting template.

A dynaml node appears in the .yml file as an expression surrounded by two
parentheses. They can be used as the value of a map or an entry in a list.

The following is a complete list of dynaml expressions:

#### ```(( foo ))```

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

#### ``` (( foo.bar.[1].baz )) ```

Look for the nearest ```foo``` key, and from there follow through to .bar.baz.

A path is a sequence of steps separated by dots. A step is either a word
for maps, or digits surrounded by brackets for list indexing.

If the path cannot be resolved, this evaluates to nil. A reference node at
the top level cannot evaluate to nil; the template will be considered not
fully resolved. If a reference is expected to sometimes not be provided,
it should be used in combination with '||' (see below) to guarantee
resolution.

Note that references are always within the template, and order does not
matter. You can refer to another dynamic node and presume it's resolved,
and the reference node will just eventually resolve once the dependent
node resolves.

e.g.:

```
      properties:
        foo: (( something.from.the.stub ))
        something: (( merge ))
```

This will resolve as long as ```something``` is resolveable, and as long as it
brings in something like this:

```
      from:
        the:
          stub: foo
```

#### ```(( "foo" ))```

String literal. The only escape character handled currently is ``` " ```.

#### ```(( "foo" bar ))```
    
Concatenation (where bar is another arbitrary expr).

e.g.
```
      domain: example.com
      uri: (( "https://" domain ))
```
In this example 'uri' will resolve to the value 'https://example.com'.

#### ```(( auto ))```
    
Context-sensitive automatic value calculation.

In a resource pool's 'size' attribute, this means calculate based on the
total instances of all jobs that declare themselves to be in the current
resource pool.

e.g.:

```
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

#### ```(( merge ))```

Bring the current path in from the stub files that are being merged in.

e.g.:

```
      foo:
        bar:
          baz: (( merge ))
```

Will try to bring in (( foo.bar.baz )) from the first stub, or the second,
etc., returning the value from the first stub that provides it.

If the corresponding value is not defined, it will return nil. This then
has the same semantics as reference expressions; a nil merge is an
unresolved template. See '||'.

#### ```(( a || b ))```

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
This will try to merge in (( mything.complicated_structure )), or, if it
cannot be merged in, use the default specified in (( foo.bar )).

#### ``` (( static_ips(0, 1, 3) )) ```

Create the static IPs for a job's network.

e.g.:
```
      jobs:
      - name: myjob
        instances: 2
        networks:
        - name: mynetwork
          static_ips: (( static_ips(0, 3, 4) ))
```

This will create 3 IPs from 'mynetwork's subnet, and return two entries,
as there are only two instances. The two entries will be the 0th and 3rd
offsets from the static IP ranges defined by the network.
