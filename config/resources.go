package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	skycloakCluster = "skycloak_cluster"
	skycloakDomain  = "skycloak_domain"
)

// gvkOverrides maps Terraform resource names to the API group and kind of the
// generated managed resources. Explicit for every resource so the generated
// API surface is deterministic and independent of upjet's name-splitting
// heuristics. All resources live in the root group (empty ShortGroup);
// only the Kind is pinned.
var gvkOverrides = map[string]schema.GroupVersionKind{
	"skycloak_application":                 {Kind: "Application"},
	"skycloak_application_role_assignment": {Kind: "ApplicationRoleAssignment"},
	"skycloak_application_secret":          {Kind: "ApplicationSecret"},
	"skycloak_captcha_domain":              {Kind: "CaptchaDomain"},
	"skycloak_client_theme_assignment":     {Kind: "ClientThemeAssignment"},
	"skycloak_cluster":                     {Kind: "Cluster"},
	"skycloak_cluster_extension":           {Kind: "ClusterExtension"},
	"skycloak_cluster_maintenance_window":  {Kind: "ClusterMaintenanceWindow"},
	"skycloak_cluster_security":            {Kind: "ClusterSecurity"},
	"skycloak_custom_extension":            {Kind: "CustomExtension"},
	"skycloak_custom_theme":                {Kind: "CustomTheme"},
	"skycloak_domain":                      {Kind: "Domain"},
	"skycloak_domain_route":                {Kind: "DomainRoute"},
	"skycloak_email_branding":              {Kind: "EmailBranding"},
	"skycloak_export":                      {Kind: "Export"},
	"skycloak_identity_provider":           {Kind: "IdentityProvider"},
	"skycloak_login_branding":              {Kind: "LoginBranding"},
	"skycloak_realm":                       {Kind: "Realm"},
	"skycloak_realm_export":                {Kind: "RealmExport"},
	"skycloak_realm_group":                 {Kind: "RealmGroup"},
	"skycloak_realm_group_membership":      {Kind: "RealmGroupMembership"},
	"skycloak_realm_import":                {Kind: "RealmImport"},
	"skycloak_realm_role":                  {Kind: "RealmRole"},
	"skycloak_realm_user":                  {Kind: "RealmUser"},
	"skycloak_realm_user_role_assignment":  {Kind: "RealmUserRoleAssignment"},
	"skycloak_siem_destination":            {Kind: "SIEMDestination"},
	"skycloak_smtp":                        {Kind: "SMTP"},
	"skycloak_theme_assignment":            {Kind: "ThemeAssignment"},
	"skycloak_webhook_subscription":        {Kind: "WebhookSubscription"},
}

// GroupKindOverrides applies the group/kind overrides from gvkOverrides.
func GroupKindOverrides() config.ResourceOption {
	return func(r *config.Resource) {
		if gvk, ok := gvkOverrides[r.Name]; ok {
			r.ShortGroup = gvk.Group
			if gvk.Kind != "" {
				r.Kind = gvk.Kind
			}
		}
	}
}

// clusterScopedResources are resources that belong to a cluster and reference
// it through a cluster_id parameter.
var clusterScopedResources = []string{
	"skycloak_application",
	"skycloak_application_role_assignment",
	"skycloak_application_secret",
	"skycloak_captcha_domain",
	"skycloak_client_theme_assignment",
	"skycloak_cluster_extension",
	"skycloak_cluster_maintenance_window",
	"skycloak_cluster_security",
	"skycloak_custom_theme",
	"skycloak_domain",
	"skycloak_domain_route",
	"skycloak_email_branding",
	"skycloak_export",
	"skycloak_identity_provider",
	"skycloak_login_branding",
	"skycloak_realm",
	"skycloak_realm_export",
	"skycloak_realm_group",
	"skycloak_realm_group_membership",
	"skycloak_realm_import",
	"skycloak_realm_role",
	"skycloak_realm_user",
	"skycloak_realm_user_role_assignment",
	"skycloak_smtp",
	"skycloak_theme_assignment",
}

// Configure adds resource-specific configurations.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator(skycloakCluster, func(r *config.Resource) {
		// Cluster provisioning is a long-running operation; reconcile it
		// asynchronously.
		r.UseAsync = true
	})

	for _, name := range clusterScopedResources {
		p.AddResourceConfigurator(name, func(r *config.Resource) {
			r.References = config.References{
				"cluster_id": {TerraformName: skycloakCluster},
			}
		})
	}

	p.AddResourceConfigurator("skycloak_domain_route", func(r *config.Resource) {
		r.References["domain_id"] = config.Reference{TerraformName: skycloakDomain}
	})

	// realm_name is a plain realm name (not an ID) on the resources that carry
	// it, so a typed reference to the Realm MR (whose external-name is a
	// provider UUID) would be wrong; cluster_id remains the only cross-ref.

	// skycloak_realm_import can take a source_export_id produced by
	// skycloak_realm_export.
	p.AddResourceConfigurator("skycloak_realm_import", func(r *config.Resource) {
		r.References["source_export_id"] = config.Reference{TerraformName: "skycloak_realm_export"}
	})
}
