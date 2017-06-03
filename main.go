package main

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"strconv"
)

var records = map[string]string{
	"test.service.": "192.168.0.2",
}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("Query for %s\n", q.Name)
			ip := records[q.Name]
			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else {
				config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
				c := new(dns.Client)

				msg := new(dns.Msg)
				msg.SetQuestion(dns.Fqdn(q.Name), q.Qtype)
				msg.RecursionDesired = true

				rr, _, err := c.Exchange(msg, net.JoinHostPort(config.Servers[0], config.Port))
				if rr == nil {
					log.Printf("*** error: %s\n", err.Error())
					return
				}

				if rr.Rcode != dns.RcodeSuccess {
					log.Printf("*** Invalid Answer name %s ofter %d query for %s\n", q.Name, q.Qtype, q.Name)
					return
				}

				for _, a := range rr.Answer {
					m.Answer = append(m.Answer, a)
				}
			}
		default:
			log.Fatalf("NoImpleted Type %d\n", q.Qtype)
		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func main() {
	dns.HandleFunc(".", handleDnsRequest)

	port := 53
	server := &dns.Server{Addr: "0.0.0.0: " + strconv.Itoa(port), Net: "udp"}

	log.Printf("Starting at %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

}
