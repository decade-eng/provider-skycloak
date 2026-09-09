/*
Copyright 2026 Decade Engineering
*/

package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// ExternalNameConfigs contains all external name configurations for this
// provider. Every Skycloak resource is identified by a provider-assigned ID
// (a UUID, or a composite such as "cluster_id/extension_id"), so all of them
// use IdentifierFromProvider.
var ExternalNameConfigs = map[string]config.ExternalName{
	"skycloak_application":                 config.IdentifierFromProvider,
	"skycloak_application_role_assignment": config.IdentifierFromProvider,
	"skycloak_application_secret":          config.IdentifierFromProvider,
	"skycloak_captcha_domain":              config.IdentifierFromProvider,
	"skycloak_client_theme_assignment":     config.IdentifierFromProvider,
	"skycloak_cluster":                     config.IdentifierFromProvider,
	"skycloak_cluster_extension":           config.IdentifierFromProvider,
	"skycloak_cluster_maintenance_window":  config.IdentifierFromProvider,
	"skycloak_cluster_security":            config.IdentifierFromProvider,
	"skycloak_custom_extension":            config.IdentifierFromProvider,
	"skycloak_custom_theme":                config.IdentifierFromProvider,
	"skycloak_domain":                      config.IdentifierFromProvider,
	"skycloak_domain_route":                config.IdentifierFromProvider,
	"skycloak_email_branding":              config.IdentifierFromProvider,
	"skycloak_export":                      config.IdentifierFromProvider,
	"skycloak_identity_provider":           config.IdentifierFromProvider,
	"skycloak_login_branding":              config.IdentifierFromProvider,
	"skycloak_realm":                       config.IdentifierFromProvider,
	"skycloak_realm_export":                config.IdentifierFromProvider,
	"skycloak_realm_group":                 config.IdentifierFromProvider,
	"skycloak_realm_group_membership":      config.IdentifierFromProvider,
	"skycloak_realm_import":                config.IdentifierFromProvider,
	"skycloak_realm_role":                  config.IdentifierFromProvider,
	"skycloak_realm_user":                  config.IdentifierFromProvider,
	"skycloak_realm_user_role_assignment":  config.IdentifierFromProvider,
	"skycloak_siem_destination":            config.IdentifierFromProvider,
	"skycloak_smtp":                        config.IdentifierFromProvider,
	"skycloak_theme_assignment":            config.IdentifierFromProvider,
	"skycloak_webhook_subscription":        config.IdentifierFromProvider,
}

// ExternalNameConfigurations applies all external name configs listed in the
// map above.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured, in include-list regex form.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
