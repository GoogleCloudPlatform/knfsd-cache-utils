server "all-attributes" {
    url  = "https://10.0.0.2:8080"
    user = "nfs-proxy"

    password {
        google_cloud_secret {
            service_account_key = "Service Account Key"
            project             = "example"
            name                = "netapp-password"
            version             = "latest"
        }
    }
}

server "minimal" {
    url  = "https://10.0.0.2:8080"
    user = "nfs-proxy"

    password {
        google_cloud_secret {
            name = "netapp-password"
        }
    }
}

server "numeric-version" {
    url  = "https://10.0.0.2:8080"
    user = "nfs-proxy"

    password {
        google_cloud_secret {
            name    = "netapp-password"
            version = 42
        }
    }
}
