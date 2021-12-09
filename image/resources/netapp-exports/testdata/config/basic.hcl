server "basic-attributes" {
    url      = "https://10.0.0.2:8080"
    user     = "nfs-proxy"
    password = file("./netapp-password")
}

server "tls" {
    url      = "https://10.0.0.2:8080"
    user     = "nfs-proxy"
    password = "secret"

    tls {
        ca_certificate    = file("./netapp-ca.pem")
        allow_common_name = true
    }
}

server "empty-tls" {
    url      = "https://10.0.0.2:8080"
    user     = "nfs-proxy"
    password = "secret"
    tls {}
}
