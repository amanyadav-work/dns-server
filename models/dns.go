package models

import "net"

//DNSHeader describes headers that are in req/res
type DNSHeader struct {
	TransactionID  uint16 // unsigned 16 bit integer, 8 bit = 8 byte so 16 bit = 2 bytes which allows reading 2 bytes from buffer into this field
	Flags          uint16
	NumQuestions   uint16
	NumAnswers     uint16
	NumAuthorities uint16
	NumAdditionals uint16
}

//DNSResourceRecord describes individual records in req and res of DNS payload body
type DNSResourceRecord struct {
	DomainName         string
	Type               uint16
	Class              uint16
	TimeToLive         uint32
	ResourceData       []byte
	ResourceDataLength uint16
}

type NameModel struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Name struct {
	Name    string `json:"name"`
	Address net.IP `json:"address"`
}
