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

package snap

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/snapcore/snapd/release"
)

var implicitSlots = []string{
	"firewall-control",
	"home",
	"locale-control",
	"log-observe",
	"mount-observe",
	"network",
	"network-bind",
	"network-control",
	"network-observe",
	"snapd-control",
	"system-observe",
	"timeserver-control",
	"timezone-control",
}

var implicitClassicSlots = []string{
	"cups-control",
	"gsettings",
	"network-manager",
	"opengl",
	"pulseaudio",
	"unity7",
	"x11",
}

// AddImplicitSlots adds implicitly defined slots to a given snap.
//
// Only the OS snap has implicit slots.
//
// It is assumed that slots have names matching the interface name. Existing
// slots are not changed, only missing slots are added.
func AddImplicitSlots(snapInfo *Info) {
	if snapInfo.Type != TypeOS {
		return
	}
	for _, ifaceName := range implicitSlots {
		if _, ok := snapInfo.Slots[ifaceName]; !ok {
			snapInfo.Slots[ifaceName] = makeImplicitSlot(snapInfo, ifaceName)
		}
	}
	if !release.OnClassic {
		return
	}
	for _, ifaceName := range implicitClassicSlots {
		if _, ok := snapInfo.Slots[ifaceName]; !ok {
			snapInfo.Slots[ifaceName] = makeImplicitSlot(snapInfo, ifaceName)
		}
	}
}

func makeImplicitSlot(snapInfo *Info, ifaceName string) *SlotInfo {
	return &SlotInfo{
		Name:      ifaceName,
		Snap:      snapInfo,
		Interface: ifaceName,
	}
}

// AddImplicitHooks adds hooks from the snap's hookdir to the snap info.
//
// Existing hooks (i.e. ones defined in the YAML) are not changed; only missing
// hooks are added.
func AddImplicitHooks(snapInfo *Info, snapf Container) error {
	var fileInfos []os.FileInfo
	var err error
	if snapf != nil {
		fileInfos
	}
	// First of all, check to ensure the hooks directory exists. If it doesn't,
	// it's not an error-- there's just nothing to do.
	hooksDir := snapInfo.HooksDir()
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return nil
	}

	fileInfos, err := ioutil.ReadDir(hooksDir)
	if err != nil {
		return fmt.Errorf("unable to read hooks directory: %s", err)
	}

	for _, fileInfo := range fileInfos {
		hookName := fileInfo.Name()
		// Don't overwrite a hook that has already been loaded from the YAML
		if _, ok := snapInfo.Hooks[hookName]; ok {
			continue
		}

		snapInfo.Hooks[hookName] = &HookInfo{Snap: snapInfo, Name: hookName}
	}

	return nil
}
