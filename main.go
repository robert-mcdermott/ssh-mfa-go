package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// getTOTPCode generates a TOTP code using the external "totp-cli" command.
func getTOTPCode(namespace, servername string) (string, error) {
	totpPassword := os.Getenv("TOTP_PASS")

	if totpPassword == "" {
		fmt.Print("TOTP Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // New line after password input
		if err != nil {
			return "", fmt.Errorf("failed to read TOTP password: %v", err)
		}
		totpPassword = string(bytePassword)
		os.Setenv("TOTP_PASS", totpPassword) // Store in environment for reuse
	}

	cmd := exec.Command("totp-cli", "generate", namespace, servername)
	cmd.Stdin = strings.NewReader(totpPassword)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate TOTP code: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	return lines[len(lines)-1], nil // Return last line (TOTP code)
}

// sshWithTOTP handles SSH login and two-factor authentication using TOTP.
func sshWithTOTP(server, namespace string) error {
	fmt.Print("SSH Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	if err != nil {
		return fmt.Errorf("failed to read SSH password: %v", err)
	}
	password := string(bytePassword)

	totpCode, err := getTOTPCode(namespace, server)
	if err != nil {
		return err
	}

	cmd := exec.Command("ssh", server)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start SSH session: %v", err)
	}
	defer ptmx.Close()

	// Save the current state of the terminal
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Handle password and TOTP input in a separate goroutine
	go func() {
		// Buffer to read output
		buf := make([]byte, 1024)
		var output string
		passwordSent := false
		totpSent := false

		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				return
			}
			
			output += string(buf[:n])
			os.Stdout.Write(buf[:n])
			
			// Check for password prompt
			if strings.Contains(output, "Password:") && !passwordSent {
				ptmx.Write([]byte(password + "\n"))
				passwordSent = true
				output = "" // Reset output buffer
			}
			
			// Check for verification code prompt
			if strings.Contains(output, "Verification code:") && !totpSent {
				ptmx.Write([]byte(totpCode + "\n"))
				totpSent = true
				output = "" // Reset output buffer
			}
		}
	}()

	// Copy stdin to the pty
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			ptmx.Write(buf[:n])
		}
	}()

	// Wait for the command to finish
	return cmd.Wait()
}

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		fmt.Println("Usage: go run main.go <server> <namespace> [username]")
		fmt.Println("  or:  go run main.go <username@server> <namespace>")
		os.Exit(1)
	}

	server := os.Args[1]
	namespace := os.Args[2]
	
	// Check if username is provided as a separate argument
	if len(os.Args) == 4 {
		username := os.Args[3]
		// Prepend username to server if it doesn't already have one
		if !strings.Contains(server, "@") {
			server = username + "@" + server
		}
	}

	if err := sshWithTOTP(server, namespace); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
