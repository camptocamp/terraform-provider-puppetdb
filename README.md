PuppetDB Terraform Provider
===========================

This is a Terraform provider to interact with the [PuppetDB](https://puppet.com/docs/puppetdb/latest/index.html). It allows to verify that a node was properly registered in a PuppetDB and to clean it upon decommissing the node.


Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.8 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/camptocamp/terraform-provider-puppetdb`

```sh
$ mkdir -p $GOPATH/src/github.com/camptocamp; cd $GOPATH/src/github.com/camptocamp
$ git clone git@github.com:camptocamp/terraform-provider-puppetdb
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/camptocamp/terraform-provider-puppetdb
$ make build
```

Using the provider
----------------------

```hcl
provider puppetdb {
  url = "https://puppetdb:8081"
  cert = "certs/puppetdb.crt"
  key = "certs/puppetdb.key"
  ca = "certs/ca.pem"
  
}

resource puppetdb_node "foo" {
   certname = "foo.example.com"
}
```


Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-puppetdb
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
