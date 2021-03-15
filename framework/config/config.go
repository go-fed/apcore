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

// Overall configuration file structure
type Config struct {
	ServerConfig      ServerConfig      `ini:"server" comment:"HTTP server configuration"`
	OAuthConfig       OAuth2Config      `ini:"oauth" comment:"OAuth 2 configuration"`
	DatabaseConfig    DatabaseConfig    `ini:"database" comment:"Database configuration"`
	ActivityPubConfig ActivityPubConfig `ini:"activitypub" comment:"ActivityPub configuration"`
	NodeInfoConfig    NodeInfoConfig    `ini:"nodeinfo" comment:"NodeInfo configuration"`
}

// Configuration section specifically for the HTTP server.
type ServerConfig struct {
	Host                        string `ini:"sr_host" comment:"(required) Host with TLD for this instance (basically, the fully qualified domain or subdomain); ignored in debug mode"`
	CertFile                    string `ini:"sr_cert_file" comment:"(required) Path to the certificate file used to establish TLS connections for HTTPS"`
	KeyFile                     string `ini:"sr_key_file" comment:"(required) Path to the private key file used to establish TLS connections for HTTPS"`
	Proxy                       bool   `ini:"sr_proxy" comment:"If you run the application server behind a proxy, this will disable TLS server."`
	CookieAuthKeyFile           string `ini:"sr_cookie_auth_key_file" comment:"(required) Path to private key file used for cookie authentication"`
	CookieEncryptionKeyFile     string `ini:"sr_cookie_encryption_key_file" comment:"Path to private key file used for cookie encryption"`
	CookieMaxAge                int    `ini:"sr_cookie_max_age" comment:"(default: 86400 seconds) Number of seconds a cookie is valid; 0 indicates no Max-Age (browser-dependent, usually session-only); negative value is invalid"`
	CookieSessionName           string `ini:"sr_cookie_session_name" comment:"(required) Cookie session name to use for the application"`
	HttpsReadTimeoutSeconds     int    `ini:"sr_https_read_timeout_seconds" comment:"Timeout in seconds for incoming HTTPS requests; a zero or unset value does not timeout"`
	HttpsWriteTimeoutSeconds    int    `ini:"sr_https_write_timeout_seconds" comment:"Timeout in seconds for outgoing HTTPS responses; a zero or unset value does not timeout"`
	HttpClientTimeoutSeconds    int    `ini:"sr_http_client_timeout_seconds" comment:"Timeout in seconds for outgoing HTTP requests; a zero or unset value does not timeout"`
	RedirectReadTimeoutSeconds  int    `ini:"sr_redirect_read_timeout_seconds" comment:"Timeout in seconds for incoming HTTP requests, which will be redirected to HTTPS; a zero or unset value does not timeout"`
	RedirectWriteTimeoutSeconds int    `ini:"sr_redirect_write_timeout_seconds" comment:"Timeout in seconds for outgoing HTTP redirect-to-HTTPS responses; a zero or unset value does not timeout"`
	StaticRootDirectory         string `ini:"sr_static_root_directory" comment:"(required) Root directory for serving static content, such as ECMAScript, CSS, favicon; !!!Warning: Everything in this directory will be served and accessible!!!"`
	SaltSize                    int    `ini:"sr_salt_size" comment:"(default: 32) The size of salts to use with passwords when hashing, anything smaller than 16 will be treated as 16"`
	BCryptStrength              int    `ini:"sr_bcrypt_strength" comment:"(default: 10) The hashing cost to use with the bcrypt hashing algorithm, between 4 and 31; the higher the cost, the slower the hash comparisons for passwords will take for attackers and regular users alike"`
	RSAKeySize                  int    `ini:"sr_rsa_private_key_size" comment:"(default: 1024) The size of the RSA private key for a user; values less than 1024 are forbidden"`
}

type OAuth2Config struct {
	AccessTokenExpiry  int `ini:"oauth_access_token_expiry" comment:"(default: 3600 seconds) Duration in seconds until an access token expires; zero or negative values are invalid."`
	RefreshTokenExpiry int `ini:"oauth_refresh_token_expiry" comment:"(default: 7200 seconds) Duration in seconds until a refresh token expires; zero or negative values are invalid."`
}

