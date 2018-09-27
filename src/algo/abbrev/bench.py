#!/usr/bin/env python3

import sys
import subprocess
import os

def main():
    pattern = sys.argv[1]
    input_path = sys.argv[2]
    fzf_args = sys.argv[3:]

    with open(input_path) as f:
        input_text = f.read()
    input_bytes = input_text.encode()

    fzf_root = sys.argv[0]
    for i in range(4):
        fzf_root = os.path.dirname(fzf_root)

    fzf_bin = os.path.abspath(os.path.join(fzf_root, 'fzf-abbrev'))

    cmd = [fzf_bin, '--filter', pattern] + fzf_args
    print(cmd)
    proc = subprocess.Popen(
            cmd,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE)
    out, err = proc.communicate(input=input_bytes)

    print('Input: {} lines\nOuput: {} lines'.format(
            input_text.count('\n'), out.decode().count('\n')))
    # print(out.decode().split('\n')[0:10].join('\n'))
    print(err.decode())
    print(proc.returncode)


if __name__ == "__main__":
    main()