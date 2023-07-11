resource "cdcovhns_name_servers" "capybaradevcloud" {
  service_name = "capybaradev.cloud"

  name_servers = {
    "ns1" = {
      host = "noel.ns.cloudflare.com"
    },
    "ns2" = {
      host = "june.ns.cloudflare.com"
    },
  }
}
