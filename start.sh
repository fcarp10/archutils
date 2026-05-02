#!/bin/sh -e

{
rc='\033[0m'
red='\033[0;31m'

check() {
    exit_code=$1
    message=$2

    if [ "$exit_code" -ne 0 ]; then
        printf '%sERROR: %s%s\n' "$red" "$message" "$rc"
        exit 1
    fi

    unset exit_code
    unset message
}

temp_file=$(mktemp)
check $? "Creating the temporary file"

checksum_file=$(mktemp)
check $? "Creating the checksum temporary file"

trap 'rm -f "$temp_file" "$checksum_file"' EXIT INT TERM

release_url="https://github.com/fcarp10/archutils/releases/latest/download"

curl -fsL "$release_url/archutils" -o "$temp_file"
check $? "Downloading archutils"

curl -fsL "$release_url/archutils.sha256" -o "$checksum_file"
check $? "Downloading checksum"

expected=$(cut -d' ' -f1 "$checksum_file")
actual=$(sha256sum "$temp_file" | cut -d' ' -f1)
if [ "$expected" != "$actual" ]; then
    printf '%sERROR: Checksum verification failed%s\n' "$red" "$rc"
    printf '  Expected: %s\n' "$expected"
    printf '  Got:      %s\n' "$actual"
    exit 1
fi

chmod +x "$temp_file"
check $? "Making archutils executable"

"$temp_file" "$@"
check $? "Executing archutils"

rm -f "$temp_file" "$checksum_file"
check $? "Deleting temporary files"
} 
