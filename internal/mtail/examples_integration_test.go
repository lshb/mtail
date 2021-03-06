// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.
// +build integration

package mtail_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/mtail/internal/metrics"
	"github.com/google/mtail/internal/metrics/datum"
	"github.com/google/mtail/internal/mtail"
	"github.com/google/mtail/internal/mtail/golden"
	"github.com/google/mtail/internal/testutil"
	"github.com/google/mtail/internal/watcher"
)

var exampleProgramTests = []struct {
	programfile string // Example program file.
	logfile     string // Sample log input.
	goldenfile  string // Expected metrics after processing.
}{
	{
		"examples/rsyncd.mtail",
		"testdata/rsyncd.log",
		"testdata/rsyncd.golden",
	},
	{
		"examples/sftp.mtail",
		"testdata/sftp_chroot.log",
		"testdata/sftp_chroot.golden",
	},
	{
		"examples/dhcpd.mtail",
		"testdata/anonymised_dhcpd_log",
		"testdata/anonymised_dhcpd_log.golden",
	},
	{
		"examples/ntpd.mtail",
		"testdata/ntp4",
		"testdata/ntp4.golden",
	},
	{
		"examples/ntpd_peerstats.mtail",
		"testdata/xntp3_peerstats",
		"testdata/xntp3_peerstats.golden",
	},
	{
		"examples/otherwise.mtail",
		"testdata/otherwise.log",
		"testdata/otherwise.golden",
	},
	{
		"examples/else.mtail",
		"testdata/else.log",
		"testdata/else.golden",
	},
	{
		"examples/types.mtail",
		"testdata/types.log",
		"testdata/types.golden",
	},
	{
		"examples/filename.mtail",
		"testdata/else.log",
		"testdata/filename.golden",
	},
	{
		"examples/logical.mtail",
		"testdata/logical.log",
		"testdata/logical.golden",
	},
	{
		"examples/strcat.mtail",
		"testdata/strcat.log",
		"testdata/strcat.golden",
	},
	{
		"examples/add_assign_float.mtail",
		"testdata/add_assign_float.log",
		"testdata/add_assign_float.golden",
	},
	{
		"examples/typed-comparison.mtail",
		"testdata/typed-comparison.log",
		"testdata/typed-comparison.golden",
	},
	{
		"examples/match-expression.mtail",
		"testdata/match-expression.log",
		"testdata/match-expression.golden",
	},
	{
		"examples/apache_combined.mtail",
		"testdata/apache-combined.log",
		"testdata/apache-combined.golden",
	},
	{
		"examples/apache_common.mtail",
		"testdata/apache-common.log",
		"testdata/apache-common.golden",
	},
	{
		"examples/metric-as-rvalue.mtail",
		"testdata/metric-as-rvalue.log",
		"testdata/metric-as-rvalue.golden",
	},
	{
		"examples/decorator.mtail",
		"testdata/decorator.log",
		"testdata/decorator.golden",
	},
	{
		"examples/stringy.mtail",
		"testdata/stringy.log",
		"testdata/stringy.golden",
	},
	{
		"examples/ip-addr.mtail",
		"testdata/ip-addr.log",
		"testdata/ip-addr.golden",
	},
	{
		"examples/vsftpd.mtail",
		"testdata/vsftpd_log",
		"testdata/vsftpd_log.golden",
	},
	{
		"examples/vsftpd.mtail",
		"testdata/vsftpd_xferlog",
		"testdata/vsftpd_xferlog.golden",
	},
	{
		"examples/lighttpd.mtail",
		"testdata/lighttpd_access.log",
		"testdata/lighttpd_accesslog.golden",
	},
	{
		"examples/mysql_slowqueries.mtail",
		"testdata/mysql_slowqueries.log",
		"testdata/mysql_slowqueries.golden",
	},
}

func TestExamplePrograms(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, tc := range exampleProgramTests {
		t.Run(fmt.Sprintf("%s on %s", tc.programfile, tc.logfile), func(t *testing.T) {
			w := watcher.NewFakeWatcher()
			store := metrics.NewStore()
			programFile := path.Join("../..", tc.programfile)
			mtail, err := mtail.New(store, w, mtail.ProgramPath(programFile), mtail.LogPathPatterns(tc.logfile), mtail.OneShot, mtail.OmitMetricSource, mtail.DumpAstTypes, mtail.DumpBytecode)
			if err != nil {
				t.Fatalf("create mtail failed: %s", err)
			}

			err = mtail.StartTailing()
			if err != nil {
				t.Fatalf("Start tailling failed: %s", err)
			}

			g, err := os.Open(tc.goldenfile)
			if err != nil {
				t.Fatalf("could not open golden file: %s", err)
			}
			defer g.Close()

			goldenStore := metrics.NewStore()
			golden.ReadTestData(g, tc.programfile, goldenStore)

			err = mtail.Close()
			if err != nil {
				t.Error(err)
			}

			diff := testutil.Diff(goldenStore, store, testutil.IgnoreUnexported(sync.RWMutex{}, datum.String{}))

			if diff != "" {
				t.Error(diff)
				t.Logf(" Golden metrics: %s", goldenStore.Metrics)
				t.Logf("Program metrics: %s", store.Metrics)
				t.Logf("yar\n%+v", store.Metrics)
			}
		})
	}
}

// This test only compiles examples, but has coverage over all examples
// provided.  This ensures we ship at least syntactically correct examples.
func TestCompileExamplePrograms(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	matches, err := filepath.Glob("../../examples/*.mtail")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range matches {
		name := filepath.Base(tc)
		t.Run(name, func(t *testing.T) {
			w := watcher.NewFakeWatcher()
			s := metrics.NewStore()
			mtail, err := mtail.New(s, w, mtail.ProgramPath(tc), mtail.CompileOnly, mtail.OmitMetricSource, mtail.DumpAstTypes, mtail.DumpBytecode)
			if err != nil {
				t.Fatal(err)
			}
			mtail.Close()
		})
	}
}
