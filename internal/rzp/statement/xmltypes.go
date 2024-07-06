package statement

import "encoding/xml"

type Listiny struct {
	XMLName      xml.Name `xml:"listiny"`
	Version      string   `xml:"version,attr"`
	Namespace    string   `xml:"xmlns,attr"`
	OsvedceniMPO string   `xml:"OsvedceniMPO"`
	Verweb       Verweb   `xml:"verweb"`
}

type Verweb struct {
	Hlavicka         Hlavicka         `xml:"Hlavicka"`
	PodnikatelDetail PodnikatelDetail `xml:"PodnikatelDetail"`
	InfoText         string           `xml:"InfoText"`
}

type Hlavicka struct {
	Nadpis       string       `xml:"Nadpis,attr"`
	CasVytvoreni CasVytvoreni `xml:"CasVytvoreni"`
}

type CasVytvoreni struct {
	Popis   string `xml:"Popis,attr"`
	Hodnota string `xml:",chardata"`
}

type PodnikatelDetail struct {
	PodnikatelOsoba    PodnikatelOsoba    `xml:"PodnikatelOsoba"`
	ObchodniJmeno      string             `xml:"ObchodniJmeno"`
	AdresaPodnikani    AdresaPodnikani    `xml:"AdresaPodnikani"`
	IdentifikacniCislo IdentifikacniCislo `xml:"IdentifikacniCislo"`
	SeznamZivnosti     SeznamZivnosti     `xml:"SeznamZivnosti"`
	EvidujiciUrad      string             `xml:"EvidujiciUrad"`
}

type PodnikatelOsoba struct {
	ZucastnenaOsobaDetail ZucastnenaOsobaDetail `xml:"ZucastnenaOsobaDetail"`
}

type ZucastnenaOsobaDetail struct {
	OsobaPoradoveCislo string          `xml:"OsobaPoradoveCislo"`
	JmenoPrijmeni      HodnotaWithDesc `xml:"JmenoPrijmeni"`
	DatumNarozeni      HodnotaWithDesc `xml:"DatumNarozeni"`
	Obcanstvi          HodnotaWithDesc `xml:"Obcanstvi"`
	TitulPredJmenem    Hodnota         `xml:"TitulPredJmenem"`
	Jmeno              HodnotaWithDesc `xml:"Jmeno"`
	Prijmeni           HodnotaWithDesc `xml:"Prijmeni"`
	TitulZaJmenem      Hodnota         `xml:"TitulZaJmenem"`
}

type AdresaPodnikani struct {
	Popis          string         `xml:"Popis,attr"`
	PlatnostAdresy PlatnostAdresy `xml:"PlatnostAdresy"`
}

type PlatnostAdresy struct {
	ZmenaAdresy ZmenaAdresy `xml:"ZmenaAdresy"`
}

type ZmenaAdresy struct {
	TextAdresy string `xml:"TextAdresy"`
}

type IdentifikacniCislo struct {
	Popis           string          `xml:"Popis,attr"`
	PlatnostHodnoty PlatnostHodnoty `xml:"PlatnostHodnoty"`
}

type PlatnostHodnoty struct {
	Hodnota string `xml:"Hodnota"`
}

type SeznamZivnosti struct {
	Popis   string    `xml:"Popis,attr"`
	Zivnost []Zivnost `xml:"Zivnost"`
}

type Zivnost struct {
	Popis                string          `xml:"Popis,attr"`
	Predmet              HodnotaWithDesc `xml:"Predmet"`
	Obor                 Obor            `xml:"Obor"`
	Druh                 HodnotaWithDesc `xml:"Druh"`
	Vznik                string          `xml:"Vznik"`
	PlatnostOpravneni    HodnotaWithDesc `xml:"PlatnostOpravneni"`
	ZivnostPoradoveCislo string          `xml:"ZivnostPoradoveCislo"`
}

type Obor struct {
	Popis string `xml:"Popis,attr"`
	Vycet Vycet  `xml:"Vycet"`
}

type Vycet struct {
	Drive []Hodnota `xml:"Drive"`
}

type HodnotaWithDesc struct {
	Popis   string `xml:"Popis,attr,omitempty"`
	Hodnota string `xml:"Hodnota"`
}

type Hodnota struct {
	Popis   string `xml:"Popis,attr,omitempty"`
	Hodnota string `xml:"Hodnota"`
}
