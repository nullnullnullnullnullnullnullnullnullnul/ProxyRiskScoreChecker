package proxy

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Proxy
	}{
		{
			"http with auth",
			"http://user:pass@example.com:8080",
			Proxy{Protocol: "http", User: "user", Password: "pass", Host: "example.com", Port: "8080"},
		},
		{
			"https no auth",
			"https://example.com:443",
			Proxy{Protocol: "https", Host: "example.com", Port: "443"},
		},
		{
			"socks5 no auth",
			"socks5://10.0.0.1:1080",
			Proxy{Protocol: "socks5", Host: "10.0.0.1", Port: "1080"},
		},
		{
			"no protocol with auth",
			"user:pass@example.com:8080",
			Proxy{Protocol: "http", User: "user", Password: "pass", Host: "example.com", Port: "8080"},
		},
		{
			"host port only",
			"10.0.0.1:8080",
			Proxy{Protocol: "http", Host: "10.0.0.1", Port: "8080"},
		},
		{
			"provider format",
			"10.0.0.1:8080:user:pass",
			Proxy{Protocol: "http", Host: "10.0.0.1", Port: "8080", User: "user", Password: "pass"},
		},
		{
			"leading and trailing whitespace",
			"  http://10.0.0.1:80  ",
			Proxy{Protocol: "http", Host: "10.0.0.1", Port: "80"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("Parse(%q) = %+v; want %+v", c.in, got, c.want)
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"not-a-proxy",
		"http://example.com",   // missing port
		"host:port:user",       // 3 colons, ambiguous
		":8080",                // empty host
		"host:notaport",        // non-numeric port
		"ftp://example.com:21", // unsupported protocol
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			if _, err := Parse(in); err == nil {
				t.Errorf("Parse(%q) succeeded; want error", in)
			}
		})
	}
}

func TestProxyURL(t *testing.T) {
	cases := []struct {
		in   Proxy
		want string
	}{
		{Proxy{Protocol: "http", Host: "h", Port: "8080"}, "http://h:8080"},
		{
			Proxy{Protocol: "socks5", User: "u", Password: "p", Host: "h", Port: "1080"},
			"socks5://u:p@h:1080",
		},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			if got := c.in.URL(); got != c.want {
				t.Errorf("URL() = %q; want %q", got, c.want)
			}
		})
	}
}
