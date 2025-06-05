package main

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

type Member struct {
	ObjectType    string `xml:"objectType,attr"`
	XRoadInstance string `xml:"xRoadInstance"`
	MemberClass   string `xml:"memberClass"`
	MemberCode    string `xml:"memberCode"`
}

type ClientList struct {
	Members []struct {
		ID   Member `xml:"id"`
		Name string `xml:"name"`
	} `xml:"member"`
}

func main() {
	// Флаг для уровня логирования
	logLevel := flag.String("log-level", "info", "Log level: fatal, error, warn, info")
	flag.Parse()

	// Применяем уровень логирования
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v \nPossible levels: fatal, error, warn, info ", err)
		os.Exit(1)
	}
	log.SetLevel(level)

	log.Info("Starting healthcheck")
	// 1. Загрузка TLS ключа и сертификата
	cert, err := tls.LoadX509KeyPair("/etc/certs/tls.crt", "/etc/certs/tls.key")
	if err != nil {
		log.Fatalf("Failed to load cert/key: %v", err)
		os.Exit(1)
	}

	// 2. Настройка TLS клиента
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,                    // не проверяем self-signed сервер
		Certificates:       []tls.Certificate{cert}, // клиентский ключ+сертификат
	}

	// 1. GET запрос
	resp, err := http.Get("http://127.0.0.1/listSecurityServerClients")
	if err != nil {
		log.Fatalf("Request failed: %v\n", err)
		os.Exit(1)
	}
	log.Infof("Get request handled successful. Got info from http://127.0.0.1/listSecurityServerClients\n")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 2. Парсинг XML
	var list ClientList
	err = xml.Unmarshal(body, &list)
	if err != nil {
		log.Fatalf("Invalid XML: %v\n", err)
		os.Exit(1)
	}

	var member Member
	for _, m := range list.Members {
		if m.ID.ObjectType == "MEMBER" {
			member = m.ID
			break
		}
	}

	if member.MemberCode == "" {
		log.Fatal("No MEMBER object found.")
		os.Exit(1)
	}

	// 3. Построить SOAP XML
	soap := fmt.Sprintf(`<?xml version="1.0"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
  xmlns:xroad="http://x-road.eu/xsd/xroad.xsd"
  xmlns:om="http://x-road.eu/xsd/op-monitoring.xsd"
  xmlns:id="http://x-road.eu/xsd/identifiers">
  <SOAP-ENV:Header>
    <xroad:client id:objectType="MEMBER">
      <id:xRoadInstance>%s</id:xRoadInstance>
      <id:memberClass>%s</id:memberClass>
      <id:memberCode>%s</id:memberCode>
    </xroad:client>
    <xroad:service id:objectType="SERVICE">
      <id:xRoadInstance>%s</id:xRoadInstance>
      <id:memberClass>%s</id:memberClass>
      <id:memberCode>%s</id:memberCode>
      <id:serviceCode>getSecurityServerHealthData</id:serviceCode>
    </xroad:service>
    <xroad:id>probe-check</xroad:id>
    <xroad:protocolVersion>4.0</xroad:protocolVersion>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <om:getSecurityServerHealthData/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`, member.XRoadInstance, member.MemberClass, member.MemberCode,
		member.XRoadInstance, member.MemberClass, member.MemberCode)

	// 4. POST SOAP запрос
	log.Info("Parsing xml from http://127.0.0.1/listSecurityServerClients is OK.\n")
	log.Info("Sending SOAP request.\n")

	tls_client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	req, err := http.NewRequest("POST", "https://127.0.0.1/", bytes.NewBuffer([]byte(soap)))
	if err != nil {
		log.Fatalf("Failed to create request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "text/xml")

	resp2, err := tls_client.Do(req)
	if err != nil {
		log.Fatalf("SOAP request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()

	// 5. Читаем и ищем тег <om:getSecurityServerHealthDataResponse>
	body2, _ := io.ReadAll(resp2.Body)
	if bytes.Contains(body2, []byte("getSecurityServerHealthDataResponse")) {
		log.Info("Proxy is healthy.\n")
		os.Exit(0)
	} else {
		log.Fatal("No health response received.\n")
		os.Exit(1)
	}
}
