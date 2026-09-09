/*
Copyright 2026 Decade Engineering
*/

// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	providerconfig "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/providerconfig"
	application "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/application"
	applicationroleassignment "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/applicationroleassignment"
	applicationsecret "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/applicationsecret"
	captchadomain "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/captchadomain"
	clientthemeassignment "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/clientthemeassignment"
	cluster "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/cluster"
	clusterextension "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/clusterextension"
	clustermaintenancewindow "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/clustermaintenancewindow"
	clustersecurity "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/clustersecurity"
	customextension "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/customextension"
	customtheme "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/customtheme"
	domain "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/domain"
	domainroute "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/domainroute"
	emailbranding "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/emailbranding"
	export "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/export"
	identityprovider "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/identityprovider"
	loginbranding "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/loginbranding"
	realm "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/realm"
	realmexport "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/realmexport"
	realmgroup "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/realmgroup"
	realmgroupmembership "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/realmgroupmembership"
	realmimport "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/realmimport"
	realmrole "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/realmrole"
	realmuser "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/realmuser"
	realmuserroleassignment "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/realmuserroleassignment"
	siemdestination "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/siemdestination"
	smtp "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/smtp"
	themeassignment "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/themeassignment"
	webhooksubscription "github.com/decade-eng/provider-skycloak/internal/controller/namespaced/skycloak/webhooksubscription"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		providerconfig.Setup,
		application.Setup,
		applicationroleassignment.Setup,
		applicationsecret.Setup,
		captchadomain.Setup,
		clientthemeassignment.Setup,
		cluster.Setup,
		clusterextension.Setup,
		clustermaintenancewindow.Setup,
		clustersecurity.Setup,
		customextension.Setup,
		customtheme.Setup,
		domain.Setup,
		domainroute.Setup,
		emailbranding.Setup,
		export.Setup,
		identityprovider.Setup,
		loginbranding.Setup,
		realm.Setup,
		realmexport.Setup,
		realmgroup.Setup,
		realmgroupmembership.Setup,
		realmimport.Setup,
		realmrole.Setup,
		realmuser.Setup,
		realmuserroleassignment.Setup,
		siemdestination.Setup,
		smtp.Setup,
		themeassignment.Setup,
		webhooksubscription.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		providerconfig.SetupGated,
		application.SetupGated,
		applicationroleassignment.SetupGated,
		applicationsecret.SetupGated,
		captchadomain.SetupGated,
		clientthemeassignment.SetupGated,
		cluster.SetupGated,
		clusterextension.SetupGated,
		clustermaintenancewindow.SetupGated,
		clustersecurity.SetupGated,
		customextension.SetupGated,
		customtheme.SetupGated,
		domain.SetupGated,
		domainroute.SetupGated,
		emailbranding.SetupGated,
		export.SetupGated,
		identityprovider.SetupGated,
		loginbranding.SetupGated,
		realm.SetupGated,
		realmexport.SetupGated,
		realmgroup.SetupGated,
		realmgroupmembership.SetupGated,
		realmimport.SetupGated,
		realmrole.SetupGated,
		realmuser.SetupGated,
		realmuserroleassignment.SetupGated,
		siemdestination.SetupGated,
		smtp.SetupGated,
		themeassignment.SetupGated,
		webhooksubscription.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		providerconfig.SetupWebhookWithManager,
		application.SetupWebhookWithManager,
		applicationroleassignment.SetupWebhookWithManager,
		applicationsecret.SetupWebhookWithManager,
		captchadomain.SetupWebhookWithManager,
		clientthemeassignment.SetupWebhookWithManager,
		cluster.SetupWebhookWithManager,
		clusterextension.SetupWebhookWithManager,
		clustermaintenancewindow.SetupWebhookWithManager,
		clustersecurity.SetupWebhookWithManager,
		customextension.SetupWebhookWithManager,
		customtheme.SetupWebhookWithManager,
		domain.SetupWebhookWithManager,
		domainroute.SetupWebhookWithManager,
		emailbranding.SetupWebhookWithManager,
		export.SetupWebhookWithManager,
		identityprovider.SetupWebhookWithManager,
		loginbranding.SetupWebhookWithManager,
		realm.SetupWebhookWithManager,
		realmexport.SetupWebhookWithManager,
		realmgroup.SetupWebhookWithManager,
		realmgroupmembership.SetupWebhookWithManager,
		realmimport.SetupWebhookWithManager,
		realmrole.SetupWebhookWithManager,
		realmuser.SetupWebhookWithManager,
		realmuserroleassignment.SetupWebhookWithManager,
		siemdestination.SetupWebhookWithManager,
		smtp.SetupWebhookWithManager,
		themeassignment.SetupWebhookWithManager,
		webhooksubscription.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
