# AppC Binary Discovery (abd)

## Overview 

abd, AppC Binary Discovery, defines a general framework for converting a
human-readable string to a downloadable artifact URI. It supports a federated
namespace, but with configurable overrides and multiple discovery mechanisms.
It is transport-agnostic and provides a simple and extensible interface.

## Motivation

Create a clean layer for naming and transport that can be used by different
container specifications and runtimes. We take inspiration from existing
systems such as rpm, apt, AppC, Docker, and others.

## Identifiers and Label Selectors

An artifact fetched through ABD is discovered based on a identifier and a set
of labels. The identifier is simply a string, and the labels are key value
pairs, where each key should be unique.

An identifier can be any arbitrary string, but it is recommended to have it be
a domain name and a path. For example, `appc.io/foo/bar`. This makes it easy to
see who is responsible for the given image, and can help with discovery of the
image using the `https-dns` strategy.

## ABD Metadata Format 

Applying a metadata fetch strategy to an identifier and set of labels returns a
blob of JSON in the ABD Metadata Format, which is a list with objects that
contain the following data:

- _Name_: the image name
- _Labels_: the set of labels which correspond to the image
- _Mirrors_: a list of objects, where each object contains a URI which the
  artifact can be retrieved from, and (optionally) a URI which a signature for
  the artifact can be retrieved from.

Note that the metadata that is produced can contain images with different
identifiers and/or labels than what was requested, and the list must be
filtered based on the desired identifier and labels before a mirror can be
selected.

_OPEN QUESTION: perhaps this is just TUF, and we embed all of the abd stuff
inside the TUF metadata?_

_OPEN QUESTION: do we want to define an interface for fetching URIs in the
mirrors? This could be used by `abd fetch`. For example, we would just pass
`http(s)://` to `wget`, or `hdfs://` to `hadoop fs get`, and so on_

Example ABD Metadata Format blob:
```
[
 {
  "name": "coreos.com/etcd",
  "labels": {
   "version": "1.0.0",
   "arch": "amd64",
   "os": "linux",
   "content-type": "aci"
  },
  "mirrors": [
   {
    "artifact": "https://github.com.../etcd-linux-amd64-1.0.0.aci",
    "signature": "https://github.com.../etcd-linux-amd64-1.0.0.aci.asc"
   }
  ]
 },
 {
  "name": "coreos.com/etcd",
  "labels": {
   "version": "1.0.0",
   "arch": "amd64",
   "os": "linux",
   "content-type": "docker",
  },
  "mirrors": [
   {
    "artifact": "docker://quay.io/coreos/etcd:1.0.0"
   },
   {
    "artifact": "docker://gcr.io/coreos/etcd:1.0.0"
   }
  ]
 }
]
```

## abd resolution process

The ABD resolution process consists of two steps. The first step is referred to
as _discovery_, and the second step is referred to as _filtering_.

In the _discovery_ step an identifier and label selectors are provided to a
metadata fetch strategy, which produces an ABD metadata blob.

In the _filtering_ step the list of mirrors provided by the ABD metadata blob
are filtered based on the provided identifier and label selectors, and a
suitable list of mirrors is produced.

## abd tool

We illustrate ABD with a simple command-line tool.

### `abd discover`

`abd discover` takes an identifier and a set of labels, applies an appropriate
metadata fetch strategy (chosen based on configuration described later in this
document), and yields a blob of JSON in the ABD Metadata Format.

```
abd discover coreos.com/etcd,content-type=aci,os=linux,arch=amd64
```

### `abd filter`

`abd filter` takes an identifier and a set of labels, reads in a blob of JSON
in the ABD Metadata Format from stdin, filters down artifacts to only ones with
a matching identifier and labels, and yields a list of mirrors.