// Configuration section specifically for the database.
type DatabaseConfig struct {
	DatabaseKind              string         `ini:"db_database_kind" comment:"(required) Only \"postgres\" supported"`
	ConnMaxLifetimeSeconds    int            `ini:"db_conn_max_lifetime_seconds" comment:"(default: indefinite) Maximum lifetime of a connection in seconds; a value of zero or unset value means indefinite"`
	MaxOpenConns              int            `ini:"db_max_open_conns" comment:"(default: infinite) Maximum number of open connections to the database; a value of zero or unset value means infinite"`
	MaxIdleConns              int            `ini:"db_max_idle_conns" comment:"(default: 2) Maximum number of idle connections in the connection pool to the database; a value of zero maintains no idle connections; a value greater than max_open_conns is reduced to be equal to max_open_conns"`
	DefaultCollectionPageSize int            `ini:"db_default_collection_page_size" comment:"(default: 10) The default collection page size when fetching a page of an ActivityStreams collection"`
	MaxCollectionPageSize     int            `ini:"db_max_collection_page_size" comment:"(default: 200) The maximum collection page size allowed when fetching a page of an ActivityStreams collection"`
	PostgresConfig            PostgresConfig `ini:"db_postgres,omitempty" comment:"Only needed if database_kind is postgres, and values are based on the github.com/jackc/pgx driver"`
}

// Configuration section specifically for ActivityPub.
type ActivityPubConfig struct {
	ClockTimezone                       string               `ini:"ap_clock_timezone" comment:"(default: UTC) Timezone for ActivityPub related operations: unset and \"UTC\" are UTC, \"Local\" is local server time, otherwise use IANA Time Zone database values"`
	OutboundRateLimitQPS                float64              `ini:"ap_outbound_rate_limit_qps" comment:"(default: 2) Per-host outbound rate limit for delivery of federated messages under steady state conditions; a negative value or value of zero is invalid"`
	OutboundRateLimitBurst              int                  `ini:"ap_outbound_rate_limit_burst" comment:"(default: 5) Per-host outbound burst tolerance for delivery of federated messages; a negative value or value of zero is invalid"`
	OutboundRateLimitPrunePeriodSeconds int                  `ini:"ap_outbound_rate_limit_prune_period_seconds" comment:"(default: 60) The time period to await before periodically removing cached per-host rate-limiters that are no longer in use, controlling how frequently pruning occurs; a negative value or value of zero is invalid"`
	OutboundRateLimitPruneAgeSeconds    int                  `ini:"ap_outbound_rate_limit_prune_age_seconds" comment:"(default: 30) The age of an unused per-host rate-limiter must be to be pruned and removed from the cache when the pruning occurs, controlling how long cached rate-limiters are kept when unused; a negative value is invalid"`
	HttpSignaturesConfig                HttpSignaturesConfig `ini:"ap_http_signatures" comment:"HTTP Signatures configuration"`
	MaxInboxForwardingRecursionDepth    int                  `ini:"ap_max_inbox_forwarding_recursion_depth" comment:"(default: 50) The maximum recursion depth to use when determining whether to do inbox forwarding, which if triggered ensures older thread participants are able to receive messages; zero means no limit (only used if the application has S2S enabled); a negative value is invalid"`
	MaxDeliveryRecursionDepth           int                  `ini:"ap_max_delivery_recursion_depth" comment:"(default: 50) The maximum depth to search for peers to deliver due to inbox forwarding, which ensures messages received by this server are propagated to them and no \"ghost reply\" problems occur; zero means no limit (only used if the application has S2S enabled); a negative value is invalid"`
	RetryPageSize                       int                  `ini:"ap_retry_page_size" comment:"(default: 25) The number of retryable deliveries to request from the database at a time; a negative value or zero value is invalid"`
	RetryAbandonLimit                   int                  `ini:"ap_retry_abandon_limit" comment:"(default: 10) The maximum number of times the app will attempt to deliver an Activity to a federated peer and fail before permanently giving up and abandoning any further attempts to deliver it; a negative value or zero value is invalid"`
	RetrySleepPeriod                    int                  `ini:"ap_retry_sleep_period_seconds" comment:"(default: 300) The time period to await between making periodic attempts to re-deliver Activities to federated peers that have never been successfully delivered; a 300-second retry sleep period with an abandon limit of 10 results in an exponential backoff of 10 delivery attempts across roughly 3 days; a negative value or zero value is invalid"`
}

