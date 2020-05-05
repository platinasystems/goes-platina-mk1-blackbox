// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/platinasystems/test"
	"github.com/platinasystems/test/ethtool"
	"github.com/platinasystems/test/netport"
)

func TestMain(m *testing.M) {
	var ecode int
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, r)
			ecode = 1
		}
		if *XethStat {
			showXethStats()
		}
		if ecode != 0 {
			os.Exit(ecode)
		}
	}()
	assertFlags()
	if *test.DryRun {
		m.Run()
		return
	}
	if os.Geteuid() != 0 {
		panic("you aren't root")
	}
	if b, err := ioutil.ReadFile("/proc/net/unix"); err == nil {
		for _, atsock := range []string{
			"@xeth",
		} {
			if bytes.Index(b, []byte(atsock)) < 0 {
				panic(fmt.Errorf("no %s, are modules loaded?",
					atsock))
			}
		}
		for _, atsock := range []string{
			"@redisd",
			"@redis.reg",
			"@redis.pub",
			"@fe1",
		} {
			if bytes.Index(b, []byte(atsock)) < 0 {
				panic(fmt.Errorf("no %s, is goes running?",
					atsock))
			}
		}
	}
	netport.Init(*Goes)
	ethtool.Init()
	if testing.Verbose() {
		uutInfo()
	}
	ecode = m.Run()
}

func Test(t *testing.T) {
	mayRun(t, "net4", func(t *testing.T) {
		mayRun(t, "ping", pingNetTest)
		mayRun(t, "dhcp", dhcpNetTest)
		mayRun(t, "static", staticNetTest)
		mayRun(t, "gobgp", gobgpNetTest)
		mayRun(t, "bird", birdNetTest)
		mayRun(t, "frr", frrNetTest)
		test.SkipIfDryRun(t)
	})
	mayRun(t, "vlan4", func(t *testing.T) {
		mayRun(t, "ping", pingVlanTest)
		mayRun(t, "dhcp", dhcpVlanTest)
		mayRun(t, "slice", sliceVlanTest)
		mayRun(t, "static", staticVlanTest)
		mayRun(t, "gobgp", gobgpVlanTest)
		mayRun(t, "bird", birdVlanTest)
		mayRun(t, "frr", frrVlanTest)
		test.SkipIfDryRun(t)
	})
	t.Log("Skipping bridge test")
	if false {
		mayRun(t, "bridge", func(t *testing.T) {
			mayRun(t, "ping", pingBridgeTest)
			test.SkipIfDryRun(t)
		})
	}
	t.Log("Skipping nsif test")
	if false {
		mayRun(t, "nsif", func(t *testing.T) {
			mayRun(t, "ip4", nsifNetTest)
			mayRun(t, "ip6", nsifIp6NetTest)
		})
	}
	mayRun(t, "net6", func(t *testing.T) {
		mayRun(t, "ping", pingIp6NetTest)
		mayRun(t, "static", staticV6NetTest)
		mayRun(t, "ospf", frrNetV6OspfTest)
		mayRun(t, "bgp", frrNetV6BgpTest)
		mayRun(t, "isis", frrNetV6IsisTest)
		mayRun(t, "dhcp", dhcpNetV6Test)
	})
	mayRun(t, "vlan6", func(t *testing.T) {
		mayRun(t, "static", staticV6VlanTest)
		mayRun(t, "bgp", frrVlanV6BgpTest)
		mayRun(t, "isis", frrVlanV6IsisTest)
		mayRun(t, "dhcp", dhcpVlanV6Test)
	})
	mayRun(t, "multipath", func(t *testing.T) {
		mayRun(t, "ip4", mpNetTest)
		mayRun(t, "ip6", mpNetIp6Test)
	})
	mayRun(t, "routes", func(t *testing.T) {
		mayRun(t, "connective", routesNetTest)
	})

	test.SkipIfDryRun(t)
}

func mayRun(t *testing.T, name string, f func(*testing.T)) bool {
	var ret bool
	t.Helper()
	if !t.Failed() {
		ret = t.Run(name, f)
	}
	return ret
}

func uutInfo() {
	fmt.Println("---")
	defer fmt.Println("...")
	o, err := exec.Command(*Goes, "show", "buildid").Output()
	if err == nil && len(o) > 0 {
		fmt.Print(*Goes, ": |\n    buildid/", string(o))
	}
	pd := *PlatformDriver
	ko := pd
	if !strings.HasSuffix(ko, ".ko") {
		ko += ".ko"
	}
	if _, err = os.Stat(ko); err == nil {
		pd = ko
	}
	o, err = exec.Command("/sbin/modinfo", pd).Output()
	if err == nil && len(o) > 0 {
		const srcversion = "srcversion:"
		s := string(o)
		i := strings.Index(s, srcversion)
		if i > 0 {
			s = s[i+len(srcversion):]
			i = strings.Index(s, "\n")
			fmt.Print(pd, ": |\n    ",
				strings.TrimLeft(s[:i+1], " \t"))
		}
	}
}

func showXethStats() {
	const dn = "/sys/kernel/platina-mk1/xeth"
	fmt.Println("---")
	defer fmt.Println("...")
	fis, err := ioutil.ReadDir(dn)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, fi := range fis {
		bn := fi.Name()
		b, err := ioutil.ReadFile(filepath.Join(dn, bn))
		if err != nil {
			fmt.Print(bn, ": ", err, "\n")
		} else if s := string(b); s != "0\n" {
			fmt.Print(bn, ": ", s)
		}
	}
}
