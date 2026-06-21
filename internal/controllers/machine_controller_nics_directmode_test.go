// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"testing"

	providernetworkinterface "github.com/ironcore-dev/libvirt-provider/internal/plugins/networkinterface"
)

// Verifies the selectable host-device NIC binding mode (--network-interface-direct-mode):
// "tap" emits a plain tap (type='ethernet' + <target managed='no'/>), while "macvtap" (and the
// empty/default) keep the original macvtap (type='direct' + <source mode='bridge'/>). Also checks
// the inverse parser round-trips the tap form back to provider Direct{Dev}.
func TestProviderNetworkInterfaceToLibvirtDirectMode(t *testing.T) {
	nic := &providernetworkinterface.NetworkInterface{
		Direct: &providernetworkinterface.Direct{Dev: "dtapvf_0"},
	}

	// tap mode → ethernet + target, no macvtap source.
	res, err := providerNetworkInterfaceToLibvirt("nic0", nic, "tap")
	if err != nil {
		t.Fatalf("tap: %v", err)
	}
	if res.iface == nil || res.iface.Source == nil {
		t.Fatalf("tap: nil iface/source: %+v", res)
	}
	if res.iface.Source.Ethernet == nil {
		t.Error("tap: expected Source.Ethernet (type='ethernet')")
	}
	if res.iface.Source.Direct != nil {
		t.Error("tap: unexpected Source.Direct (macvtap) in tap mode")
	}
	if res.iface.Target == nil || res.iface.Target.Dev != "dtapvf_0" || res.iface.Target.Managed != "no" {
		t.Errorf("tap: expected Target{Dev:dtapvf_0, Managed:no}, got %+v", res.iface.Target)
	}

	// default ("") and explicit "macvtap" → macvtap direct source, no ethernet/target.
	for _, mode := range []string{"macvtap", ""} {
		res, err := providerNetworkInterfaceToLibvirt("nic0", nic, mode)
		if err != nil {
			t.Fatalf("mode %q: %v", mode, err)
		}
		if res.iface.Source.Direct == nil ||
			res.iface.Source.Direct.Dev != "dtapvf_0" ||
			res.iface.Source.Direct.Mode != "bridge" {
			t.Errorf("mode %q: expected macvtap Source.Direct{dtapvf_0, bridge}, got %+v", mode, res.iface.Source)
		}
		if res.iface.Source.Ethernet != nil || res.iface.Target != nil {
			t.Errorf("mode %q: unexpected ethernet/target in macvtap mode", mode)
		}
	}

	// inverse parser maps the tap (ethernet/target) form back to provider Direct{Dev}.
	tapRes, err := providerNetworkInterfaceToLibvirt("nic0", nic, "tap")
	if err != nil {
		t.Fatalf("tap (inverse setup): %v", err)
	}
	back, err := libvirtInterfaceToProviderNetworkInterface(tapRes.iface)
	if err != nil {
		t.Fatalf("inverse(tap): %v", err)
	}
	if back.Direct == nil || back.Direct.Dev != "dtapvf_0" {
		t.Errorf("inverse(tap): expected Direct{Dev:dtapvf_0}, got %+v", back)
	}
}
