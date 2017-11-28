#!/bin/sh
#
#  OpenSSL コマンドを使って証明書を作成します。
#
DAYS=3650
SUBJECT="/C=JP/ST=Tokyo/L=Chiyoda-ku/O=Hirakawa-cho/OU=eLV/CN=tavle.example.com"

# CA
openssl genrsa -out ca-privatekey.pem 2048
openssl req -new -key ca-privatekey.pem -out ca-csr.pem -subj $SUBJECT
openssl req -x509 -key ca-privatekey.pem -in ca-csr.pem -out ca-crt.pem -days $DAYS

# Server
openssl genrsa -out server-privatekey.pem
openssl req -new -key server-privatekey.pem -out server-csr.pem -subj $SUBJECT
openssl x509 -req -CA ca-crt.pem -CAkey ca-privatekey.pem -CAcreateserial -in server-csr.pem -out server-crt.pem -days $DAYS

cp server-privatekey.pem key.pem
cp server-crt.pem cert.pem
