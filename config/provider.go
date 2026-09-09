package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	// xpshim is injected into the vendored Terraform provider source by
	// `make provider-source`; see hack/xpshim.go.tmpl.
	"github.com/sky-cloak/terraform-provider-skycloak/xpshim"
)

const (
	resourcePrefix = "skycloak"
	modulePath     = "github.com/decade-eng/provider-skycloak"

	// tfProviderVersion is reported in the Terraform provider's user agent.
	// Keep in sync with TERRAFORM_PROVIDER_VERSION in the Makefile.
	tfProviderVersion = "0.5.1"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("skycloak.crossplane.io"),
		ujconfig.WithIncludeList(nil),
		ujconfig.WithTerraformPluginFrameworkIncludeList(ExternalNameConfigured()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithTerraformPluginFrameworkProvider(xpshim.NewProvider(tfProviderVersion)),
		ujconfig.WithDefaultResourceOptions(
			GroupKindOverrides(),
			ExternalNameConfigurations(),
		))

	Configure(pc)

	pc.ConfigureResources()
	return pc
}

// GetProviderNamespaced returns the namespaced provider configuration
func GetProviderNamespaced() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("skycloak.m.crossplane.io"),
		ujconfig.WithIncludeList(nil),
		ujconfig.WithTerraformPluginFrameworkIncludeList(ExternalNameConfigured()),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithTerraformPluginFrameworkProvider(xpshim.NewProvider(tfProviderVersion)),
		ujconfig.WithDefaultResourceOptions(
			GroupKindOverrides(),
			ExternalNameConfigurations(),
		),
		ujconfig.WithExampleManifestConfiguration(ujconfig.ExampleManifestConfiguration{
			ManagedResourceNamespace: "crossplane-system",
		}))

	Configure(pc)

	pc.ConfigureResources()
	return pc
}