```
cat <<EOF | abd filter coreos.com/etcd,content-type=aci,os=linux,arch=amd64
[
 {
   "name": "coreos.com/etcd",
   "labels": {
    "content-type": "aci",
    "os": "linux",
    "arch": "amd64"
   },
   "mirrors": [
    {
     "artifact": "https://github.com.../etcd-linux-amd64.aci",
     "signature": "https://github.com.../etcd-linux-amd64.aci.asc"
    }
   ]
 },
 {
   "name": "coreos.com/fleet",
   "labels": {
    "content-type": "aci",
    "os": "linux",
    "arch": "amd64"
   },
   "mirrors": [
    {
     "artifact": "https://github.com.../fleet-linux-amd64.aci",
     "signature": "https://github.com.../fleet-linux-amd64.aci.asc"
    }
   ]
 }
]
EOF
```

### `abd mirrors`

`abd mirrors` combines `abd discover` and `abd filter`, and will convert an
identifier and a set of labels into a list of URIs.

```
abd mirrors coreos.com/etcd,content-type=aci,os=linux,arch=amd64
```

### `abd fetch`

`abd fetch` takes `abd mirrors` one step further, by selecting one of the URIs
and actually fetching the artifact.

```
abd fetch coreos.com/etcd,content-type=aci,os=linux,arch=amd64
```

## ABD metadata fetch strategies 

ABD provides a simple interface for implementing fetch strategy plugins, and
ships with some plugins for common strategies.

Each strategy exists as an independent binary, with the desired identifier and
labels presented as arguments and a configuration (specified later) given to
the binary on stdin. Once a strategy has generated an ABD metadata blob (or an
error), the results are written to stdout.

If an error is encountered, the strategy should write out a JSON object
containing a single key named `error`, with its value a string describing the
error that was encountered.

### Well-known strategies (i.e., would ship with `abd`)

#### `https-dns`

This strategy looks for metadata over HTTPS, making consecutive requests
walking up the path of the image. It passes labels as query parameters, which
the server can optionally use to provide server side filtering (as an
optimisation).

An image identifier ultimately contains a domain name and a path, separated by
a `/`. The `https-dns` strategy splits the domain name and path, and assembles
a URI with these components and the provided labels according to the following
scheme:

```
https://{domain}/.abd/{path}?{labels}
```

For example, if the image name `example.com/foo/bar` is provided, with the
labels `label1=baz1` and `label2=baz2`the resulting URI will be:

```
https://example.com/.abd/foo/bar?label1=baz1&label2=baz2
```

If a GET request to this URI does not result in an appropriate JSON blob, the
`https-dns` strategy will walk up the path, making a request to:

```
https://example.com/.abd/foo?label1=baz1&label2=baz2
```

And if this fails, the final attempt will make a request to:

```
https://example.com/.abd?label1=baz1&label2=baz2
```

The motivation for inserting the `.abd` string into the URI is to allow for
user facing HTML files to exist at the actual URI represented by the
identifier.

#### `noop`

This strategy always fails; it is intended to be used to block retrievals for a
certain prefix. It returns an empty metadata blob (or, equivalently, one with
an empty list of mirrors).

#### `template`

This strategy substitutes in the provided identifier and labels to a template
stored in its configuration, and returns the results.

The template must be a JSON array with objects in it, where each object must
have the `mirrors` key and not include the `name` or `labels` keys. The `name`
and `label` keys are automatically filled in based on the identifier and labels
provided to the strategy.

Any key or string in the `template` that is enclosed in a `<` and a `>` is
replaced based on the characters between the angle brackets, where the contest
must be either `identifier` or one of the provided label keys.

If the following was the template set in this strategy's configuration:

```
[
 {
  "artifact": "https://example.com/<os>/<arch>/<identifier>-<version>.aci",
  "signature": "https://example.com/<os>/<arch>/<identifier>-<version>.aci.asc"
 }
]
```

The resulting ABD metadata blob returned by this strategy for the query
`coreos.com/etcd,content-type=aci,os=linux,arch=amd64,version=2.2.0`
would be:

```
[
 {
  "name": "coreos/com/etcd",
  "labels": {
   "content-type": "aci",
   "os": "linux",
   "arch": "amd64",
   "version": "2.2.0"
  },
  "mirrors": [
   {
    "artifact": "https://example.com/linux/amd64/coreos.com/etcd-2.2.0.aci",
    "signature": "https://example.com/linux/amd64/coreos.com/etcd-2.2.0.aci.asc",
   }
  ]
 }
]
```

