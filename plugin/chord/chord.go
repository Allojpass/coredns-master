package chord

import (
	"context"
	"encoding/json"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"io/ioutil"
	"net"
	"net/http"
)

const name = "chord"

// Chord is a plugin that returns the IP address which is fetched from chord
type Chord struct {
	Next plugin.Handler
}

// 用来接收Chord返回值的类型
type Receive struct {
	Msg string   `json:"msg"`
	Ips []string `json:"ips"`
}

// ServeDNS implements the plugin.Handler interface.
func (c Chord) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	serviceName := state.Name()
	answers := []dns.RR{}

	var data Receive
	resp, err := http.Get("http://localhost:12000/test/serviceDiscovery?serviceName=" + serviceName)
	if err != nil {
		//这里的err用于判断是否能够成功发送请求，当后端没启动的时候，无法正确建立连接，这里的err就会不为nil
		//所以如果chord后端出现问题那么直接调用下一个插件也就是forward
		return plugin.NextOrFailure(c.Name(), c.Next, ctx, w, r)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodystr := string(body)
	var rr dns.RR
	if err := json.Unmarshal([]byte(bodystr), &data); err == nil {
		//这里要把data中的数据写入给client返回的dns消息中
		//同时这里需要加入对于后端是否能解析该域名的判断
		//如果后端返回的message中说明无法处理这个请求，那么需要调用forward插件也就是进入插件链下一个插件
		if data.Msg == "Unable to process request" {
			return plugin.NextOrFailure(c.Name(), c.Next, ctx, w, r)
		} else if data.Msg == "Success" {
			//循环添加data中的每一个ip到最后的返回answer中
			for _, value := range data.Ips {
				rr = &dns.A{}
				rr.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass(), Ttl: 3600}
				rr.(*dns.A).A = net.ParseIP(value).To4()
				answers = append(answers, rr)
			}
		} else {
			//当后端未能正确返回的时候说明Chord也出现了问题，同样丢给forward插件处理
			return plugin.NextOrFailure(c.Name(), c.Next, ctx, w, r)
		}
	} else {
		//这里只有当body解析出错的时候err才会不为nil，但即使body为空这里仍然能走完解析流程不会报错
		return plugin.NextOrFailure(c.Name(), c.Next, ctx, w, r)
	}
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers
	w.WriteMsg(m)
	return 0, nil
}

// Name implements the Handler interface.
func (c Chord) Name() string { return name }
