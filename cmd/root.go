package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/crc64nvme"
	"github.com/spf13/cobra"
)

var (
	filePaths []string
	dirPath   string
	recursive bool
	checkPath string
	encoding  string
)

var rootCmd = &cobra.Command{
	Use:          "s3sum",
	Short:        "Compute CRC64NVME checksums for S3 object integrity verification",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if encoding != "base64" && encoding != "hex" {
			return fmt.Errorf("unsupported encoding: %s (must be \"base64\" or \"hex\")", encoding)
		}
		if recursive && dirPath == "" {
			return fmt.Errorf("--recursive requires --directory")
		}
		if checkPath != "" && (len(filePaths) > 0 || dirPath != "") {
			return fmt.Errorf("--check is incompatible with --file and --directory")
		}

		if checkPath != "" {
			return runCheck(cmd)
		}
		return runSum(cmd)
	},
}

func computeChecksum(r io.Reader) ([]byte, error) {
	h := crc64nvme.New()
	if _, err := io.Copy(h, r); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func encodeChecksum(sum []byte) string {
	switch encoding {
	case "hex":
		return hex.EncodeToString(sum)
	default:
		return base64.StdEncoding.EncodeToString(sum)
	}
}

func runSum(cmd *cobra.Command) error {
	type input struct {
		r     io.Reader
		label string
	}

	seen := make(map[string]bool)
	var paths []string
	addPath := func(p string) {
		abs, err := filepath.Abs(p)
		if err != nil {
			abs = p
		}
		if !seen[abs] {
			seen[abs] = true
			paths = append(paths, p)
		}
	}
	for _, fp := range filePaths {
		addPath(fp)
	}
	if dirPath != "" {
		if recursive {
			err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					addPath(path)
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			entries, err := os.ReadDir(dirPath)
			if err != nil {
				return err
			}
			for _, e := range entries {
				if !e.IsDir() {
					addPath(filepath.Join(dirPath, e.Name()))
				}
			}
		}
	}

	var inputs []input
	if len(paths) == 0 {
		inputs = append(inputs, input{r: cmd.InOrStdin(), label: "-"})
	} else {
		for _, fp := range paths {
			f, err := os.Open(fp)
			if err != nil {
				return err
			}
			defer f.Close()
			inputs = append(inputs, input{r: f, label: fp})
		}
	}

	out := cmd.OutOrStdout()
	for _, in := range inputs {
		sum, err := computeChecksum(in.r)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s  %s\n", encodeChecksum(sum), in.label)
	}
	return nil
}

func runCheck(cmd *cobra.Command) error {
	var r io.Reader
	if checkPath == "-" {
		r = cmd.InOrStdin()
	} else {
		f, err := os.Open(checkPath)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	out := cmd.OutOrStdout()
	var failures int
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid checksum line: %q", line)
		}
		expectedChecksum, filePath := parts[0], parts[1]

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		sum, err := computeChecksum(f)
		f.Close()
		if err != nil {
			return err
		}

		actual := encodeChecksum(sum)
		if actual == expectedChecksum {
			fmt.Fprintf(out, "%s: OK\n", filePath)
		} else {
			fmt.Fprintf(out, "%s: FAILED\n", filePath)
			failures++
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if failures > 0 {
		return fmt.Errorf("%d checksum(s) did NOT match", failures)
	}
	return nil
}

func init() {
	rootCmd.Flags().StringSliceVarP(&filePaths, "file", "f", nil, "path to input file (repeatable; default: stdin)")
	rootCmd.Flags().StringVarP(&dirPath, "directory", "d", "", "path to directory (checksums all files)")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "recurse into subdirectories (requires --directory)")
	rootCmd.Flags().StringVarP(&checkPath, "check", "c", "", "verify checksums from file (or \"-\" for stdin)")
	rootCmd.Flags().StringVarP(&encoding, "encoding", "e", "base64", "output encoding: base64 or hex")

}

func Root() *cobra.Command {
	return rootCmd
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}
