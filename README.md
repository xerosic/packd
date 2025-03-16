# 📦 packd - an archive file format

## features

- Seamless encryption and decryption of data using asymmetric encryption
- Good compression and decompression rates
- Easily implemented in any programming language (this implementation is written in Golang)


## installation

The easiest way to install packd is by using `go install`:

```bash
go install github.com/xerosic/packd/cmd/pcdt@latest
```

You can also clone the repository and build the binary yourself:

```bash
git clone github.com/xerosic/packd.git
cd packd
go build ./cmd/pcdt
```


## specifications

### .pakd file structure

A .pakd file is essentially made up of *boxes* of compressed data, each containing their path and the box size (8 bytes).

The data is compressed using [zstd](https://github.com/facebook/zstd), providing good compression and decompression rates within reasonable time.

#### PIE

The first 8 bytes are dedicated to the **P**ackd **I**dentification **E**ntry (aka magic bytes, file signature, etc...), which is `<!PAKD!>` (`3C 21 50 41 4B 44 21 3E` in bytes). 

A packd file is **NOT** a valid archive if those bytes are not present at the beginning of the file.

---

#### First Header

The first header is 8 bytes long and contains general information about the archive. For now only 1 byte is used, which is the *encrypted* flag.

- 1 byte for the encryption flag. If the flag is 0x00, the data is not encrypted. If the flag is 0x01, the data is encrypted.

The rest of the data is assumed to be empty and is ignored.

---

#### Header 

The header for each file in the archive (box) is 10 bytes long and has the following characteristics:

- 2 bytes for the path length
- 8 bytes for the box size

---

#### Box

The box itself contains the path of the file and the compressed data.

- *n* bytes for the path
- *x* bytes for the data

The data is placed after the path so it can be accessed from *[header_addr] + path_length*.

---


`⚠️ The endianness is to be considered little-endian.`

### encryption

The encryption is done by using an hybrid method of the [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) and [RSA](https://en.wikipedia.org/wiki/RSA_(cryptosystem)) algorithms.

The AES key is generated randomly and is used to encrypt the data. The AES key is then encrypted using the RSA public key and stored in the archive (to bypass RSA message size limit).
The data can be then decrypted using the private key associated with the public key used to encrypt the file.

This method aims to provide a secure way to encrypt and decrypt data without the need to store the password (in case of symmetric encryption) in an edge server (which may not be considered secure).

