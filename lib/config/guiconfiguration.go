// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
)

type GUIConfiguration struct {
	Enabled                   bool     `xml:"enabled,attr" json:"enabled" default:"true"`
	RawAddress                string   `xml:"address" json:"address" default:"127.0.0.1:8384"`
	RawUnixSocketPermissions  string   `xml:"unixSocketPermissions,omitempty" json:"unixSocketPermissions"`
	User                      string   `xml:"user,omitempty" json:"user"`
	Password                  string   `xml:"password,omitempty" json:"password"`
	AuthMode                  AuthMode `xml:"authMode,omitempty" json:"authMode"`
	RawUseTLS                 bool     `xml:"tls,attr" json:"useTLS"`
	APIKey                    string   `xml:"apikey,omitempty" json:"apiKey"`
	InsecureAdminAccess       bool     `xml:"insecureAdminAccess,omitempty" json:"insecureAdminAccess"`
	Theme                     string   `xml:"theme" json:"theme" default:"default"`
	Debugging                 bool     `xml:"debugging,attr" json:"debugging"`
	InsecureSkipHostCheck     bool     `xml:"insecureSkipHostcheck,omitempty" json:"insecureSkipHostcheck"`
	InsecureAllowFrameLoading bool     `xml:"insecureAllowFrameLoading,omitempty" json:"insecureAllowFrameLoading"`
}

func (c GUIConfiguration) IsAuthEnabled() bool {
	return c.AuthMode == AuthModeLDAP || (len(c.User) > 0 && len(c.Password) > 0)
}

func (c GUIConfiguration) IsOverridden() bool {
	return os.Getenv("STGUIADDRESS") != ""
}

func (c GUIConfiguration) Address() string {
	if override := os.Getenv("STGUIADDRESS"); override != "" {
		// This value may be of the form "scheme://address:port" or just
		// "address:port". We need to chop off the scheme. We try to parse it as
		// an URL if it contains a slash. If that fails, return it as is and let
		// some other error handling handle it.

		if strings.Contains(override, "/") {
			url, err := url.Parse(override)
			if err != nil {
				return override
			}
			if strings.HasPrefix(url.Scheme, "unix") {
				return url.Path
			}
			return url.Host
		}

		return override
	}

	return c.RawAddress
}

func (c GUIConfiguration) UnixSocketPermissions() os.FileMode {
	perm, err := strconv.ParseUint(c.RawUnixSocketPermissions, 8, 32)
	if err != nil {
		// ignore incorrectly formatted permissions
		return 0
	}
	return os.FileMode(perm) & os.ModePerm
}

func (c GUIConfiguration) Network() string {
	if override := os.Getenv("STGUIADDRESS"); strings.Contains(override, "/") {
		url, err := url.Parse(override)
		if err != nil {
			return "tcp"
		}
		if strings.HasPrefix(url.Scheme, "unix") {
			return "unix"
		}
	}
	if strings.HasPrefix(c.RawAddress, "/") {
		return "unix"
	}
	return "tcp"
}

func (c GUIConfiguration) UseTLS() bool {
	if override := os.Getenv("STGUIADDRESS"); override != "" {
		if strings.HasPrefix(override, "http") {
			return strings.HasPrefix(override, "https:")
		}
		if strings.HasPrefix(override, "unix") {
			return strings.HasPrefix(override, "unixs:")
		}
	}
	return c.RawUseTLS
}

func (c GUIConfiguration) URL() string {
	if strings.HasPrefix(c.RawAddress, "/") {
		return "unix://" + c.RawAddress
	}

	u := url.URL{
		Scheme: "http",
		Host:   c.Address(),
		Path:   "/",
	}

	if c.UseTLS() {
		u.Scheme = "https"
	}

	if strings.HasPrefix(u.Host, ":") {
		// Empty host, i.e. ":port", use IPv4 localhost
		u.Host = "127.0.0.1" + u.Host
	} else if strings.HasPrefix(u.Host, "0.0.0.0:") {
		// IPv4 all zeroes host, convert to IPv4 localhost
		u.Host = "127.0.0.1" + u.Host[7:]
	} else if strings.HasPrefix(u.Host, "[::]:") {
		// IPv6 all zeroes host, convert to IPv6 localhost
		u.Host = "[::1]" + u.Host[4:]
	}

	return u.String()
}

// IsValidAPIKey returns true when the given API key is valid, including both
// the value in config and any overrides
func (c GUIConfiguration) IsValidAPIKey(apiKey string) bool {
	switch apiKey {
	case "":
		return false

	case c.APIKey, os.Getenv("STGUIAPIKEY"):
		return true

	default:
		return false
	}
}

func (c GUIConfiguration) Copy() GUIConfiguration {
	return c
}
