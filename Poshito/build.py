import os
import sys
import string
import secrets
import argparse
import subprocess

def generate_random_string(length):
    characters = string.ascii_letters + string.digits + string.punctuation
    random_string = ''.join(secrets.choice(characters) for _ in range(length))
    return random_string


# compiler args
compile_cmd = "cd Agent && GOOS=windows GOARCH=amd64 {} build"

def main():
    global compile_cmd
    parser = argparse.ArgumentParser(prog="build", description="Poshito-C2 agent builder")
    parser.add_argument("format", choices=["exe", "dll"])
    parser.add_argument("-nx", "--no-upx", action="store_true", help="don't UPX")
    parser.add_argument("-ns", "--no-upx-sec-obf", action="store_true", help="don't obfuscate UPX section names")
    parser.add_argument("-ng", "--no-garble", action="store_true", help="don't use Garble (use standard Go compiler)")
    parser.add_argument("-en", "--export-name", metavar="<name>", type=str, help="dll export name")

    args = parser.parse_args()

    if args.no_garble:
        compile_cmd = compile_cmd.format("go")
    else:
        compile_cmd = compile_cmd.format("garble")

    ret = subprocess.run(compile_cmd, capture_output=True, shell=True)
    if ret.returncode != 0:
        print("[-] ERROR: Could not compile")
        return
    else:
        print(f"[+] Compiled successfully")


if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        quit()