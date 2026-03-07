package xmpp

import (
	"encoding/xml"

	"gosrc.io/xmpp/stanza"
)

const (
	nsHTTPUpload = "urn:xmpp:http:upload:0"
)

type UploadRequest struct {
	XMLName     xml.Name `xml:"urn:xmpp:http:upload:0 request"`
	Filename    string   `xml:"filename,attr"`
	Size        int64    `xml:"size,attr"`
	ContentType string   `xml:"content-type,attr,omitempty"`
}

func (u UploadRequest) Namespace() string {
	return nsHTTPUpload
}

func (u UploadRequest) GetSet() *stanza.ResultSet {
	return nil
}

type UploadSlotResponse struct {
	XMLName xml.Name `xml:"slot"`
	Put     Put      `xml:"put"`
	Get     Get      `xml:"get"`
}

type Put struct {
	XMLName xml.Name `xml:"put"`
	URL     string   `xml:"url,attr"`
}

type Get struct {
	XMLName xml.Name `xml:"get"`
	URL     string   `xml:"url,attr"`
}

func (u UploadSlotResponse) Namespace() string {
	return nsHTTPUpload
}

func (u UploadSlotResponse) GetSet() *stanza.ResultSet {
	return nil
}

func init() {
	stanza.TypeRegistry.MapExtension(stanza.PKTIQ, xml.Name{Space: nsHTTPUpload, Local: "request"}, UploadRequest{})
	stanza.TypeRegistry.MapExtension(stanza.PKTIQ, xml.Name{Space: nsHTTPUpload, Local: "slot"}, UploadSlotResponse{})
}
