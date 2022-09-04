/*
Package doubleforward provides a helper for connecting goproxy through a second forward proxy, supporting HTTPS.
*/
package doubleforward

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// ConnectDialViaProxy can be used as (*goproxy.ProxyHttpServer).ConnectDial to connect via a forward proxy.
// If the provided hostname cannot be parsed by url.Parse, ConnectDialViaProxy returns an error.
func ConnectDialViaProxy(host, username, password string) (func(network, addr string) (net.Conn, error), error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	proxyAddr := u.Host
	if u.Port() == "" {
		if u.Scheme == "https" {
			proxyAddr = net.JoinHostPort(proxyAddr, "443")
		} else {
			proxyAddr = net.JoinHostPort(proxyAddr, "80")
		}
	}

	useTLS := u.Port() == "443" || u.Scheme == "https"

	// make request template
	var reqTemplate strings.Builder
	fmt.Fprint(&reqTemplate, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n")
	if username != `` {
		credentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		fmt.Fprintf(&reqTemplate, "Proxy-Authorization: Basic %s\r\n", credentials)
	}
	reqTemplate.WriteString("\r\n")

	return func(network, addr string) (net.Conn, error) {
		// dial connection
		var conn net.Conn
		var err error
		if useTLS {
			conn, err = tls.Dial(`tcp`, proxyAddr, nil)
		} else {
			conn, err = net.Dial(`tcp`, proxyAddr)
		}
		if err != nil {
			return nil, err
		}

		// send CONNECT request
		_, err = fmt.Fprintf(conn, reqTemplate.String(), addr, addr)
		if err != nil {
			return nil, err
		}

		// expect 200 OK response
		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(`received response with status %q in respone to CONNECT request to proxy at %s`,
				resp.Status, proxyAddr)
		}

		return conn, nil
	}, nil
}
