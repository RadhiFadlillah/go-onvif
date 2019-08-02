package onvif

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/clbanning/mxj"
	"github.com/satori/go.uuid"
)

var httpClient = &http.Client{Timeout: time.Second * 5}

// SOAP contains data for SOAP request
type SOAP struct {
	Body     string
	XMLNs    []string
	User     string
	Password string
	TokenAge time.Duration
}

// SendRequest sends SOAP request to xAddr
func (soap SOAP) SendRequest(xaddr string) (mxj.Map, error) {
	// Create SOAP request
	request := soap.createRequest()

	// Make sure URL valid and add authentication in xAddr
	urlXAddr, err := url.Parse(xaddr)
	if err != nil {
		return nil, err
	}

	if soap.User != "" {
		urlXAddr.User = url.UserPassword(soap.User, soap.Password)
	}

	// Create HTTP request
	buffer := bytes.NewBuffer([]byte(request))
	req, err := http.NewRequest("POST", urlXAddr.String(), buffer)
	req.Header.Set("Content-Type", "application/soap+xml")
	req.Header.Set("Charset", "utf-8")

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse XML to map
	mapXML, err := mxj.NewMapXml(responseBody)
	if err != nil {
		return nil, err
	}

	// Check if SOAP returns fault
	fault, _ := mapXML.ValueForPathString("Envelope.Body.Fault.Reason.Text.#text")
	if fault != "" {
		return nil, errors.New(fault)
	}

	return mapXML, nil
}

func (soap SOAP) createRequest() string {
	// Create request envelope
	request := `<?xml version="1.0" encoding="UTF-8"?>`
	request += `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"`

	// Set XML namespace
	for _, namespace := range soap.XMLNs {
		request += " " + namespace
	}
	request += ">"

	// Set request header
	if soap.User != "" {
		soapHeader, err := soap.createUserToken()
		if err != nil {
			request += "<s:Header />"
			soap.Body = `
				<s:Fault>
				<s:Code>
					<s:Value>s:Receiver</s:Value>
				</s:Code>
				<s:Reason>
					<s:Text xml:lang="en-US">Error creating user token</s:Text>
				</s:Reason>
				</s:Fault>`
		} else {
			request += "<s:Header>" + soapHeader + "</s:Header>"
		}
	}

	// Set request body
	request += "<s:Body>" + soap.Body + "</s:Body>"

	// Close request envelope
	request += "</s:Envelope>"

	// Clean request
	request = regexp.MustCompile(`\>\s+\<`).ReplaceAllString(request, "><")
	request = regexp.MustCompile(`\s+`).ReplaceAllString(request, " ")

	return request
}

func (soap SOAP) createUserToken() (string, error) {
	newUuid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	nonce := newUuid.Bytes()

	nonce64 := base64.StdEncoding.EncodeToString(nonce)
	timestamp := time.Now().Add(soap.TokenAge).UTC().Format(time.RFC3339)
	token := string(nonce) + timestamp + soap.Password

	sha := sha1.New()
	sha.Write([]byte(token))
	shaToken := sha.Sum(nil)
	shaDigest64 := base64.StdEncoding.EncodeToString(shaToken)

	return `<Security s:mustUnderstand="1" xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
  		<UsernameToken>
    		<Username>` + soap.User + `</Username>
    		<Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">` + shaDigest64 + `</Password>
    		<Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">` + nonce64 + `</Nonce>
    		<Created xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">` + timestamp + `</Created>
		</UsernameToken>
	</Security>`, nil
}
