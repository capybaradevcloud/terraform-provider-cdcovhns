package api

import (
	"context"
	"time"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/ovh/go-ovh/ovh"
)

type OvhAuthCurrentCredential struct {
	OvhSupport    bool             `json:"ovhSupport"`
	Status        string           `json:"status"`
	ApplicationId int64            `json:"applicationId"`
	CredentialId  int64            `json:"credentialId"`
	Rules         []ovh.AccessRule `json:"rules"`
	Expiration    time.Time        `json:"expiration"`
	LastUse       time.Time        `json:"lastUse"`
	Creation      time.Time        `json:"creation"`
}

type OVHCredentials struct {
	Endpoint          string
	ApplicationKey    string
	ApplicationSecret string
	ConsumerKey       string
}

type APIClient struct {
	Client *ovh.Client
	ctx    context.Context
}

func GetClient(data OVHCredentials, ctx context.Context) (*APIClient, error) {
	var cred OvhAuthCurrentCredential

	client, err := ovh.NewClient(
		data.Endpoint,
		data.ApplicationKey,
		data.ApplicationSecret,
		data.ConsumerKey,
	)
	if err != nil {
		return nil, err
	}

	// TODO: add terraform version
	client.UserAgent = "Terraform"

	if client.Client.Transport == nil {
		client.Client.Transport = cleanhttp.DefaultTransport()
	}

	if err := client.Get("/auth/currentCredential", &cred); err != nil {
		return nil, err
	}

	return &APIClient{
		Client: client,
		ctx:    ctx,
	}, nil
}
