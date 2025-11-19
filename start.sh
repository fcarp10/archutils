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

curl -fsL "https://github.com/fcarp10/archutils/releases/latest/download/archutils" -o "$temp_file"
check $? "Downloading archutils"

chmod +x "$temp_file"
check $? "Making archutils executable"

"$temp_file" "$@"
check $? "Executing archutils"

rm -f "$temp_file"
check $? "Deleting the temporary file"
} 
