# s3sum

Compute CRC64NVME checksums for verifying S3 object integrity. Output format matches `sha256sum` conventions.

## Install

```
go install github.com/gabesullice/s3sum@latest
```

Or build from source:

```
make
sudo make install
```

## Usage

```
s3sum [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Path to input file (repeatable) |
| `--directory` | `-d` | Path to directory (checksums all files) |
| `--recursive` | `-r` | Recurse into subdirectories (requires `--directory`) |
| `--check` | `-c` | Verify checksums from file (or `-` for stdin) |
| `--encoding` | `-e` | Output encoding: `base64` (default) or `hex` |

## Examples

Read from stdin:

```
echo -n "hello" | s3sum
M3eFcAZSQlc=  -
```

Checksum a single file:

```
s3sum -f myfile.txt
M3eFcAZSQlc=  myfile.txt
```

Checksum multiple files:

```
s3sum -f file1.txt -f file2.txt
M3eFcAZSQlc=  file1.txt
aBcDeFgHiJk=  file2.txt
```

Checksum all files in a directory:

```
s3sum -d ./data
M3eFcAZSQlc=  data/file1.txt
aBcDeFgHiJk=  data/file2.txt
```

Recursive directory checksum:

```
s3sum -d ./data -r
M3eFcAZSQlc=  data/file1.txt
aBcDeFgHiJk=  data/sub/file2.txt
```

Hex-encoded output:

```
s3sum -e hex -f myfile.txt
3377857006524257  myfile.txt
```

Save and verify checksums:

```
s3sum -d ./data > checksums.txt
s3sum -c checksums.txt
data/file1.txt: OK
data/file2.txt: OK
```

Verify checksums from stdin:

```
s3sum -d ./data | s3sum -c -
data/file1.txt: OK
data/file2.txt: OK
```
