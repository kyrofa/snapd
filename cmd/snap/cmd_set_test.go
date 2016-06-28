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
	"os/user"
	"path/filepath"
	"sort"

	"gopkg.in/check.v1"

	"github.com/snapcore/snapd/arch"
	snaprun "github.com/snapcore/snapd/cmd/snap"
	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/snap/snaptest"
	"github.com/snapcore/snapd/testutil"
)

func (s *SnapSuite) TestSnapRunHookIntegration(c *check.C) {
	// mock installed snap
	dirs.SetRootDir(c.MkDir())
	defer func() { dirs.SetRootDir("/") }()

	snaptest.MockSnap(c, string(mockYaml), &snap.SideInfo{
		Revision: snap.R(42),
	})

	// and mock the server
	s.mockServer(c)

	// redirect exec
	execArg0 := ""
	execArgs := []string{}
	execEnv := []string{}
	restorer := snaprun.MockSyscallExec(func(arg0 string, args []string, envv []string) error {
		execArg0 = arg0
		execArgs = args
		execEnv = envv
		return nil
	})
	defer restorer()

	// Run a hook from the active revision
	err := snaprun.SnapRunHook("snapname", "hook-name", "")
	c.Assert(err, check.IsNil)
	c.Check(execArg0, check.Equals, "/usr/bin/ubuntu-core-launcher")
	c.Check(execArgs, check.DeepEquals, []string{
		"/usr/bin/ubuntu-core-launcher",
		"snap.snapname.hook.hook-name",
		"snap.snapname.hook.hook-name",
		"/usr/lib/snapd/snap-exec",
		filepath.Join(dirs.GlobalRootDir, "/snap/snapname/42/meta/hooks/hook-name")})
	c.Check(execEnv, testutil.Contains, "SNAP_REVISION=42")
}

func (s *SnapSuite) TestSnapRunHookSpecificRevisionIntegration(c *check.C) {
	// mock installed snap
	dirs.SetRootDir(c.MkDir())
	defer func() { dirs.SetRootDir("/") }()

	// Create both revisions 41 and 42
	snaptest.MockSnap(c, string(mockYaml), &snap.SideInfo{
		Revision: snap.R(41),
	})
	snaptest.MockSnap(c, string(mockYaml), &snap.SideInfo{
		Revision: snap.R(42),
	})

	// and mock the server
	s.mockServer(c)

	// redirect exec
	execArg0 := ""
	execArgs := []string{}
	execEnv := []string{}
	restorer := snaprun.MockSyscallExec(func(arg0 string, args []string, envv []string) error {
		execArg0 = arg0
		execArgs = args
		execEnv = envv
		return nil
	})
	defer restorer()

	// Run a hook on revision 41
	err := snaprun.SnapRunHook("snapname", "hook-name", "41")
	c.Assert(err, check.IsNil)
	c.Check(execArg0, check.Equals, "/usr/bin/ubuntu-core-launcher")
	c.Check(execArgs, check.DeepEquals, []string{
		"/usr/bin/ubuntu-core-launcher",
		"snap.snapname.hook.hook-name",
		"snap.snapname.hook.hook-name",
		"/usr/lib/snapd/snap-exec",
		filepath.Join(dirs.GlobalRootDir, "/snap/snapname/41/meta/hooks/hook-name")})
	c.Check(execEnv, testutil.Contains, "SNAP_REVISION=41")
}

func (s *SnapSuite) TestSnapRunHookMissingRevisionIntegration(c *check.C) {
	// mock installed snap
	dirs.SetRootDir(c.MkDir())
	defer func() { dirs.SetRootDir("/") }()

	// Only create revision 42
	snaptest.MockSnap(c, string(mockYaml), &snap.SideInfo{
		Revision: snap.R(42),
	})

	// and mock the server
	s.mockServer(c)

	// redirect exec
	restorer := snaprun.MockSyscallExec(func(arg0 string, args []string, envv []string) error {
		return nil
	})
	defer restorer()

	// Attempt to run a hook on revision 41, which doesn't exist
	err := snaprun.SnapRunHook("snapname", "hook-name", "41")
	c.Assert(err, check.NotNil)
	c.Check(err, check.ErrorMatches, "cannot find installed snap \"snapname\" at revision 41")
}

func (s *SnapSuite) TestSnapRunHookInvalidRevisionIntegration(c *check.C) {
	err := snaprun.SnapRunHook("snapname", "hook-name", "invalid")
	c.Assert(err, check.NotNil)
	c.Check(err, check.ErrorMatches, "invalid snap revision: \"invalid\"")
}

func (s *SnapSuite) mockServer(c *check.C) {
	n := 0
	s.RedirectClientToTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch n {
		case 0:
			c.Check(r.Method, check.Equals, "GET")
			c.Check(r.URL.Path, check.Equals, "/v2/snaps")
			fmt.Fprintln(w, `{"type": "sync", "result": [{"name": "snapname", "status": "active", "version": "1.0", "developer": "someone", "revision":42}]}`)
		default:
			c.Fatalf("expected to get 1 requests, now on %d", n+1)
		}

		n++
	})
}
