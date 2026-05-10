package main

import (
	"bytes"
	"dns-go/models"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const PORT = 8282
const (
	TypeA             uint16 = 1
	ClassINET         uint16 = 1
	FlagResponse      uint16 = 1 << 15 // move by 15 so one sits at 16 -> unsigned 16 bit integer
	UDPMaxMessageSize uint   = 512
)

func dbLookup(queryResourceRecord models.DNSResourceRecord) ([]models.DNSResourceRecord, []models.DNSResourceRecord, []models.DNSResourceRecord) {

	var ansResourceRecord []models.DNSResourceRecord
	var authResourceRecord []models.DNSResourceRecord
	var addResourceRecord []models.DNSResourceRecord

	names, err := GetNames()
	if err != nil {
		log.Println("Error at lookup:", err.Error())
	}

	if queryResourceRecord.Type != TypeA || queryResourceRecord.Class != ClassINET {
		return ansResourceRecord, authResourceRecord, addResourceRecord
	}

	for _, name := range names {
		if strings.Contains(queryResourceRecord.DomainName, name.Name) {
			ansResourceRecord = append(ansResourceRecord, models.DNSResourceRecord{
				DomainName:         name.Name,
				Type:               TypeA,
				Class:              ClassINET,
				TimeToLive:         31337,
				ResourceData:       net.ParseIP(name.Address.String()).To4(),
				ResourceDataLength: 4,
			})
		}
	}
	return ansResourceRecord, authResourceRecord, addResourceRecord

}

func readDomainName(reqBuffer *bytes.Buffer) (string, error) {
	var domainName string
	b, err := reqBuffer.ReadByte()

	for ; b != 0 && err == nil; b, err = reqBuffer.ReadByte() {
		partLength := int(b)
		partByte := reqBuffer.Next(partLength)
		PartName := string(partByte)

		if len(domainName) == 0 {
			domainName = PartName
		} else {
			domainName += "." + PartName
		}
	}

	return domainName, err
}

func writeDomainName(resBuffer *bytes.Buffer, domain string) error {
	parts := strings.Split(domain, ".")

	for _, part := range parts {
		partLength := len(part)
		resBuffer.WriteByte(byte(partLength))
		resBuffer.Write([]byte(part))
	}

	err := resBuffer.WriteByte(0)

	return err
}

func handleDNSClient(requestBytes []byte, serverConn *net.UDPConn, clientAddr *net.UDPAddr) {

	//Read Request
	var reqBuffer = bytes.NewBuffer(requestBytes)
	var queryHeader models.DNSHeader

	err := binary.Read(reqBuffer, binary.BigEndian, &queryHeader)

	if err != nil {
		log.Println("Error Decoding Header: ", err.Error())
	}

	queryResourceRecords := make([]models.DNSResourceRecord, queryHeader.NumQuestions)

	for idx := range queryResourceRecords {
		queryResourceRecords[idx].DomainName, err = readDomainName(reqBuffer)
		if err != nil {
			log.Println("Error reading domain:", err.Error())
		}
		queryResourceRecords[idx].Type = binary.BigEndian.Uint16(reqBuffer.Next(2))
		queryResourceRecords[idx].Class = binary.BigEndian.Uint16(reqBuffer.Next(2))
	}

	// Lookup Domain
	var answerResourceRecords = make([]models.DNSResourceRecord, 0)
	var authorityResourceRecords = make([]models.DNSResourceRecord, 0)
	var additionalResourceRecords = make([]models.DNSResourceRecord, 0)

	for _, queryResourceRecord := range queryResourceRecords {
		newAnswerRR, newAuthorityRR, newAdditionalRR := dbLookup(queryResourceRecord)

		answerResourceRecords = append(answerResourceRecords, newAnswerRR...)
		authorityResourceRecords = append(authorityResourceRecords, newAuthorityRR...)
		additionalResourceRecords = append(additionalResourceRecords, newAdditionalRR...)

	}

	//Write response
	var resBuffer = new(bytes.Buffer)
	var resHeader models.DNSHeader

	resHeader = models.DNSHeader{
		TransactionID:  queryHeader.TransactionID,
		Flags:          queryHeader.Flags,
		NumQuestions:   queryHeader.NumQuestions,
		NumAnswers:     uint16((len(answerResourceRecords))),
		NumAuthorities: uint16((len(authorityResourceRecords))),
		NumAdditionals: uint16((len(additionalResourceRecords))),
	}

	err = Write(resBuffer, resHeader)

	for _, queryResourceRecord := range queryResourceRecords {
		err = writeDomainName(resBuffer, queryResourceRecord.DomainName)
		if err != nil {
			log.Println("Error Writing Domain name to buffer(QueryResourceRecord): ", err)
		}

		Write(resBuffer, queryResourceRecord.Type)
		Write(resBuffer, queryResourceRecord.Class)
	}

	for _, answerResourceRecord := range answerResourceRecords {
		err = writeDomainName(resBuffer, answerResourceRecord.DomainName)
		if err != nil {
			log.Println("Error Writing Domain name to buffer (AnswerResourceRecord): ", err)
		}
		Write(resBuffer, answerResourceRecord.Type)
		Write(resBuffer, answerResourceRecord.Class)
		Write(resBuffer, answerResourceRecord.TimeToLive)
		Write(resBuffer, answerResourceRecord.ResourceDataLength)
		Write(resBuffer, answerResourceRecord.ResourceData)
	}

	for _, authorityResourceRecord := range authorityResourceRecords {
		err = writeDomainName(resBuffer, authorityResourceRecord.DomainName)
		if err != nil {
			log.Println("Error Writing Domain name to buffer (AnswerResourceRecord): ", err)
		}
		Write(resBuffer, authorityResourceRecord.Type)
		Write(resBuffer, authorityResourceRecord.Class)
		Write(resBuffer, authorityResourceRecord.TimeToLive)
		Write(resBuffer, authorityResourceRecord.ResourceDataLength)
		Write(resBuffer, authorityResourceRecord.ResourceData)
	}

	for _, additionalResourceRecord := range additionalResourceRecords {
		err = writeDomainName(resBuffer, additionalResourceRecord.DomainName)
		if err != nil {
			log.Println("Error Writing Domain name to buffer (AnswerResourceRecord): ", err)
		}
		Write(resBuffer, additionalResourceRecord.Type)
		Write(resBuffer, additionalResourceRecord.Class)
		Write(resBuffer, additionalResourceRecord.TimeToLive)
		Write(resBuffer, additionalResourceRecord.ResourceDataLength)
		Write(resBuffer, additionalResourceRecord.ResourceData)
	}
	if len(answerResourceRecords) > 0 {
		log.Println("Answer:", answerResourceRecords[0].DomainName, answerResourceRecords[0].ResourceData)
	} else {
		log.Println("No answer records for this query", queryResourceRecords)
	}
	serverConn.WriteToUDP(resBuffer.Bytes(), clientAddr)

}

// util
func Write(w io.Writer, data any) error {
	return binary.Write(w, binary.BigEndian, data)
}

func main() {
	udpAdd, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Println("Failed to Resolve UDP addresss: ", err.Error())
		os.Exit(1)
	}

	udpConn, err := net.ListenUDP("udp", udpAdd)
	if err != nil {
		log.Println("Failed to Listen on UDP: ", err.Error())
		os.Exit(1)
	}

	log.Println("UDP server running on", PORT)
	defer udpConn.Close()

}
