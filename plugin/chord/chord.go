package chord

import (
	"context"
	"encoding/json"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"io/ioutil"
	"net"
	"net/http"
)

const name = "chord"

// Chord is a plugin that returns the IP address which is fetched from chord
type Chord struct{}

// ServeDNS implements the plugin.Handler interface.
func (c Chord) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	serviceName := state.Name()
	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true
	var data map[string]string
	resp, err := http.Get("http://localhost:8080/DNS?serviceName=" + serviceName)
	if err != nil {
		return dns.RcodeServerFailure, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodystr := string(body)
	var rr dns.RR
	if err := json.Unmarshal([]byte(bodystr), &data); err == nil {
		//这里要把data中的数据写入给client返回的dns消息中
		rr = &dns.A{}
		rr.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass()}
		rr.(*dns.A).A = net.ParseIP(data["ip"]).To4()
	} else {
		return dns.RcodeServerFailure, err
	}
	a.Extra = []dns.RR{rr}
	w.WriteMsg(a)
	return 0, nil
}

// Name implements the Handler interface.
func (c Chord) Name() string { return name }
