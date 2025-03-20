# SSH MFA (TOTP) Wrapper Utility

This utility is a Go application that automates SSH login for servers protected with multi-factor authentication (MFA). It securely provides both the SSH password and the TOTP verification code using the `totp-cli` tool.

## Prerequisites

### Install Dependencies
The program uses the following Go packages:
```sh
go mod tidy
```

## Prerequisites

### Install `totp-cli`
The program uses [`totp-cli`](https://github.com/yitsushi/totp-cli) to generate TOTP codes. Install it following the instructions for your environment, or go to the releases page and download the appropriate version for your environment.

Ensure that your SSH servers are already configured with `totp-cli` for generating codes.

## Build

```sh
go build -o ssh-mfa
```

## Usage
Run the program using the following command:
```sh
./mfa-ssh <server> <namespace> [username]
```
Or alternatively:
```sh
./mfa-ssh <username@server> <namespace>
```

Where:
- `<server>` is the remote SSH server (e.g., `example.com`).
- `<username@server>` is the remote SSH server with username (e.g., `user2@example.com`).
- `<namespace>` is the namespace used in `totp-cli` to generate the correct TOTP code.
- `[username]` (optional) is the username for SSH login if not using the combined format.

### TOTP Profile Setup
When using the user@server format, make sure to create your TOTP profile with the same identifier:
```sh
totp-cli new <namespace> user@server
```

## Environment Variables (Optional)
The script can use an environment variable `TOTP_PASS` to store the password required for `totp-cli` to decrypt the stored credentials.

### Setting the Environment Variable
You can export the password before running the program to avoid being prompted:
```sh
export TOTP_PASS=<your_totp_password>
./ssh-mfa <server> <namespace> [username]
```
Or using the combined format:
```sh
export TOTP_PASS=<your_totp_password>
./ssh-mfa <username@server> <namespace>
```
Alternatively, if `TOTP_PASS` is not set, the program will prompt for it during execution.

## How It Works
1. The program checks if the `TOTP_PASS` environment variable is set:
   - If set, it uses it directly.
   - If not set, it prompts the user to enter the password and stores it for subsequent use.
2. The program logs in to the SSH server:
   - It first prompts for and enters the SSH password.
   - It then generates a TOTP code using `totp-cli` and enters it when prompted.
3. The program uses a pseudo-terminal (PTY) to interact with the SSH process and automatically handle password and verification code prompts.
4. Once authenticated, the SSH session is handed over to the user.

## Notes
- Ensure that `totp-cli` is installed and configured correctly with your accounts before using this program.
- The program only automates login and does not store any credentials persistently.
- When using different usernames for SSH, be sure to set up your TOTP profiles using the full `user@server` format.
- The program uses Go's PTY package to handle interactive terminal sessions securely.
