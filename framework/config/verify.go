// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2019 Cory Slep
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"errors"
	"fmt"
)

func (c *Config) Verify() error {
	if err := c.ServerConfig.Verify(); err != nil {
		return err
	}
	if err := c.OAuthConfig.Verify(); err != nil {
		return err
	}
	if err := c.DatabaseConfig.Verify(); err != nil {
		return err
	}
	if err := c.ActivityPubConfig.Verify(); err != nil {
		return err
	}
	return nil
}

func (c *ServerConfig) Verify() error {
	if len(c.Host) == 0 {
		return errors.New("sr_host is empty, but it is required")
	}
	if len(c.CertFile) == 0 {
		return errors.New("sr_cert_file is empty, but it is required")
	}
	if len(c.KeyFile) == 0 {
		return errors.New("sr_key_file is empty, but it is required")
	}
	if len(c.CookieAuthKeyFile) == 0 {
		return errors.New("sr_cookie_auth_key_file is empty, but it is required")
	}
	if len(c.CookieSessionName) == 0 {
		return errors.New("sr_cookie_session_name is empty, but it is required")
	}
	if len(c.StaticRootDirectory) == 0 {
		return errors.New("sr_static_root_directory is empty, but it is required")
	}
	const minKeySize = 1024
	if c.RSAKeySize < minKeySize {
		return fmt.Errorf("sr_rsa_private_key_size is configured to be < %d, which is forbidden: %d", minKeySize, c.RSAKeySize)
	}
	return nil
}

func (c *OAuth2Config) Verify() error {
	if c.AccessTokenExpiry <= 0 {
		return fmt.Errorf("oauth_access_token_expiry is zero or negative, which is forbidden: %d", c.AccessTokenExpiry)
	}
	if c.RefreshTokenExpiry <= 0 {
		return fmt.Errorf("oauth_refresh_token_expiry is zero or negative, which is forbidden: %d", c.RefreshTokenExpiry)
	}
	return nil
}

func (c *DatabaseConfig) Verify() error {
	if len(c.DatabaseKind) == 0 {
		return errors.New("db_database_kind is empty, but it is required")
	}
	if c.DatabaseKind == "postgres" {
		if err := c.PostgresConfig.Verify(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ActivityPubConfig) Verify() error {
	if c.OutboundRateLimitQPS <= 0 {
		return fmt.Errorf("ap_outbound_rate_limit_qps is zero or negative, which is forbidden: %d", c.OutboundRateLimitQPS)
	}
	if c.OutboundRateLimitBurst <= 0 {
		return fmt.Errorf("ap_outbound_rate_limit_burst is zero or negative, which is forbidden: %d", c.OutboundRateLimitBurst)
	}
	if c.OutboundRateLimitPrunePeriodSeconds <= 0 {
		return fmt.Errorf("ap_outbound_rate_limit_prune_period_seconds is zero or negative, which is forbidden: %d", c.OutboundRateLimitPrunePeriodSeconds)
	}
	if c.OutboundRateLimitPruneAgeSeconds < 0 {
		return fmt.Errorf("ap_outbound_rate_limit_prune_age_seconds is negative, which is forbidden: %d", c.OutboundRateLimitPruneAgeSeconds)
	}
	if c.MaxInboxForwardingRecursionDepth < 0 {
		return fmt.Errorf("ap_max_inbox_forwarding_recursion_depth is negative, which is forbidden: %d", c.MaxInboxForwardingRecursionDepth)
	}
	if c.MaxDeliveryRecursionDepth < 0 {
		return fmt.Errorf("ap_max_delivery_recursion_depth is negative, which is forbidden: %d", c.MaxDeliveryRecursionDepth)
	}
	if c.RetryPageSize <= 0 {
		return fmt.Errorf("ap_retry_page_size is zero or negative, which is forbidden: %d", c.RetryPageSize)
	}
	if c.RetryAbandonLimit <= 0 {
		return fmt.Errorf("ap_retry_abandon_limit is zero or negative, which is forbidden: %d", c.RetryAbandonLimit)
	}
	if c.RetrySleepPeriod <= 0 {
		return fmt.Errorf("ap_retry_sleep_period_seconds is zero or negative, which is forbidden: %d", c.RetrySleepPeriod)
	}
	if err := c.HttpSignaturesConfig.Verify(); err != nil {
		return err
	}
	return nil
}

func (c *HttpSignaturesConfig) Verify() error {
	return nil
}

func (c *PostgresConfig) Verify() error {
	return nil
}
