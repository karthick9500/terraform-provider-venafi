package venafi

import (
	"fmt"
	"github.com/Venafi/vcert"
	"github.com/Venafi/vcert/pkg/endpoint"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
)

const (
	messageVenafiPingFailed       = "Failed to ping Venafi endpoint: "
	messageVenafiPingSucessfull   = "Venafi ping sucessfull"
	messageVenafiClientInitFailed = "Failed to initialize Venafi client: "
	messageVenafiConfigFailed     = "Failed to build config for Venafi issuer: "
	messageUseFakemode            = "Using fake mode to issue certificate"
	messageUseCloud               = "Using Cloud to issue certificate"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	//TODO: provide backwards compability
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VENAFI_URL", nil),
				Description: `The Venafi Web Service URL.. Example: https://tpp.venafi.example/vedsdk`,
			},

			"zone": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VENAFI_ZONE", "Default"),
				Description: `Name of Venafi Platfrom or Cloud policy. 
Example for Platform: testpolicy\\vault
Example for Venafi Cloud: Default`,
			},

			"tpp_username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VENAFI_USER", nil),
				Description: `web API user for Venafi Platfrom Example: admin`,
			},
			"tpp_password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VENAFI_PASS", nil),
				Description: `Password for web API user Example: password`,
			},
			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VENAFI_API", nil),
				Description: `API key for Venafi Cloud. Example: 142231b7-cvb0-412e-886b-6aeght0bc93d`,
			},
			"trust_bundle": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: `Use to specify a PEM formatted file with certificates to be used as trust anchors when communicating with the remote server.
Example:
  trust_bundle = "${file("chain.pem")}"`,
			},
			"fake_mode": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VENAFI_FAKEMODE", nil),
				Description: `Set it to true to use face CA instead of Cloud or Platform to issue certificates. Useful for testing.`,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"venafi_certificate": resourceVenafiCertificate(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	log.Printf("Configuring provider\n")
	apiKey := d.Get("api_key").(string)
	url := d.Get("url").(string)
	tppUser := d.Get("tpp_username").(string)
	tppPassword := d.Get("tpp_password").(string)
	zone := d.Get("zone").(string)
	fakeMode := d.Get("fake_mode").(bool)
	trustBundle := d.Get("trust_bundle").(string)

	var cfg vcert.Config

	if fakeMode {
		log.Printf(messageUseFakemode)
		cfg = vcert.Config{
			ConnectorType: endpoint.ConnectorTypeFake,
			LogVerbose:    true,
		}
	} else if tppUser != "" && tppPassword != "" {
		log.Printf("Using Platform with url %s to issue certificate\n", url)
		cfg = vcert.Config{
			ConnectorType: endpoint.ConnectorTypeTPP,
			BaseUrl:       url,
			Credentials: &endpoint.Authentication{
				User:     tppUser,
				Password: tppPassword,
			},
			Zone:       zone,
			LogVerbose: true,
		}
	} else if apiKey != "" {
		if url != "" {
			log.Println(messageUseCloud)
			cfg = vcert.Config{
				ConnectorType: endpoint.ConnectorTypeCloud,
				BaseUrl:       url,
				Credentials: &endpoint.Authentication{
					APIKey: apiKey,
				},
				Zone:       zone,
				LogVerbose: true,
			}
		} else {
			log.Println(messageUseCloud)
			cfg = vcert.Config{
				ConnectorType: endpoint.ConnectorTypeCloud,
				Credentials: &endpoint.Authentication{
					APIKey: apiKey,
				},
				Zone:       zone,
				LogVerbose: true,
			}
		}
	} else {
		return nil, fmt.Errorf(messageVenafiConfigFailed)
	}

	if trustBundle != "" {
		log.Printf("Importing trusted certificate: \n %s", trustBundle)
		cfg.ConnectionTrust = trustBundle
	}
	cl, err := vcert.NewClient(&cfg)
	if err != nil {
		log.Printf(messageVenafiClientInitFailed + err.Error())
		return nil, err
	}
	err = cl.Ping()
	if err != nil {
		log.Printf(messageVenafiPingFailed + err.Error())
		return nil, err
	}
	return &cfg, nil
}
