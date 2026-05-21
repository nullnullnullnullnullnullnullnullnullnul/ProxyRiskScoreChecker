// Command proxyriskscore reads a list of proxies, resolves each one's
// outbound IP, queries the IPQualityScore reputation API for that IP, and
// writes the proxies whose reported fraud_score is 0 to the output file.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/nullnullnullnullnullnullnullnullnullnul/ProxyRiskScoreChecker/internal/ipqs"
	"github.com/nullnullnullnullnullnullnullnullnullnul/ProxyRiskScoreChecker/internal/proxy"
)

const (
	defaultInput      = "proxies.txt"
	defaultOutput     = "clean.txt"
	defaultStrictness = 0
	defaultTimeout    = 10 * time.Second
	envAPIKey         = "IPQS_API_KEY"
)

type options struct {
	input      string
	output     string
	strictness int
	apiKey     string
	timeout    time.Duration
}

func parseFlags(args []string) (options, error) {
	var opts options
	fs := flag.NewFlagSet("proxyriskscore", flag.ContinueOnError)
	fs.StringVar(&opts.input, "input", defaultInput, "input file with one proxy per line")
	fs.StringVar(&opts.output, "output", defaultOutput, "output file for proxies with fraud_score=0")
	fs.IntVar(&opts.strictness, "strictness", defaultStrictness, "IPQS strictness level (0-3)")
	fs.StringVar(&opts.apiKey, "api-key", "", "IPQS API key (falls back to $"+envAPIKey+")")
	fs.DurationVar(&opts.timeout, "timeout", defaultTimeout, "per-request timeout")
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	if opts.apiKey == "" {
		opts.apiKey = os.Getenv(envAPIKey)
	}
	if opts.apiKey == "" {
		return opts, fmt.Errorf("missing IPQS API key: pass --api-key or set $%s", envAPIKey)
	}
	if opts.strictness < 0 || opts.strictness > 3 {
		return opts, fmt.Errorf("strictness must be in [0,3], got %d", opts.strictness)
	}
	return opts, nil
}

func readProxies(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filename, err)
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			out = append(out, s)
		}
	}
	return out, nil
}

func writeProxies(filename string, lines []string) error {
	if len(lines) == 0 {
		return os.WriteFile(filename, nil, 0o644)
	}
	return os.WriteFile(filename, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func run(ctx context.Context, log *slog.Logger, opts options) error {
	raws, err := readProxies(opts.input)
	if err != nil {
		return err
	}
	if len(raws) == 0 {
		return errors.New("input file is empty")
	}
	log.Info("loaded proxies", "count", len(raws), "file", opts.input)

	validator := proxy.Validator{Timeout: opts.timeout}
	client := ipqs.Client{APIKey: opts.apiKey, Timeout: opts.timeout}

	var clean []string
	for _, raw := range raws {
		p, err := proxy.Parse(raw)
		if err != nil {
			log.Warn("skip: parse failed", "proxy", raw, "err", err)
			continue
		}
		outbound, err := validator.OutboundIP(ctx, p)
		if err != nil {
			log.Warn("skip: proxy probe failed", "proxy", raw, "err", err)
			continue
		}
		resp, err := client.CheckIP(ctx, outbound, opts.strictness)
		if err != nil {
			log.Warn("skip: ipqs lookup failed", "proxy", raw, "ip", outbound, "err", err)
			continue
		}
		if resp.FraudScore == 0 {
			log.Info("clean", "proxy", raw, "ip", outbound)
			clean = append(clean, raw)
		} else {
			log.Info("flagged", "proxy", raw, "ip", outbound, "score", resp.FraudScore)
		}
	}

	if err := writeProxies(opts.output, clean); err != nil {
		return fmt.Errorf("write %s: %w", opts.output, err)
	}
	log.Info("done", "clean", len(clean), "total", len(raws), "output", opts.output)
	return nil
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Error("invalid arguments", "err", err)
		os.Exit(2)
	}
	if err := run(context.Background(), log, opts); err != nil {
		log.Error("run failed", "err", err)
		os.Exit(1)
	}
}
