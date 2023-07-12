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
