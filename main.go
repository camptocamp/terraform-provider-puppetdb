package main

/* Bootstrap the plugin for PuppetDB */

import (
	"github.com/camptocamp/terraform-provider-puppetdb/puppetdb"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: puppetdb.Provider,
	})
}
