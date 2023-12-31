---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cdcovhns_name_servers Resource - cdcovhns"
subcategory: ""
description: |-
  OVH Name Server resource
---

# cdcovhns_name_servers (Resource)

OVH Name Server resource

## Example Usage

```terraform
resource "cdcovhns_name_servers" "examplecom" {
  service_name = "example.com"

  name_servers = {
    "ns1" = {
      host = "noel.ns.example.com"
    },
    "ns2" = {
      host = "june.ns.example.com"
    },
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name_servers` (Attributes Map) (see [below for nested schema](#nestedatt--name_servers))
- `service_name` (String) Domain name

### Read-Only

- `type` (String) OVH Name Servers type - 'external' (if external name servers like Cloudflare) or 'hosted' if OVH Name Servers.

<a id="nestedatt--name_servers"></a>
### Nested Schema for `name_servers`

Required:

- `host` (String) DNS Hostname

Optional:

- `ip` (String) DNS IP address

Read-Only:

- `id` (Number)
- `is_used` (Boolean)
- `to_delete` (Boolean)

## Import

Import is supported using the following syntax:

```shell
terraform import cdcovhns_name_servers.examplecom example.com
```
