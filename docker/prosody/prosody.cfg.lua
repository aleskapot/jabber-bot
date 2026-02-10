-- Prosody configuration for development testing

-- Basic server settings
modules_enabled = {
    "roster";
    "saslauth";
    "tls";
    "dialback";
    "disco";
    "carbons";
    "pep";
    "private";
    "blocklist";
    "vcard";
    "version";
    "uptime";
    "time";
    "ping";
    "admin_adhoc";
    "offline";
    "c2s";
    "s2s";
}

-- Allow registration for development
allow_registration = true
registration_whitelist = { "^dev@", "^test@" }

-- SSL/TLS configuration
ssl = {
    key = "/etc/prosody/certs/localhost.key";
    certificate = "/etc/prosody/certs/localhost.crt";
}

-- Virtual hosts
VirtualHost "localhost"

-- Admin users
admins = { "admin@localhost" }

-- Logging
log = {
    info = "/var/log/prosody/prosody.log";
    error = "/var/log/prosody/prosody.err";
    "*syslog";
}

-- Storage (memory for development)
storage = "internal"