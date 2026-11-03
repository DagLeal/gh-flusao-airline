package app

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type attempt struct {
	Count int
	Until time.Time
}

func loginLimiter(next http.Handler) http.Handler {
	var mu sync.Mutex
	attempts := map[string]attempt{}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/login" {
			next.ServeHTTP(w, r)
			return
		}
		ip, _, e := net.SplitHostPort(r.RemoteAddr)
		if e != nil {
			ip = r.RemoteAddr
		}
		mu.Lock()
		now := time.Now()
		v := attempts[ip]
		if now.After(v.Until) {
			v = attempt{Until: now.Add(15 * time.Minute)}
		}
		if v.Count >= 20 {
			mu.Unlock()
			w.Header().Set("Retry-After", "900")
			http.Error(w, "Muitas tentativas. Aguarde 15 minutos.", http.StatusTooManyRequests)
			return
		}
		v.Count++
		attempts[ip] = v
		if len(attempts) > 10000 {
			for key, value := range attempts {
				if now.After(value.Until) {
					delete(attempts, key)
				}
			}
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
