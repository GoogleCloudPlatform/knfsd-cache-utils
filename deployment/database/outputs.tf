output "name" {
  value = google_sql_database_instance.fsids.name
}

output "sql_user" {
  value = google_sql_user.knfsd.name
}

output "database_name" {
  value = google_sql_database.fsids.name
}

output "connection_name" {
  value = google_sql_database_instance.fsids.connection_name
}

output "public_ip_address" {
  value = google_sql_database_instance.fsids.public_ip_address
}

output "private_ip_address" {
  value = google_sql_database_instance.fsids.private_ip_address
}

output "ip_address" {
  value = google_sql_database_instance.fsids.ip_address
}

output "server_ca_cert" {
  value = google_sql_database_instance.fsids.server_ca_cert
}
