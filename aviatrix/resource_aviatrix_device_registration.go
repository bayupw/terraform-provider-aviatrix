package aviatrix

import (
	"fmt"
	"log"
	"strconv"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v2/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAviatrixDeviceRegistration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAviatrixDeviceRegistrationCreate,
		Read:   resourceAviatrixDeviceRegistrationRead,
		Update: resourceAviatrixDeviceRegistrationUpdate,
		Delete: resourceAviatrixDeviceRegistrationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the device.",
			},
			"public_ip": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Public IP address of the device.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username to use to connect to the device.",
			},
			"key_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"password", "key_file"},
				Description:  "Path to private key file.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: envDefaultFunc("AVIATRIX_DEVICE_PASSWORD"),
				Description: "Password to connect to the device. " +
					"This attribute can also be set via environment variable 'AVIATRIX_DEVICE_PASSWORD'. " +
					"If both are set the value in the config file will be used.",
			},
			"host_os": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ios",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ios", "aviatrix"}, false),
				Description:  "Device host OS. Default value is 'ios'. Valid values are 'ios' or 'aviatrix'.",
			},
			"ssh_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     22,
				Description: "SSH port to use to connect to the device. Defaults to 22 if not set.",
			},
			"address_1": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Address line 1.",
			},
			"address_2": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Address line 2.",
			},
			"city": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "City",
			},
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "State",
			},
			"country": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ISO two-letter country code.",
			},
			"zip_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zip code.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description.",
			},
			"software_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "software_version can be used to set the desired software version of the CaaG. " +
					"If set, we will attempt to update the gateway to the specified version. " +
					"If left blank, the gateway software version will continue to be managed through the aviatrix_controller_config resource.",
			},
			"is_caag": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this device is a Managed CloudN device (CaaG)",
			},
		},
	}
}

// marshalDeviceRegistrationInput marshals the ResourceData into a Device struct.
func marshalDeviceRegistrationInput(d *schema.ResourceData) *goaviatrix.Device {
	return &goaviatrix.Device{
		Name:        d.Get("name").(string),
		PublicIP:    d.Get("public_ip").(string),
		Username:    d.Get("username").(string),
		KeyFile:     d.Get("key_file").(string),
		Password:    d.Get("password").(string),
		HostOS:      d.Get("host_os").(string),
		SshPort:     d.Get("ssh_port").(int),
		SshPortStr:  strconv.Itoa(d.Get("ssh_port").(int)),
		Address1:    d.Get("address_1").(string),
		Address2:    d.Get("address_2").(string),
		City:        d.Get("city").(string),
		State:       d.Get("state").(string),
		Country:     d.Get("country").(string),
		ZipCode:     d.Get("zip_code").(string),
		Description: d.Get("description").(string),
	}
}

func resourceAviatrixDeviceRegistrationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	device := marshalDeviceRegistrationInput(d)

	if err := client.RegisterDevice(device); err != nil {
		return fmt.Errorf("could not register device: %v", err)
	}

	d.SetId(device.Name)
	return nil
}

func resourceAviatrixDeviceRegistrationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	name := d.Get("name").(string)
	if name == "" {
		id := d.Id()
		log.Printf("[DEBUG] Looks like an import, no device name received. Import Id is %s", id)
		d.SetId(id)
		name = id
	}

	device := &goaviatrix.Device{
		Name: name,
	}

	device, err := client.GetDevice(device)
	if err == goaviatrix.ErrNotFound {
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not find device %s: %v", name, err)
	}

	d.Set("name", device.Name)
	d.Set("public_ip", device.PublicIP)
	d.Set("username", device.Username)
	d.Set("host_os", device.HostOS)
	d.Set("ssh_port", device.SshPort)
	d.Set("address_1", device.Address1)
	d.Set("address_2", device.Address2)
	d.Set("city", device.City)
	d.Set("state", device.State)
	d.Set("country", device.Country)
	d.Set("zip_code", device.ZipCode)
	d.Set("description", device.Description)
	d.Set("software_version", device.SoftwareVersion)
	d.Set("is_caag", device.IsCaag)

	d.SetId(device.Name)
	return nil
}

func resourceAviatrixDeviceRegistrationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	device := marshalDeviceRegistrationInput(d)

	if err := client.UpdateDevice(device); err != nil {
		return fmt.Errorf("could not update device registration information: %v", err)
	}

	if d.HasChange("software_version") {
		isCaag := d.Get("is_caag").(bool)
		if !isCaag {
			return fmt.Errorf("'software_version' can only be updated for managed cloudN (CaaG) devices")
		}
		softwareVersion := d.Get("software_version").(string)
		err := client.UpgradeGateway(&goaviatrix.Gateway{GwName: device.Name, SoftwareVersion: softwareVersion})
		if err != nil {
			return fmt.Errorf("could not upgrade CaaG: %v", err)
		}
	}

	d.SetId(device.Name)
	return nil
}

func resourceAviatrixDeviceRegistrationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*goaviatrix.Client)

	br := marshalDeviceRegistrationInput(d)

	if err := client.DeregisterDevice(br); err != nil {
		return fmt.Errorf("could not deregister device: %v", err)
	}

	d.SetId(br.Name)
	return nil
}
