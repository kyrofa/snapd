// -*- Mode: Go; indent-tabs-mode: t -*-
// +build !integrationcoverage

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main_test

import (
	"fmt"
	"net/http"

	"gopkg.in/check.v1"

	snapset "github.com/snapcore/snapd/cmd/snap"
	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/snap/snaptest"
)

var validApplyYaml = []byte(`name: snapname
version: 1.0
hooks:
 apply-config:
`)

func (s *SnapSuite) TestInvalidSetParameters(c *check.C) {
	invalidParameters := []string{"set", "snap-name", "key", "value"}
	_, err := snapset.Parser().ParseArgs(invalidParameters)
	c.Check(err, check.ErrorMatches, ".*invalid config:.*(want key=value).*")
}

func (s *SnapSuite) TestSnapApplyConfigIntegration(c *check.C) {
	// mock installed snap
	dirs.SetRootDir(c.MkDir())
	defer func() { dirs.SetRootDir("/") }()

	snaptest.MockSnap(c, string(validApplyYaml), &snap.SideInfo{
		Revision: snap.R(42),
	})

	// and mock the server
	s.mockSetConfigServer(c)

	// Set a config value for the active snap
	err := snapset.ApplyConfig("snapname", map[string]string{"key": "value"})
	c.Assert(err, check.IsNil)
}

func (s *SnapSuite) mockSetConfigServer(c *check.C) {
	s.RedirectClientToTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/snaps/snapname/config":
			c.Check(r.Method, check.Equals, "PUT")
			c.Check(DecodedRequestBody(c, r), check.DeepEquals, map[string]interface{}{
				"config": map[string]interface{}{"key": "value"},
			})
			fmt.Fprintln(w, `{"type":"async", "status-code": 202, "change": "zzz"}`)
		case "/v2/changes/zzz":
			c.Check(r.Method, check.Equals, "GET")
			fmt.Fprintln(w, `{"type":"sync", "result":{"ready": true, "status": "Done"}}`)
		default:
			c.Fatalf("unexpected path %q", r.URL.Path)
		}
	})
}
