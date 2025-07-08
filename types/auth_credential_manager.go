// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"slices"
)

// CredentialManager manages authentication credentials through a structured workflow.
//
// The CredentialManager orchestrates the complete lifecycle of authentication
// credentials, from initial loading to final preparation for use. It provides
// a centralized interface for handling various credential types and authentication
// schemes while maintaining proper credential hygiene (refresh, exchange, caching).
//
// This class is only for use by Agent Development Kit.
//
// # Experimental
//
// This feature is experimental and may change or be removed in future versions without notice. It may
// introduce breaking changes at any time.
type CredentialManager struct {
	authConfig        *AuthConfig
	exchangerRegistry *CredentialExchangerRegistry
	refresherRegistry *CredentialRefresherRegistry
}

// NewCredentialManager creates a new CredentialManager with the given AuthConfig.
func NewCredentialManager(authConfig *AuthConfig) *CredentialManager {
	cm := &CredentialManager{
		authConfig:        authConfig,
		exchangerRegistry: NewCredentialExchangerRegistry(),
		refresherRegistry: NewCredentialRefresherRegistry(),
	}

	// Register default exchangers and refreshers
	// TODO(adk-python): support service account credential exchanger
	oauth2Refresher := &OAuth2CredentialRefresher{}
	cm.refresherRegistry.Register(OAuth2CredentialTypes, oauth2Refresher)
	cm.refresherRegistry.Register(OpenIDConnectCredentialTypes, oauth2Refresher)

	return cm
}

// RegisterCredentialExchanger register a credential exchanger for a credential type.
func (cm *CredentialManager) RegisterCredentialExchanger(ctx context.Context, credentialType AuthCredentialTypes, exchanger CredentialExchanger) {
	cm.exchangerRegistry.Register(credentialType, exchanger)
}

// RequestCredential requests a credential for the given [ToolContext].
func (cm *CredentialManager) RequestCredential(ctx context.Context, toolCtx *ToolContext) {
	toolCtx.RequestCredential(cm.authConfig)
}

// GetAuthCredential load and prepare authentication credential through a structured workflow.
func (cm *CredentialManager) GetAuthCredential(ctx context.Context, toolCtx *ToolContext) (*AuthCredential, error) {
	// Step 1: Validate credential configuration
	if err := cm.validateCredential(); err != nil {
		return nil, err
	}

	// Step 2: Check if credential is already ready (no processing needed)
	if cm.isCredentialReady() {
		return cm.authConfig.RawAuthCredential, nil
	}

	// Step 3: Try to load existing processed credential
	credential, err := cm.loadExistingCredential(ctx, toolCtx)
	if err != nil {
		return nil, err
	}

	// Step 4: If no existing credential, load from auth response
	// TODO(adk-python): instead of load from auth response, we can store auth response in
	// credential service.
	wasFromAuthResponse := false
	if credential == nil {
		credential = cm.loadFromAuthResponse(ctx, toolCtx)
		if credential != nil {
			wasFromAuthResponse = true
		}
	}

	// Step 5: If still no credential available, return None
	if credential == nil {
		return nil, nil
	}

	// Step 6: Exchange credential if needed (e.g., service account to access token)
	var wasExchanged bool
	credential, wasExchanged = cm.exchangeCredential(ctx, credential)

	// Step 7: Refresh credential if expired
	var wasRefreshed bool
	if !wasExchanged {
		credential, wasRefreshed = cm.refreshCredential(ctx, credential)
	}

	// Step 8: Save credential if it was modified
	if wasFromAuthResponse || wasExchanged || wasRefreshed {
		cm.saveCredential(ctx, toolCtx, credential)
	}

	return credential, nil
}

// loadExistingCredential load existing credential from credential service or cached exchanged credential.
func (cm *CredentialManager) loadExistingCredential(ctx context.Context, toolCtx *ToolContext) (*AuthCredential, error) {
	// Try loading from credential service first
	credentials, err := cm.loadFromCredentialService(ctx, toolCtx)
	if err != nil {
		return nil, err
	}
	if credentials != nil {
		return credentials, err
	}

	// Check if we have a cached exchanged credential
	if cm.authConfig.ExchangedAuthCredential != nil {
		return cm.authConfig.ExchangedAuthCredential, err
	}

	return nil, nil
}

// loadFromCredentialService load credential from credential service if available.
func (cm *CredentialManager) loadFromCredentialService(ctx context.Context, toolCtx *ToolContext) (*AuthCredential, error) {
	credentialService := toolCtx.InvocationContext().CredentialService
	if credentialService != nil {
		// NOTE(adk-python): This should be made async in a future refactor
		// For now, assuming synchronous operation
		return credentialService.LoadCredential(ctx, cm.authConfig, toolCtx)
	}

	return nil, nil
}

// loadFromAuthResponse load credential from auth response in tool context.
func (cm *CredentialManager) loadFromAuthResponse(ctx context.Context, toolCtx *ToolContext) *AuthCredential {
	return toolCtx.GetAuthResponse(cm.authConfig)
}

// exchangeCredential exchange credential if needed and return the credential and whether it was exchanged.
func (cm *CredentialManager) exchangeCredential(ctx context.Context, credential *AuthCredential) (*AuthCredential, bool) {
	exchanger, ok := cm.exchangerRegistry.GetExchanger(credential.AuthType)
	if !ok {
		return credential, false
	}

	exchangedCredential, err := exchanger.Exchange(ctx, credential, cm.authConfig.AuthScheme)
	if err != nil {
		return credential, false
	}

	return exchangedCredential, true
}

// refreshCredential refresh credential if expired and return the credential and whether it was refreshed.
func (cm *CredentialManager) refreshCredential(ctx context.Context, credential *AuthCredential) (*AuthCredential, bool) {
	refresher, ok := cm.refresherRegistry.GetRefresher(credential.AuthType)
	if !ok {
		return credential, false
	}

	if refresher.IsRefreshNeeded(ctx, credential, cm.authConfig.AuthScheme) {
		refreshedCredential, err := refresher.Refresh(ctx, credential, cm.authConfig.AuthScheme)
		if err != nil {
			return credential, false
		}
		return refreshedCredential, true
	}

	return credential, false
}

// isCredentialReady check if credential is ready to use without further processing.
func (cm *CredentialManager) isCredentialReady() bool {
	rawCredential := cm.authConfig.RawAuthCredential
	if rawCredential == nil {
		return false
	}

	// Simple credentials that don't need exchange or refresh
	authTypes := []AuthCredentialTypes{
		APIKeyCredentialTypes,
		HTTPCredentialTypes,
	}
	return slices.Contains(authTypes, rawCredential.AuthType)
}

// validateCredential validate credential configuration and raise errors if invalid.
//
// TODO(zchee): implements
func (cm *CredentialManager) validateCredential() error {
	return nil
}

// saveCredential save credential to credential service if available.
func (cm *CredentialManager) saveCredential(ctx context.Context, toolCtx *ToolContext, credential *AuthCredential) {
	credentialService := toolCtx.InvocationContext().CredentialService
	if credentialService != nil {
		// Update the exchanged credential in config
		cm.authConfig.ExchangedAuthCredential = credential
		credentialService.SaveCredential(ctx, cm.authConfig, toolCtx)
	}
}
