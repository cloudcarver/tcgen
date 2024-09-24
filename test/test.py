import subprocess
from os import path

TCGEN_BIN="../bin/tcgen"
TESTMOD_DIR="./testmod"

p = subprocess.Popen(f"{TCGEN_BIN} -config tcgen.yaml", shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

subprocess.run(["go", "run", "main.go"], cwd=TESTMOD_DIR)
