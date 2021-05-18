package connect

import (
	_ "embed" //golint
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

var (
	//go:embed status-text.tmpl
	statusTemplate string
)

// Status is used to create JSON output
type Status struct {
	Summary    string `json:"-"`
	Identifier string `json:"identifier"`
	Version    string `json:"version"`
	Arch       string `json:"arch"`
	Status     string `json:"status"`
	RegCode    string `json:"regcode,omitempty"`
	StartsAt   string `json:"starts_at,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	SubStatus  string `json:"subscription_status,omitempty"`
	Type       string `json:"type,omitempty"`
}

// GetProductStatuses returns statuses of installed products
func GetProductStatuses(format string) string {
	statuses, err := getStatuses()
	if err != nil {
		Error.Println(err)
		return fmt.Sprintf("ERROR: %s", err)
	}
	if format == "json" {
		jsonStr, err := json.Marshal(statuses)
		if err != nil {
			Error.Println(err)
			return fmt.Sprintf("ERROR: %s", err)
		}
		return string(jsonStr)
	}
	if format == "text" {
		text, err := getStatusText(statuses)
		if err != nil {
			Error.Println(err)
			return fmt.Sprintf("ERROR: %s", err)
		}
		return text
	}
	panic("Parameter must be \"json\" or \"text\"")
}

func getStatuses() ([]Status, error) {
	var statuses []Status
	products, err := GetInstalledProducts()
	if err != nil {
		return statuses, err
	}

	activations := []Activation{}
	if CredentialsExists() {
		activations, err = GetActivations()
		if err != nil {
			return statuses, err
		}
	}

	activationMap := make(map[string]Activation)
	for _, activation := range activations {
		activationMap[activation.ToTriplet()] = activation
	}

	for _, product := range products {
		status := Status{
			Summary:    product.Summary,
			Identifier: product.Name,
			Version:    product.Version,
			Arch:       product.Arch,
			Status:     "Not Registered",
		}
		key := product.ToTriplet()
		activation, inMap := activationMap[key]
		// TODO registered but not activated?
		if inMap && !activation.IsFree() {
			status.RegCode = activation.RegCode
			layout := "2006-01-02 15:04:05 MST"
			status.StartsAt = activation.StartsAt.Format(layout)
			status.ExpiresAt = activation.ExpiresAt.Format(layout)
			status.SubStatus = activation.Status
			status.Type = activation.Type
			status.Status = "Registered"
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func getStatusText(statuses []Status) (string, error) {
	t, err := template.New("status-text").Parse(statusTemplate)
	if err != nil {
		return "", err
	}
	var outWriter strings.Builder
	err = t.Execute(&outWriter, statuses)
	if err != nil {
		return "", err
	}
	return outWriter.String(), nil
}
