//go:generate go run github.com/Al2Klimov/go-gen-source-repos

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	. "github.com/Al2Klimov/go-monplug-utils"
	"github.com/masif-upgrader/agent/v1"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type loadThresholds struct {
	warn, crit OptionalThreshold
}

func (lt *loadThresholds) mkFlags(cli *flag.FlagSet, subject string) {
	cli.Var(&lt.warn, subject+"-warn", "e.g. @~:42")
	cli.Var(&lt.crit, subject+"-crit", "e.g. @~:42")
}

type loadsThresholds struct {
	m1, m5, m15 loadThresholds
}

func (lt *loadsThresholds) mkFlags(cli *flag.FlagSet, subject string) {
	lt.m1.mkFlags(cli, subject+"-1m")
	lt.m5.mkFlags(cli, subject+"-5m")
	lt.m15.mkFlags(cli, subject+"-15m")
}

type allThresholds struct {
	query, install, update, configure, remove, purge, error loadsThresholds
}

func main() {
	os.Exit(ExecuteCheck(onTerminal, checkMasifupgraderAgent))
}

func onTerminal() (output string) {
	return fmt.Sprintf(
		"For the terms of use, the source code and the authors\n"+
			"see the projects this program is assembled from:\n\n  %s\n",
		strings.Join(GithubcomAl2klimovGo_gen_source_repos, "\n  "),
	)
}

var loadsThresholdsToMetrics = []struct {
	subject   string
	threshold func(*allThresholds) *loadsThresholds
	metric    func(*v1.Load) *[3]float64
}{
	{
		"query",
		func(at *allThresholds) *loadsThresholds { return &at.query },
		func(load *v1.Load) *[3]float64 { return &load.Query },
	},
	{
		"install",
		func(at *allThresholds) *loadsThresholds { return &at.install },
		func(load *v1.Load) *[3]float64 { return &load.Install },
	},
	{
		"update",
		func(at *allThresholds) *loadsThresholds { return &at.update },
		func(load *v1.Load) *[3]float64 { return &load.Update },
	},
	{
		"configure",
		func(at *allThresholds) *loadsThresholds { return &at.configure },
		func(load *v1.Load) *[3]float64 { return &load.Configure },
	},
	{
		"remove",
		func(at *allThresholds) *loadsThresholds { return &at.remove },
		func(load *v1.Load) *[3]float64 { return &load.Remove },
	},
	{
		"purge",
		func(at *allThresholds) *loadsThresholds { return &at.purge },
		func(load *v1.Load) *[3]float64 { return &load.Purge },
	},
	{
		"error",
		func(at *allThresholds) *loadsThresholds { return &at.error },
		func(load *v1.Load) *[3]float64 { return &load.Error },
	},
}

var loadThresholdsToMetrics = []struct {
	subject   string
	threshold func(*loadsThresholds) *loadThresholds
	metric    func(*[3]float64) float64
}{
	{
		"1m",
		func(lt *loadsThresholds) *loadThresholds { return &lt.m1 },
		func(load *[3]float64) float64 { return load[0] },
	},
	{
		"5m",
		func(lt *loadsThresholds) *loadThresholds { return &lt.m5 },
		func(load *[3]float64) float64 { return load[1] },
	},
	{
		"15m",
		func(lt *loadsThresholds) *loadThresholds { return &lt.m15 },
		func(load *[3]float64) float64 { return load[2] },
	},
}

func checkMasifupgraderAgent() (output string, perfdata PerfdataCollection, errs map[string]error) {
	cli := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	restsock := cli.String("restsock", "", "ReST API socket path")

	var respTime loadThresholds

	respTime.mkFlags(cli, "resptime")

	var threshold allThresholds

	threshold.query.mkFlags(cli, "query")
	threshold.install.mkFlags(cli, "install")
	threshold.update.mkFlags(cli, "update")
	threshold.configure.mkFlags(cli, "configure")
	threshold.remove.mkFlags(cli, "remove")
	threshold.purge.mkFlags(cli, "purge")
	threshold.error.mkFlags(cli, "error")

	if errCli := cli.Parse(os.Args[1:]); errCli != nil {
		os.Exit(3)
	}

	if *restsock == "" {
		fmt.Fprintln(os.Stderr, "-restsock missing")
		cli.Usage()
		os.Exit(3)
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var dialer net.Dialer
				return dialer.DialContext(ctx, "unix", *restsock)
			},
		},
	}

	start := time.Now()

	resp, errReq := client.Get("http://localhost/v1/load")
	if errReq != nil {
		return "", nil, map[string]error{
			fmt.Sprintf("GET http://%s/v1/load", url.PathEscape(*restsock)): errReq,
		}
	}

	body, errRA := ioutil.ReadAll(resp.Body)
	if errRA != nil {
		return "", nil, map[string]error{
			fmt.Sprintf("GET http://%s/v1/load", url.PathEscape(*restsock)): errRA,
		}
	}

	end := time.Now()

	if resp.StatusCode != 200 {
		return "", nil, map[string]error{
			fmt.Sprintf("GET http://%s/v1/load", url.PathEscape(*restsock)): errors.New(resp.Status),
		}
	}

	var payload v1.Load
	if errJU := json.Unmarshal(body, &payload); errJU != nil {
		return "", nil, map[string]error{
			fmt.Sprintf("GET http://%s/v1/load", url.PathEscape(*restsock)): errJU,
		}
	}

	perfdata = append(
		make(PerfdataCollection, 0, 2+len(loadsThresholdsToMetrics)*len(loadThresholdsToMetrics)),
		Perfdata{
			Label: "time",
			UOM:   "us",
			Value: float64(end.Sub(start)) / float64(time.Microsecond),
			Warn:  respTime.warn,
			Crit:  respTime.crit,
			Min:   OptionalNumber{true, 0},
		},
		Perfdata{
			Label: "size",
			UOM:   "B",
			Value: float64(len(body)),
			Min:   OptionalNumber{true, 0},
		},
	)

	for _, lsttm := range loadsThresholdsToMetrics {
		lst := lsttm.threshold(&threshold)
		lsm := lsttm.metric(&payload)

		for _, lttm := range loadThresholdsToMetrics {
			lt := lttm.threshold(lst)

			perfdata = append(perfdata, Perfdata{
				Label: fmt.Sprintf("%s_%s", lsttm.subject, lttm.subject),
				Value: lttm.metric(lsm),
				Warn:  lt.warn,
				Crit:  lt.crit,
				Min:   OptionalNumber{true, 0},
			})
		}
	}

	return
}
