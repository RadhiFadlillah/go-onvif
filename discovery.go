package onvif

import (
	"errors"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/clbanning/mxj"
	"github.com/satori/go.uuid"
)

var errWrongDiscoveryResponse = errors.New("Response is not related to discovery request")

// StartDiscovery send a WS-Discovery message and wait for all matching device to respond
func StartDiscovery(duration time.Duration) ([]Device, error) {
	// Get list of interface address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []Device{}, err
	}

	// Fetch IPv4 address
	ipAddrs := []string{}
	for _, addr := range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if ok && !ipAddr.IP.IsLoopback() && ipAddr.IP.To4() != nil {
			ipAddrs = append(ipAddrs, ipAddr.IP.String())
		}
	}

	// Create initial discovery results
	discoveryResults := []Device{}

	// Discover device on each interface's network
	for _, ipAddr := range ipAddrs {
		devices, err := discoverDevices(ipAddr, duration)
		if err != nil {
			return []Device{}, err
		}

		discoveryResults = append(discoveryResults, devices...)
	}

	return discoveryResults, nil
}

func discoverDevices(ipAddr string, duration time.Duration) ([]Device, error) {
	// Create WS-Discovery request
	newUuid, err := uuid.NewV4()
	if err != nil {
		return []Device{}, err
	}
	requestID := "uuid:" + newUuid.String()

	request := `		
		<?xml version="1.0" encoding="UTF-8"?>
		<e:Envelope
		    xmlns:e="http://www.w3.org/2003/05/soap-envelope"
		    xmlns:w="http://schemas.xmlsoap.org/ws/2004/08/addressing"
		    xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery"
		    xmlns:dn="http://www.onvif.org/ver10/network/wsdl">
		    <e:Header>
		        <w:MessageID>` + requestID + `</w:MessageID>
		        <w:To e:mustUnderstand="true">urn:schemas-xmlsoap-org:ws:2005:04:discovery</w:To>
		        <w:Action a:mustUnderstand="true">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe
		        </w:Action>
		    </e:Header>
		    <e:Body>
		        <d:Probe>
		            <d:Types>dn:NetworkVideoTransmitter</d:Types>
		        </d:Probe>
		    </e:Body>
		</e:Envelope>`

	// Clean WS-Discovery message
	request = regexp.MustCompile(`\>\s+\<`).ReplaceAllString(request, "><")
	request = regexp.MustCompile(`\s+`).ReplaceAllString(request, " ")

	// Create UDP address for local and multicast address
	localAddress, err := net.ResolveUDPAddr("udp4", ipAddr+":0")
	if err != nil {
		return []Device{}, err
	}

	multicastAddress, err := net.ResolveUDPAddr("udp4", "239.255.255.250:3702")
	if err != nil {
		return []Device{}, err
	}

	// Create UDP connection to listen for respond from matching device
	conn, err := net.ListenUDP("udp", localAddress)
	if err != nil {
		return []Device{}, err
	}
	defer conn.Close()

	// Set connection's timeout
	err = conn.SetDeadline(time.Now().Add(duration))
	if err != nil {
		return []Device{}, err
	}

	// Send WS-Discovery request to multicast address
	_, err = conn.WriteToUDP([]byte(request), multicastAddress)
	if err != nil {
		return []Device{}, err
	}

	// Create initial discovery results
	discoveryResults := []Device{}

	// Keep reading UDP message until timeout
	for {
		// Create buffer and receive UDP response
		buffer := make([]byte, 10*1024)
		_, _, err = conn.ReadFromUDP(buffer)

		// Check if connection timeout
		if err != nil {
			if udpErr, ok := err.(net.Error); ok && udpErr.Timeout() {
				break
			} else {
				return discoveryResults, err
			}
		}

		// Read and parse WS-Discovery response
		device, err := readDiscoveryResponse(requestID, buffer)
		if err != nil && err != errWrongDiscoveryResponse {
			return discoveryResults, err
		}

		// Push device to results
		discoveryResults = append(discoveryResults, device)
	}

	return discoveryResults, nil
}

// readDiscoveryResponse reads and parses WS-Discovery response
func readDiscoveryResponse(messageID string, buffer []byte) (Device, error) {
	// Inital result
	result := Device{}

	// Parse XML to map
	mapXML, err := mxj.NewMapXml(buffer)
	if err != nil {
		return result, err
	}

	// Check if this response is for our request
	responseMessageID, _ := mapXML.ValueForPathString("Envelope.Header.RelatesTo")
	if responseMessageID != messageID {
		return result, errWrongDiscoveryResponse
	}

	// Get device's ID and clean it
	deviceID, _ := mapXML.ValueForPathString("Envelope.Body.ProbeMatches.ProbeMatch.EndpointReference.Address")
	deviceID = strings.Replace(deviceID, "urn:uuid:", "", 1)

	// Get device's name
	deviceName := ""
	scopes, _ := mapXML.ValueForPathString("Envelope.Body.ProbeMatches.ProbeMatch.Scopes")
	for _, scope := range strings.Split(scopes, " ") {
		if strings.HasPrefix(scope, "onvif://www.onvif.org/name/") {
			deviceName = strings.Replace(scope, "onvif://www.onvif.org/name/", "", 1)
			deviceName = strings.Replace(deviceName, "_", " ", -1)
			break
		}
	}

	// Get device's xAddrs
	xAddrs, _ := mapXML.ValueForPathString("Envelope.Body.ProbeMatches.ProbeMatch.XAddrs")
	listXAddr := strings.Split(xAddrs, " ")
	if len(listXAddr) == 0 {
		return result, errors.New("Device does not have any xAddr")
	}

	// Finalize result
	result.ID = deviceID
	result.Name = deviceName
	result.XAddr = listXAddr[0]

	return result, nil
}
