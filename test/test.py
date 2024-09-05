import subprocess
from os import path

TCGEN_BIN="../bin/tcgen"
TEST_YAML_PATH="./test.yaml"
TESTMOD_DIR="./testmod"

p = subprocess.Popen(f"{TCGEN_BIN} -path {TEST_YAML_PATH}", shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

with open(path.join(TESTMOD_DIR, "fn", "fn_gen.go"), "w+") as f:
    for line in p.stdout.readlines():
        f.write(line.decode())
retval = p.wait()


subprocess.run(["go", "run", "main.go"], cwd=TESTMOD_DIR)
