// -*- Mode: Go; indent-tabs-mode: t -*-

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

package hookstate_test

import (
	"testing"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/hooks"
	"github.com/snapcore/snapd/overlord/hookstate"
	"github.com/snapcore/snapd/overlord/state"
)

func TestHookManager(t *testing.T) { TestingT(t) }

type hookManagerSuite struct {
	state   *state.State
	manager *hookstate.HookManager
}

var _ = Suite(&hookManagerSuite{})

func (s *hookManagerSuite) SetUpTest(c *C) {
	dirs.SetRootDir(c.MkDir())
	s.state = state.New(nil)
	manager, err := hookstate.Manager(s.state)
	c.Assert(err, IsNil)
	s.manager = manager
}

func (s *hookManagerSuite) TearDownTest(c *C) {
	s.manager.Stop()
	dirs.SetRootDir("")
}

func (s *hookManagerSuite) TestSmoke(c *C) {
	s.manager.Ensure()
	s.manager.Wait()
}

func (s *hookManagerSuite) TestRunHookInstruction(c *C) {
	s.state.Lock()
	defer s.state.Unlock()

	taskSet, err := hookstate.RunHook(s.state, "test-snap", "test-hook")
	c.Assert(err, IsNil, Commentf("RunHook unexpectedly failed"))
	c.Assert(taskSet, NotNil, Commentf("Expected RunHook to provide a task set"))

	tasks := taskSet.Tasks()
	c.Assert(tasks, HasLen, 1, Commentf("Expected task set to contain 1 task"))

	task := tasks[0]
	c.Check(task.Kind(), Equals, "run-hook")

	var hook hooks.HookRef
	err = task.Get("hook", &hook)
	c.Check(err, IsNil, Commentf("Expected task to contain hook"))
	c.Check(hook.Snap, Equals, "test-snap")
	c.Check(hook.Hook, Equals, "test-hook")
}

func (s *hookManagerSuite) TestRunHookTask(c *C) {
	s.state.Lock()
	taskSet, err := hookstate.RunHook(s.state, "test-snap", "test-hook")
	c.Assert(err, IsNil, Commentf("RunHook unexpectedly failed"))
	c.Assert(taskSet, NotNil, Commentf("Expected RunHook to provide a task set"))

	change := s.state.NewChange("kind", "summary")
	change.AddAll(taskSet)
	s.state.Unlock()

	s.manager.Ensure()
	s.manager.Wait()

	s.state.Lock()
	defer s.state.Unlock()

	tasks := taskSet.Tasks()
	c.Assert(tasks, HasLen, 1, Commentf("Expected task set to contain 1 task"))
	task := tasks[0]

	c.Check(task.Kind(), Equals, "run-hook")
	c.Check(task.Status(), Equals, state.DoneStatus)
	c.Check(change.Status(), Equals, state.DoneStatus)
}
