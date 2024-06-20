package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
)

/**
 * Run below code after running below command:
 * $ mitmweb --ssl-insecure
 *
 * You can simply install mitmproxy by running below command:
 * $ brew install mitmproxy
 *
 * If you are using another OS, please refer to the official documentation:
 * https://docs.mitmproxy.org/stable/overview-installation/
 */

func TestHTTP1_WebSocket(t *testing.T) {
	testcases := []struct {
		name  string
		isTLS bool
	}{
		{name: "plain", isTLS: false}, // FIXME: There is a bug with proxy dialing on gorilla/websocket
		{name: "secure", isTLS: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(tt *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			time.Sleep(1 * time.Second)

			dialer := Dialer{
				Proxy:            http.ProxyURL(&url.URL{Scheme: "http", Host: "localhost:8080"}),
				HandshakeTimeout: 5 * time.Second,
			}

			var (
				scheme string
				port   int
			)
			if tc.isTLS {
				dialer.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
				scheme = "wss"
				port = 8082
			} else {
				scheme = "ws"
				port = 8081
			}

			ws, _, err := dialer.DialContext(ctx, fmt.Sprintf("%s://localhost:%d", scheme, port), nil)
			if err != nil {
				tt.Fatalf("failed to dial websocket: %v", err)
				return
			}
			defer ws.Close()

			maxRequests := 5
			currentRequest := 0
			for currentRequest < maxRequests {
				currentRequest++
				if err := ws.WriteMessage(TextMessage, []byte("hello, world!")); err != nil {
					tt.Fatal(err)
					return
				}

				_, message, err := ws.ReadMessage()
				if err != nil {
					tt.Fatalf("failed to read message: %v", err)
					return
				} else if string(message) != "hello, world!" {
					tt.Fatalf("unexpected message: %s", message)
					return
				}
			}

			_ = ws.WriteMessage(CloseMessage, FormatCloseMessage(CloseNormalClosure, ""))

			cancel()
		})
	}
}
