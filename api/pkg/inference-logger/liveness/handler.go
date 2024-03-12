package liveness

import (
	"net/http"
	"strings"

	"knative.dev/networking/pkg/http/header"
	"knative.dev/pkg/network"
)

type Probe struct {
	// next The handler calls next handler in the composed handler.
	next http.Handler

	// inner The inner handler proxies the request to the model/transformer.
	inner http.Handler
}

func NewProbe(inner, next http.Handler) *Probe {
	return &Probe{
		next:  next,
		inner: inner,
	}
}

func (p *Probe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.isLivenessCheck(r) {
		p.inner.ServeHTTP(w, r)
		return
	}
	p.next.ServeHTTP(w, r)
}

func (p *Probe) isLivenessCheck(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("User-Agent"), header.KubeProbeUAPrefix) &&
		r.Header.Get(network.KubeletProbeHeaderName) != ""
}