## ABD Client Configuration (metadata fetch strategy configuration)

The configuration defines how the abd tool chooses which metadata fetch
strategy to use, given a certain identifier and labels.

The configuration format is loosely inspired by Debian apt repositories'
[sources.list configuration][apt-sources-list]. 
- Lexically-ordered configuration files, each containing a single strategy
  configuration
- Given an identifier, the first strategy with a prefix that matches that
  identifier is used
- Each configuration has two fields prescribed by abd: `prefix` and `strategy`
- All other fields defined in the configuration are passed unaltered to the
  strategy

Below are some example configurations for various use cases.

### Example configurations

#### `https-dns`

```
{
 "prefix": "*",
 "strategy": "abd/https-dns"
}
```

#### `noop`

```
{
    "prefix": "*",
    "strategy": "abd/noop"
}
```

#### `template`

```
{
    "prefix": "*",
    "strategy": "abd/template",
    "template": [
        {
            "artifact": "https://example.com/<version>/<arch>/<os>/<identifier>-<version>.aci",
            "signature": "https://example.com/<version>/<arch>/<os>/<identifier>-<version>.aci.asc"
        }
    ]
}

```

#### Scenario: default configuration: use https+dns for all

_OPEN QUESTION: this is what we want the "default" abd behaviour to be; but
perhaps rather than having it implicit, we could keep it explicit, i.e. set the
expectation that it is actually shipped as a default configuration file with
abd_

```
$ ls /usr/lib/abd/sources.list.d
zz-default.conf
$ cat /usr/lib/abd/sources.list.d/zz-default.conf
{
 "prefix": "*",
 "strategy": "abd/https-dns"
}
```

#### Scenario: local store for `coreos` images, + `https+dns` for everything
#### else

```
$ ls /usr/lib/abd/sources.list.d/
10-local.conf 
zz-default.conf
$ cat /usr/lib/abd/sources.list.d/10-local.conf
{
    "prefix": "coreos.com",
    "strategy": "abd/template",
    "template": [
        {
            "artifact": "/var/abd/<identifier>"
        }
    ]
}
$ cat /usr/lib/abd/sources.list.d/zz-default.conf
{
    "prefix": "*",
    "strategy": "abd/https-dns"
}
```

#### Scenario: block all fetching

```
$ cat /usr/lib/abd/sources.list.d/zz-default.conf
{
    "prefix": "*",
    "strategy": "abd/noop"
}
```

[apt-sources-list]: http://manpages.debian.org/cgi-bin/man.cgi?sektion=5&query=sources.list&apropos=0&manpath=sid&locale=en

## Applications Utilizing ABD

Application developers that wish to utilize ABD can shell out to the abd
binary, import the golang library `github.com/appc/abd/abd`, or implement the
semantics defined here themselves.

### Default Labels

It is suggested (but not required) that regardless of labels specified by the
user, the given application using ABD will add additional labels to filter down
for artifacts designed for that application. For example when fetching App
Container Images with rkt, the label `content-type=aci` is added.

### Public Keys

ABD has a field to provide a URI to each artifact's signature, however there is
no prescribed way to discover the signing key for a given artifact. It is on
the onus of the application using ABD to acquire and correlate public keys to
ABD artifacts.

Fortunately, ABD can be used to find a public key for an ABD artifact. One
potential scheme to accomplish this would be to simply re-request the
identifier with the previous set of labels, except replacing the `content-type`
label with `gpg-public-key`.

## Addendum: ABD and AppC

If ABD is implemented (as an independent specification), the authors of AppC
propose that it replace the existing discovery section of AppC. 

There are a number of outstanding issues/questions around AppC discovery that
this should resolve:
- Most notably, discovery should be reframed as a selectable set of discovery
  strategies.
- Use TXT DNS records for discovery (this could be implemented as a plugin).
- Where do arch/os/version labels come from (these would no longer be special
  labels, and simple discovery would be scrapped).
- Adopting TUF (part of this proposal).
- Supporting S3-backed repositories (should be easy to implement with new
  framework).
