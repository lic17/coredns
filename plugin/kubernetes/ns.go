package kubernetes

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

func isDefaultNS(name, zone string) bool {
	return strings.Index(name, defaultNSName) == 0 && strings.Index(name, zone) == len(defaultNSName)
}

// nsAddrs returns the A or AAAA records for the CoreDNS service in the cluster. If the service cannot be found,
// it returns a record for the local address of the machine we're running on.
func (k *Kubernetes) nsAddrs(external bool, zone string) []dns.RR {
	var (
		//svcNames []string
		//svcIPs   []net.IP
		svcIPNum = 0
		rrs      []dns.RR
	)

	// Find the CoreDNS Endpoints
	for _, localIP := range k.localIPs {
		endpoints := k.APIConn.EpIndexReverse(localIP.String())

		// Collect IPs for all Services of the Endpoints
		for _, endpoint := range endpoints {
			svcs := k.APIConn.SvcIndex(endpoint.Index)
			for _, svc := range svcs {
				if external {
					svcName := strings.Join([]string{svc.Name, svc.Namespace, zone}, ".")
					for _, exIP := range svc.ExternalIPs {
						//svcNames = append(svcNames, svcName)
						//svcIPs = append(svcIPs, net.ParseIP(exIP))
						svcIPNum++
						ip := net.ParseIP(exIP)
						rr := createRR(svcName, ip)
						rrs = append(rrs, rr)

					}
					continue
				}
				svcName := strings.Join([]string{svc.Name, svc.Namespace, Svc, zone}, ".")
				if svc.Headless() {
					// For a headless service, use the endpoints IPs
					for _, s := range endpoint.Subsets {
						for _, a := range s.Addresses {
							//svcNames = append(svcNames, endpointHostname(a, k.endpointNameMode)+"."+svcName)
							//svcIPs = append(svcIPs, net.ParseIP(a.IP))
							svcIPNum++
							ip := net.ParseIP(a.IP)
							rr := createRR(endpointHostname(a, k.endpointNameMode)+"."+svcName, ip)
							rrs = append(rrs, rr)

							rr = createRR(svc.ExternalName, ip)
							rrs = append(rrs, rr)

						}
					}
				} else {
					for _, clusterIP := range svc.ClusterIPs {
						//svcNames = append(svcNames, svcName)
						//svcIPs = append(svcIPs, net.ParseIP(clusterIP))
						svcIPNum++
						ip := net.ParseIP(clusterIP)
						rr := createRR(svcName, ip)
						rrs = append(rrs, rr)

						rr = createRR(svc.ExternalName, ip)
						rrs = append(rrs, rr)

					}
				}
			}
		}
	}

	// If no local IPs matched any endpoints, use the localIPs directly
	if svcIPNum == 0 {
		//svcIPs = make([]net.IP, len(k.localIPs))
		//svcNames = make([]string, len(k.localIPs))
		for _, localIP := range k.localIPs {
			//svcNames[i] = defaultNSName + zone
			//svcIPs[i] = localIP

			rr := createRR(defaultNSName+zone, localIP)
			rrs = append(rrs, rr)

		}
	}
	/*
		// Create an RR slice of collected IPs
		rrs := make([]dns.RR, len(svcIPs))
		for i, ip := range svcIPs {
			if ip.To4() == nil {
				rr := new(dns.AAAA)
				rr.Hdr.Class = dns.ClassINET
				rr.Hdr.Rrtype = dns.TypeAAAA
				rr.Hdr.Name = svcNames[i]
				rr.AAAA = ip
				rrs[i] = rr
				continue
			}
			rr := new(dns.A)
			rr.Hdr.Class = dns.ClassINET
			rr.Hdr.Rrtype = dns.TypeA
			rr.Hdr.Name = svcNames[i]
			rr.A = ip
			rrs[i] = rr
		}
	*/
	return rrs
}

func createRR(name string, ip net.IP) dns.RR {
	fmt.Println(name, ip)
	if ip.To4() == nil {
		rr := new(dns.AAAA)
		rr.Hdr.Class = dns.ClassINET
		rr.Hdr.Rrtype = dns.TypeAAAA
		rr.Hdr.Name = name
		rr.AAAA = ip
		return rr
	}

	rr := new(dns.A)
	rr.Hdr.Class = dns.ClassINET
	rr.Hdr.Rrtype = dns.TypeA
	rr.Hdr.Name = name
	rr.A = ip
	return rr
}

const defaultNSName = "ns.dns."
