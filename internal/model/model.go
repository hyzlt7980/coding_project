package model

import "time"

type ServiceRecord struct {
	InstanceName string            `json:"instance_name"`
	ServiceType  string            `json:"service_type"`
	ServiceShort string            `json:"service_short"`
	HostName     string            `json:"hostname,omitempty"`
	Port         int               `json:"port,omitempty"`
	TTL          uint32            `json:"ttl,omitempty"`
	IPv4         []string          `json:"ipv4,omitempty"`
	IPv6         []string          `json:"ipv6,omitempty"`
	TXT          map[string]string `json:"txt,omitempty"`
	TXTOrder     []string          `json:"txt_order,omitempty"`
	RawTXT       []string          `json:"raw_txt,omitempty"`
	Banner       map[string]string `json:"banner,omitempty"`
}

type DiscoveryResult struct {
	Services   []ServiceRecord `json:"services"`
	PTRAnswers []string        `json:"ptr_answers"`
	Wait       time.Duration   `json:"wait"`
}
