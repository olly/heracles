# Heracles

## Usage

    heracles init

    heracles generate-server-ca

    heracles generate-client-ca

    heracles unlock-server-ca
      // Removes the password from `server/ca.key`.
      //   - ensures path is in .gitignore
      //   - touches file at path, chowns to root:root, chmods to 400
      //   - writes unlocked key to `server/ca.unlocked.key`

    heracles add-host logs.example.com

    heracles remove-host logs.example.com

    heracles add-user mortal@example.com
      // Generates a certificate / key pair for a user.
      //   - ensures `client/users/mortal@example.com.key` is in .gitignore
      //   - generates client key
      //   - generates temporary CSR
      //   - signs temporary CSR with client certificate `client/users/mortal@example.com.crt`
      //   - updates `client/trusted-users` with client certificate
      //   - signs `client/trusted-users` with signing key
      //   - commits to repository "Added 'mortal@example.com' client certificate."

    heracles remove-user mortal@example.com

## Data

    config
    client/
      ca.key
      ca.crt
      trusted-users
      trusted-users.sig
      users/
        mortal@example.com.crt
    server/
      ca.key
      ca.crt
      hosts/
        logs.example.com.crt

## Config

    [client]
    ca_key_size = 4096
    user_key_size = 2048
    signing_key_id = ABCDEF
    
    [server]
    ca_key_size = 4096
    host_key_size = 2048

## Commands

    # Create the CA Key and Certificate for signing Client Certs
    openssl genrsa -des3 -out ca.key 4096
    openssl req -new -x509 -days 365 -key ca.key -out ca.crt
    
    # Create the Server Key, CSR, and Certificate
    openssl genrsa -des3 -out server.key 1024
    openssl req -new -key server.key -out server.csr
    
    # We're self signing our own server cert here.  This is a no-no in production.
    openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt
    
    # Create the Client Key and CSR
    openssl genrsa -des3 -out client.key 1024
    openssl req -new -key client.key -out client.csr
    
    # Sign the client certificate with our CA cert.  Unlike signing our own server cert, this is what we want to do.
    openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out client.crt

