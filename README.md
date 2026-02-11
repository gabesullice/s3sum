# s3sum

Like `sha256sum`, but for S3. Computes CRC64NVME checksums so you can verify object integrity before or after upload.

## Install

```
go install github.com/gabesullice/s3sum@latest
```

Or build from source:

```
make
sudo make install
```

## Examples

Reads from stdin by default:

```
echo -n "hello" | s3sum
M3eFcAZSQlc=  -
```

Checksum one or more files with `-f`:

```
s3sum -f file1.txt -f file2.txt
M3eFcAZSQlc=  file1.txt
aBcDeFgHiJk=  file2.txt
```

Checksum a whole directory with `-d`, add `-r` to recurse:

```
s3sum -d ./data -r
M3eFcAZSQlc=  data/file1.txt
aBcDeFgHiJk=  data/sub/file2.txt
```

Save checksums and verify them later with `-c`:

```
s3sum -d ./data > checksums.txt
s3sum -c checksums.txt
data/file1.txt: OK
data/file2.txt: OK
```

Or pipe directly:

```
s3sum -d ./data | s3sum -c -
```

Output defaults to base64 (matching S3). Use `-e hex` for hex.
