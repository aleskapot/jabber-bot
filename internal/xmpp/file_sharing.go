package xmpp

import "encoding/xml"

type FileSharing struct {
	XMLName     xml.Name      `xml:"urn:xmpp:sfs:0 file-sharing"`
	Disposition string        `xml:"disposition,attr,omitempty"`
	ID          string        `xml:"id,attr,omitempty"`
	File        *FileMetadata `xml:"file,omitempty"`
	Sources     *FileSources  `xml:"sources,omitempty"`
}

type FileMetadata struct {
	XMLName   xml.Name   `xml:"urn:xmpp:file:metadata:0 file"`
	MediaType string     `xml:"media-type,omitempty"`
	Name      string     `xml:"name,omitempty"`
	Size      int64      `xml:"size,omitempty"`
	Hashes    []FileHash `xml:"hash,omitempty"`
	Desc      string     `xml:"desc,omitempty"`
}

type FileHash struct {
	XMLName xml.Name `xml:"urn:xmpp:hashes:2 hash"`
	Algo    string   `xml:"algo,attr"`
	Value   string   `xml:",innerxml"`
}

type FileSources struct {
	XMLName    xml.Name        `xml:"sources"`
	URLSources []URLDataSource `xml:"url-data,omitempty"`
}

type URLDataSource struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/url-data url-data"`
	Target  string   `xml:"target,attr"`
}

type Fallback struct {
	XMLName xml.Name `xml:"urn:xmpp:fallback:0 fallback"`
	For     string   `xml:"for,attr"`
	Body    string   `xml:"body"`
}