// Configuration for HTTP Signatures.
type HttpSignaturesConfig struct {
	Algorithms      []string `ini:"http_sig_algorithms" comment:"(default: \"rsa-sha256,rsa-sha512\") Comma-separated list of algorithms used by the go-fed/httpsig library to sign outgoing HTTP signatures; the first algorithm in this list will be the one used to verify other peers' HTTP signatures"`
	DigestAlgorithm string   `ini:"http_sig_digest_algorithm" comment:"(default: \"SHA-256\") RFC 3230 algorithm for use in signing header Digests"`
	GetHeaders      []string `ini:"http_sig_get_headers" comment:"(default: \"(request-target),Date\") Comma-separated list of HTTP headers to sign in GET requests; must contain \"(request-target)\" and \"Date\""`
	PostHeaders     []string `ini:"http_sig_post_headers" comment:"(default: \"(request-target),Date,Digest\") Comma-separated list of HTTP headers to sign in POST requests; must contain \"(request-target)\", \"Date\", and \"Digest\""`
}

// Configuration section specifically for Postgres databases.
type PostgresConfig struct {
	DatabaseName            string `ini:"pg_db_name" comment:"(required) Database name"`
	UserName                string `ini:"pg_user" comment:"(required) User to connect as (any password will be prompted)"`
	Host                    string `ini:"pg_host" comment:"(default: localhost) The Postgres host to connect to"`
	Port                    int    `ini:"pg_port" comment:"(default: 5432) The port to connect to"`
	Password                string `ini:"password" comment:"The database password to use to connect"`
	SSLMode                 string `ini:"pg_ssl_mode" comment:"(default: require) SSL mode to use when connecting (options are: \"disable\", \"require\", \"verify-ca\", \"verify-full\")"`
	FallbackApplicationName string `ini:"pg_fallback_application_name" comment:"An application_name to fall back to if one is not provided"`
	ConnectTimeout          int    `ini:"pg_connect_timeout" comment:"(default: indefinite) Maximum wait when connecting to a database, zero or unset means indefinite"`
	SSLCert                 string `ini:"pg_ssl_cert" comment:"PEM-encoded certificate file location"`
	SSLKey                  string `ini:"pg_ssl_key" comment:"PEM-encoded private key file location"`
	SSLRootCert             string `ini:"pg_ssl_root_cert" comment:"PEM-encoded root certificate file location"`
	Schema                  string `ini:"pg_schema" comment:"Postgres schema prefix to use"`
}

// Configuration section specifically for NodeInfo.
type NodeInfoConfig struct {
	EnableNodeInfo                         bool `ini:"ni_enable_nodeinfo" comment:"(default: true) Whether to share basic server and software information at a somewhat-Fediverse-understood endpoint for public use; NodeInfo is upstream of the NodeInfo2 fork and in general admins will either wish to enable or disable both"`
	EnableNodeInfo2                        bool `ini:"ni_enable_nodeinfo2" comment:"(default: true) Whether to share basic server, organization, and software information at a somewhat-Fediverse-understood endpoint for public use; NodeInfo2 is a fork of NodeInfo and in general admins will either wish to enable or disable both"`
	EnableAnonymousStatsSharing            bool `ini:"ni_enable_anon_stats_sharing" comment:"(default: true) Whether to share anonymized statistics about user counts, counts of user activity over various periods of time, local post counts, and local comment counts to the public; for sufficiently small instances the statistics are always shared with noise introduced; if none of the NodeInfos are enabled then this option does nothing"`
	AnonymizedStatsCacheInvalidatedSeconds int  `ini:"ni_anon_stats_cache_invalidated_seconds" comment:"(default: 86400) The number of seconds before the anonymized node statistics are refreshed and updated; in the meantime the existing values will be cached and served for this period of time"`
}
