package main

import (
	"embed"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed web/index.html
var webFS embed.FS

var (
	listenAddr = envOr("LISTEN", ":8080")
	trustProxy = envBool("TRUST_PROXY", false)
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return strings.EqualFold(v, "true")
}

func pickIP(r *http.Request) net.IP {
	if trustProxy {
		if ff := r.Header.Get("X-Forwarded-For"); ff != "" {
			first := strings.TrimSpace(strings.Split(ff, ",")[0])
			if ip := net.ParseIP(first); ip != nil {
				return ip
			}
		}
		if xr := r.Header.Get("X-Real-IP"); xr != "" {
			if ip := net.ParseIP(xr); ip != nil {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = strings.Split(r.RemoteAddr, ":")[0]
	}
	host = strings.Split(host, "%")[0]
	return net.ParseIP(host)
}

func familyOf(ip net.IP) string {
	if ip != nil && ip.To4() != nil {
		return "ipv4"
	}
	return "ipv6"
}

func classify(ip net.IP) (string, bool) {
	if ip == nil {
		return "unknown", false
	}
	switch {
	case ip.IsLoopback():
		return "loopback", false
	case ip.IsPrivate():
		return "private", true
	case ip.IsLinkLocalUnicast():
		return "link-local", false
	case ip.IsUnspecified():
		return "unspecified", false
	default:
		return "public", false
	}
}

type addr struct {
	IP      string `json:"ip"`
	Family  string `json:"family"`
	Class   string `json:"class"`
	Private bool   `json:"private"`
}

func toAddr(ip net.IP) addr {
	s := ip.String()
	if ip4 := ip.To4(); ip4 != nil {
		s = ip4.String()
	}
	kind, priv := classify(ip)
	return addr{IP: s, Family: familyOf(ip), Class: kind, Private: priv}
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if r.URL.Path == "/" {
			w.Header().Set("Cache-Control", "no-store")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s from %s (%s) %s", r.Method, r.URL.Path, r.Proto, r.RemoteAddr, familyOf(pickIP(r)), time.Since(start))
	})
}

type pageData struct {
	Client addr
}

func main() {
	tmpl := template.Must(template.ParseFS(webFS, "web/index.html"))

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, pageData{Client: toAddr(pickIP(r))}); err != nil {
			log.Printf("render index: %v", err)
		}
	})

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           middleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("IPv6 status server starting on %s", listenAddr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}