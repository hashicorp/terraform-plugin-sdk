package validation

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// SingleIP returns a SchemaValidateFunc which tests if the provided value
// is of type string, and in valid single Value notation
func SingleIP() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		ip := net.ParseIP(v)
		if ip == nil {
			es = append(es, fmt.Errorf("expected %s to contain a valid IP, got: %s", k, v))
		}
		return
	}
}

func IPv6Address(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	ip := net.ParseIP(v)
	if six := ip.To16(); six == nil {
		errors = append(errors, fmt.Errorf("expected %s to contain a valid IPv6 address, got: %s", k, v))
	}

	return warnings, errors
}

func IPv4Address(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	ip := net.ParseIP(v)
	if four := ip.To4(); four == nil {
		errors = append(errors, fmt.Errorf("expected %s to contain a valid IPv4 address, got: %s", k, v))
	}

	return warnings, errors
}

// IPRange returns a SchemaValidateFunc which tests if the provided value
// is of type string, and in valid Value range notation
func IPRange() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		ips := strings.Split(v, "-")
		if len(ips) != 2 {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP range, got: %s", k, v))
			return
		}
		ip1 := net.ParseIP(ips[0])
		ip2 := net.ParseIP(ips[1])
		if ip1 == nil || ip2 == nil || bytes.Compare(ip1, ip2) > 0 {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP range, got: %s", k, v))
		}
		return
	}
}

func CIDR(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	_, _, err := net.ParseCIDR(v)
	if err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid IPv4 Value, got %v: %v", k, i, err))
	}

	return warnings, errors
}

// CIDRNetwork returns a SchemaValidateFunc which tests if the provided value
// is of type string, is in valid Value network notation, and has significant bits between min and max (inclusive)
func CIDRNetwork(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		_, ipnet, err := net.ParseCIDR(v)
		if err != nil {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid Value, got: %s with err: %s", k, v, err))
			return
		}

		if ipnet == nil || v != ipnet.String() {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid network Value, expected %s, got %s",
				k, ipnet, v))
		}

		sigbits, _ := ipnet.Mask.Size()
		if sigbits < min || sigbits > max {
			es = append(es, fmt.Errorf(
				"expected %q to contain a network Value with between %d and %d significant bits, got: %d",
				k, min, max, sigbits))
		}

		return
	}
}

func MACAddress(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return
	}

	if _, err := net.ParseMAC(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a valid MAC address, got %v: %v", k, i, err))
	}

	return warnings, errors
}

func PortNumber(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(int)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be int", k))
		return
	}

	if v < 1 || 65535 < v {
		errors = append(errors, fmt.Errorf("expected %q to be a valid port number, got: %v", k, v))
	}

	return warnings, errors
}

func PortNumberOrZero(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(int)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be int", k))
		return
	}

	if v < 0 || 65535 < v {
		errors = append(errors, fmt.Errorf("expected %q to be a valid port number or 0, got: %v", k, v))
	}

	return warnings, errors
}
