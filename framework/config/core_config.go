// apcore is a server framework for implementing an ActivityPub application.
// Copyright (C) 2021 Cory Slep
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

func (c *Config) Host() string {
	return c.ServerConfig.Host
}

func (c *Config) ClockTimezone() string {
	return c.ActivityPubConfig.ClockTimezone
}

func (c *Config) Schema() string {
	return c.DatabaseConfig.PostgresConfig.Schema
}
