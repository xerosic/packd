# 📦 packd - yet another archive file format



## features

- Seamless encryption and decryption of data
- Fast compression and decompression rates
- Easily implemented in any programming language
- Simple file structure


---

## specifications

### .pakd file structure

A .pakd file is essentially made up of *boxes* of compressed data, each containing their path and the box size (8 bytes).

The data is compressed using [zstd](https://github.com/facebook/zstd), providing fast compression and decompression rates within reasonable time.

#### PIE

The first 8 bytes are dedicated to the **P**ackd **I**dentification **E**ntry (aka magic bytes, file signature, etc...), which is `<!PAKD!>` (`3C 21 50 41 4B 44 21 3E` in bytes). 

A packd file is **NOT** a valid archive if those bytes are not present at the beginning of the file.

#### First Header

The first header is 8 bytes long and contains general information about the archive. For now only 1 byte is used, which is the *encrypted* flag.

- 1 byte for the encryption flag. If the flag is 0x00, the data is not encrypted. If the flag is 0x01, the data is encrypted.

The rest of the data is assumed to be empty and is ignored.

#### Header 

The header for each file in the archive (box) is 10 bytes long and has the following characteristics:

- 2 bytes for the path length
- 8 bytes for the box size

#### Box

The box itself contains the path of the file and the compressed data.

- *n* bytes for the path
- *x* bytes for the data

The data is placed after the path so it can be accessed from *[header_addr] + path_length*.

---
